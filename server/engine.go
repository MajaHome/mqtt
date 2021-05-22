package server

import (
	"crypto/rand"
	"gorm.io/driver/sqlite"
	_ "gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"mqtt/model"
	"mqtt/packet"
	"net"
)

type Engine struct {
	db       *gorm.DB
	clients  map[string]*Client
	channel  chan *Event
	send     map[uint16]Event
	delivery map[uint16]Event
	retain   map[string]Event
}

func NewEngine() *Engine {
	log.Println("initialize database")
	conn, err := gorm.Open(sqlite.Open("mqtt.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	log.Println("migrate database")
	conn.AutoMigrate(&model.User{})
	conn.AutoMigrate(&model.Subscription{})
	conn.AutoMigrate(&model.RetainMessage{})

	e := &Engine{
		db:       conn,
		clients:  make(map[string]*Client),
		channel:  make(chan *Event),
		send:     make(map[uint16]Event),
		delivery: make(map[uint16]Event),
		retain:   make(map[string]Event),
	}

	go e.manageClients()

	return e
}

func (e *Engine) Process(server *Server) {
	for {
		conn, err := server.Accept()
		if err != nil {
			log.Println("err accept", err.Error())
			conn.Close()
			continue
		}

		// process with CONNECT
		go func() {
			pkt, err := server.ReadPacket(conn)
			if err != nil {
				log.Println("err read packet", err.Error())
				conn.Close()
				return
			}

			if pkt.Type() == packet.CONNECT {
				res := packet.NewConnAck()

				connPacket := pkt.(*packet.ConnPacket)

				// check version, now only 4 (3.11)
				if connPacket.Version != 4 {
					res.ReturnCode = uint8(packet.ConnectUnacceptableProtocol)
					return
				}

				if len(connPacket.ClientID) == 0 && !connPacket.CleanSession {
					res.ReturnCode = uint8(packet.ConnectIndentifierRejected)
					return
				}

				res.ReturnCode = uint8(packet.ConnectAccepted)

				// check authorization
				if len(connPacket.Username) > 0 {
					var user model.User
					result := e.db.Take(&user, "user_name = ? and password = ?", connPacket.Username, connPacket.Password)
					if result.Error != nil {
						log.Println("username/password is wrong")
						res.ReturnCode = uint8(packet.ConnectBadUserPass)
					}
				}

				var client *Client
				if res.ReturnCode == uint8(packet.ConnectAccepted) {
					res.Session = false
					if !connPacket.CleanSession && e.clients[connPacket.ClientID] != nil {
						res.Session = !connPacket.CleanSession
					}

					client = e.saveClient(connPacket.ClientID, conn, !connPacket.CleanSession)
					if connPacket.Will != nil && res.Session {
						client.will = connPacket.Will
					}
				}

				err = server.WritePacket(conn, res)
				if err != nil {
					log.Println("err write packet", err.Error())
					conn.Close()
					return
				}

				// close connection if not authorized
				if res.ReturnCode != uint8(packet.ConnectAccepted) {
					log.Println("err connection declined")
					conn.Close()
					return
				}

				// TODO Restore subscription

				// start manage client
				go client.Start(server)
			} else {
				log.Println("wrong packet. expect CONNECT")
				conn.Close()
				return
			}
		}()
	}
}

func (e *Engine) manageClients() {
	for event := range e.channel {
		log.Println("engineChan receive message: " + event.String())

		switch event.packetType {
		case packet.DISCONNECT:
			client := e.clients[event.clientId]
			client.Stop()
			if !client.session {
				delete(e.clients, event.clientId)
			}
		case packet.SUBSCRIBE:
			// todo: do not support wildcard topic names on subscribe yet
			var retain model.RetainMessage
			if row, err := e.db.Where(&model.RetainMessage{Topic: event.topic.name}).Find(&retain).Rows(); err == nil {
				for row.Next() {
					revent := &Event{messageId: retain.MessageId, topic: EventTopic{name: retain.Topic, qos: retain.Qos},
						payload: retain.Payload, qos: retain.Qos, retain: true}
					e.publishMessage(revent)
				}
			}

			// if not clean session - save subscription
			if e.clients[event.clientId].session {
				subs := model.Subscription{ClientId: event.clientId, Topic: event.topic.name, Qos: event.topic.qos}
				e.db.Create(&subs)
			}
		case packet.UNSUBSCRIBE:
			client := e.clients[event.clientId]

			if client.session {
				e.db.Where(&model.Subscription{ClientId: event.clientId, Topic: event.topic.name}).Delete(&model.Subscription{})
			}
		case packet.PUBLISH: // in
			client := e.clients[event.clientId]
			if event.retain {
				retain := model.RetainMessage{MessageId: event.messageId, Topic: event.topic.name,
					Payload: event.payload, Qos: event.topic.qos}
				e.db.Create(&retain)
			}

			// todo if topic has higher qos - increase qos in published message. ps. not sure. can send message to
			// unsubscribed topic

			switch packet.QoS(event.qos) {
			case packet.AtMostOnce:
				e.publishMessage(event)
			case packet.AtLeastOnce: // PUBLISH -> PUBACK
				res := &Event{packetType: packet.PUBACK, messageId: event.messageId, clientId: event.clientId}
				client.clientChan <- res

				// save message
				e.delivery[event.messageId] = *res

				// send published message to all subscribed clients
				e.publishMessage(event)
			case packet.ExactlyOnce: // PUBLISH ->PUBREC, PUBREL - PUBCOMP
				res := &Event{packetType: packet.PUBREC, messageId: event.messageId, clientId: event.clientId}
				client.clientChan <- res

				// record packet in delivery queue
				e.delivery[event.messageId] = *event
			}
		case packet.PUBACK: // out
			// we receive answer on publish command,
			_, ok := e.send[event.messageId]
			if ok {
				// remove from sent queue
				delete(e.send, event.messageId)
			} else {
				log.Printf("error process PUBACK, message %d for %s was not found\n", event.messageId, event.clientId)
			}
		case packet.PUBREC: // out
			// we receive answer on PUBLISH with qos2,
			answer, ok := e.send[event.messageId]
			if ok && event.clientId == answer.clientId {
				// send PUBREL
				res := &Event{packetType: packet.PUBREL, clientId: event.clientId, messageId: event.messageId}
				e.clients[event.clientId].clientChan <- res
			} else {
				log.Printf("error process PUBREC, message %d for %s was not found\n", event.messageId, event.clientId)
			}
		case packet.PUBREL: // in
			event, ok := e.delivery[event.messageId]
			if ok {
				// send answer PUBCOMP
				res := &Event{packetType: packet.PUBCOMP, messageId: event.messageId, clientId: event.clientId}
				e.clients[event.clientId].clientChan <- res

				// send publish
				e.publishMessage(&event)

				// remove from delivery queue
				delete(e.delivery, event.messageId)
			} else {
				log.Printf("error process PUBREL, message %d for %s was not found\n", event.messageId, event.clientId)
			}
		case packet.PUBCOMP: // out
			// we receive answer on PUBREL
			_, ok := e.send[event.messageId]
			if ok {
				// remove from sent queue with qos2
				delete(e.send, event.messageId)
			} else {
				log.Printf("error process PUBCOMP, message %d for %s was not found\n", event.messageId, event.clientId)
			}
		default:
			log.Println("engineChan: unexpected disconnect")

			client := e.clients[event.clientId]
			if client != nil {
				// if will message is set for client - send this message to all clients
				if client.will != nil {
					will := &Event{
						topic: EventTopic{name: client.will.Topic},
						payload: client.will.Payload,
						qos: client.will.QoS.Int(),
						retain: client.will.Retain,
					}
					for _, c := range e.clients {
						if c != nil {
							client.clientChan <- will
						}
					}
				}

				// delete clean session
				if !client.session {
					delete(e.clients, event.clientId)
				}
			}
		}
	}
}

func (e *Engine) publishMessage(event *Event) {
	event.packetType = packet.PUBLISH

	for _, client := range e.clients {
		if client != nil {
			if client.Contains(event.topic.name) {
				client.clientChan <- event

				if event.qos > 0 {
					e.send[event.messageId] = *event
				}
			}
		}
	}
}

func (e *Engine) saveClient(id string, conn net.Conn, session bool) *Client {
	var cid string = id
	if id == "" {
		// generate temporary ident
		ident := make([]byte, 8)
		rand.Read(ident)
		cid = string(cid)
	}

	if e.clients[cid] != nil {
		e.clients[cid].conn = conn
	} else {
		client := NewClient(cid, conn, session, e.channel)
		e.clients[cid] = client
	}

	return e.clients[cid]
}
