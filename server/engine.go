package server

import (
	"crypto/rand"
	"github.com/MajaSuite/mqtt/model"
	"github.com/MajaSuite/mqtt/packet"
	"github.com/MajaSuite/mqtt/transport"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"net"
)

type Engine struct {
	db       *gorm.DB
	clients  map[string]*Client
	channel  chan *transport.Event
	send     map[uint16]transport.Event
	delivery map[uint16]transport.Event
	retain   map[string]transport.Event
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
		channel:  make(chan *transport.Event),
		send:     make(map[uint16]transport.Event),
		delivery: make(map[uint16]transport.Event),
		retain:   make(map[string]transport.Event),
	}

	go e.manageClients()

	return e
}

func (e *Engine) Process(server *transport.Server) {
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

		switch event.PacketType {
		case packet.DISCONNECT:
			client := e.clients[event.ClientId]
			client.Stop()
			if !client.session {
				delete(e.clients, event.ClientId)
			}
		case packet.SUBSCRIBE:
			// todo: do not support wildcard topic names on subscribe yet
			var retain model.RetainMessage
			if row, err := e.db.Where(&model.RetainMessage{Topic: event.Topic.Name}).Find(&retain).Rows(); err == nil {
				for row.Next() {
					revent := &transport.Event{MessageId: retain.MessageId, Topic: transport.EventTopic{Name: retain.Topic, Qos: retain.Qos},
						Payload: retain.Payload, Qos: retain.Qos, Retain: true}
					e.publishMessage(revent)
				}
			}

			// if not clean session - save subscription
			if e.clients[event.ClientId].session {
				subs := model.Subscription{ClientId: event.ClientId, Topic: event.Topic.Name, Qos: event.Topic.Qos}
				e.db.Create(&subs)
			}
		case packet.UNSUBSCRIBE:
			client := e.clients[event.ClientId]

			if client.session {
				e.db.Where(&model.Subscription{ClientId: event.ClientId, Topic: event.Topic.Name}).Delete(&model.Subscription{})
			}
		case packet.PUBLISH: // in
			client := e.clients[event.ClientId]
			if event.Retain {
				retain := model.RetainMessage{MessageId: event.MessageId, Topic: event.Topic.Name,
					Payload: event.Payload, Qos: event.Topic.Qos}
				e.db.Create(&retain)
			}

			// todo if topic has higher qos - increase qos in published message. ps. not sure. can send message to
			// unsubscribed topic

			switch packet.QoS(event.Qos) {
			case packet.AtMostOnce:
				e.publishMessage(event)
			case packet.AtLeastOnce: // PUBLISH -> PUBACK
				res := &transport.Event{PacketType: packet.PUBACK, MessageId: event.MessageId, ClientId: event.ClientId}
				client.clientChan <- res

				// save message
				e.delivery[event.MessageId] = *res

				// send published message to all subscribed clients
				e.publishMessage(event)
			case packet.ExactlyOnce: // PUBLISH ->PUBREC, PUBREL - PUBCOMP
				res := &transport.Event{PacketType: packet.PUBREC, MessageId: event.MessageId, ClientId: event.ClientId}
				client.clientChan <- res

				// record packet in delivery queue
				e.delivery[event.MessageId] = *event
			}
		case packet.PUBACK: // out
			// we receive answer on publish command,
			_, ok := e.send[event.MessageId]
			if ok {
				// remove from sent queue
				delete(e.send, event.MessageId)
			} else {
				log.Printf("error process PUBACK, message %d for %s was not found\n", event.MessageId, event.ClientId)
			}
		case packet.PUBREC: // out
			// we receive answer on PUBLISH with qos2,
			answer, ok := e.send[event.MessageId]
			if ok && event.ClientId == answer.ClientId {
				// send PUBREL
				res := &transport.Event{PacketType: packet.PUBREL, ClientId: event.ClientId, MessageId: event.MessageId}
				e.clients[event.ClientId].clientChan <- res
			} else {
				log.Printf("error process PUBREC, message %d for %s was not found\n", event.MessageId, event.ClientId)
			}
		case packet.PUBREL: // in
			event, ok := e.delivery[event.MessageId]
			if ok {
				// send answer PUBCOMP
				res := &transport.Event{PacketType: packet.PUBCOMP, MessageId: event.MessageId, ClientId: event.ClientId}
				e.clients[event.ClientId].clientChan <- res

				// send publish
				e.publishMessage(&event)

				// remove from delivery queue
				delete(e.delivery, event.MessageId)
			} else {
				log.Printf("error process PUBREL, message %d for %s was not found\n", event.MessageId, event.ClientId)
			}
		case packet.PUBCOMP: // out
			// we receive answer on PUBREL
			_, ok := e.send[event.MessageId]
			if ok {
				// remove from sent queue with qos2
				delete(e.send, event.MessageId)
			} else {
				log.Printf("error process PUBCOMP, message %d for %s was not found\n", event.MessageId, event.ClientId)
			}
		default:
			log.Println("engineChan: unexpected disconnect")

			client := e.clients[event.ClientId]
			if client != nil {
				// if will message is set for client - send this message to all clients
				if client.will != nil {
					will := &transport.Event{
						Topic:   transport.EventTopic{Name: client.will.Topic},
						Payload: client.will.Payload,
						Qos:     client.will.QoS.Int(),
						Retain:  client.will.Retain,
					}
					for _, c := range e.clients {
						if c != nil {
							client.clientChan <- will
						}
					}
				}

				// delete clean session
				if !client.session {
					delete(e.clients, event.ClientId)
				}
			}
		}
	}
}

func (e *Engine) publishMessage(event *transport.Event) {
	event.PacketType = packet.PUBLISH

	for _, client := range e.clients {
		if client != nil {
			if client.Contains(event.Topic.Name) {
				client.clientChan <- event

				if event.Qos > 0 {
					e.send[event.MessageId] = *event
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
