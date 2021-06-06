package broker

import (
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
	channel  chan transport.Event
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
		channel:  make(chan transport.Event),
		send:     make(map[uint16]transport.Event),
		delivery: make(map[uint16]transport.Event),
		retain:   make(map[string]transport.Event),
	}

	go e.clientThread()
	return e
}

func (e *Engine) ManageServer(server *Server) {
	for {
		conn, err := server.Accept()
		if err != nil {
			log.Println("err accept", err.Error())
			conn.Close()
			continue
		}

		// process with CONNECT
		go func() {
			pkt, err := packet.ReadPacket(conn)
			if err != nil {
				log.Println("err read packet", err.Error())
				conn.Close()
				return
			}

			if pkt.Type() == packet.CONNECT {
				res := packet.NewConnAck()

				connPacket := pkt.(*packet.ConnPacket)
				var id string = connPacket.ClientID

				// check version, now only 4 (3.11)
				if connPacket.Version != 4 {
					res.ReturnCode = uint8(packet.ConnectUnacceptableProtocol)
					return
				}

				if len(id) == 0 && !connPacket.CleanSession {
					res.ReturnCode = uint8(packet.ConnectIndentifierRejected)
					return
				}

				res.ReturnCode = uint8(packet.ConnectAccepted)

				// check authorization
				if len(connPacket.Username) > 0 {
					var user model.User
					result := e.db.Take(&user, "user_name = ? and password = ?", connPacket.Username,
						connPacket.Password)

					if result.Error != nil {
						log.Println("username/password is wrong")
						res.ReturnCode = uint8(packet.ConnectBadUserPass)
					} else {
						if id == "" {
							id = connPacket.Username
						}
					}
				}

				var client *Client
				if res.ReturnCode == uint8(packet.ConnectAccepted) {
					res.Session = !connPacket.CleanSession
					client = e.saveClient(id, conn, res.Session)
					if connPacket.Will != nil && res.Session {
						client.will = connPacket.Will
					}
				}

				err = packet.WritePacket(conn, res)
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

				if res.Session {
					// TODO Restore subscription
				}

				// start manage client
				go client.Start()
			} else {
				log.Println("wrong packet. expect CONNECT")
				conn.Close()
				return
			}
		}()
	}
}

func (e *Engine) clientThread() {
	for event := range e.channel {
		log.Println("engine receive message: " + event.String())

		switch event.PacketType {
		case packet.DISCONNECT:
			client := e.clients[event.ClientId]
			client.Stop()
			if !client.session {
				delete(e.clients, event.ClientId)
			}
		case packet.SUBSCRIBE:
			var retains []model.RetainMessage
			if rows, err := e.db.Find(&retains).Rows(); err == nil {
				for rows.Next() {
					var msg model.RetainMessage
					e.db.ScanRows(rows, &msg)
					if packet.MatchTopic(event.Topic.Name, msg.Topic) {
						retain := transport.Event{
							ClientId: event.ClientId,
							MessageId: msg.MessageId,
							Topic: transport.EventTopic{Name: msg.Topic, Qos: msg.Qos},
							Payload: msg.Payload,
							Qos: msg.Qos,
							Retain: true,
						}
						e.publishMessage(retain)
					}
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
				if event.Payload != "" {
					// save retain message
					retain := model.RetainMessage{MessageId: event.MessageId, Topic: event.Topic.Name,
						Payload: event.Payload, Qos: event.Topic.Qos}
					e.db.Create(&retain)
				} else {
					// remove retain message
					e.db.Where("topic = ? and qos = ?", event.Topic.Name, event.Qos).Delete(&model.RetainMessage{})
					continue
				}
			}

			switch packet.QoS(event.Qos) {
			case packet.AtMostOnce:
				e.publishMessage(event)
			case packet.AtLeastOnce: // PUBLISH -> PUBACK
				res := transport.Event{PacketType: packet.PUBACK, MessageId: event.MessageId, ClientId: event.ClientId}
				client.clientChan <- res

				// save message
				e.delivery[event.MessageId] = res

				// send published message to all subscribed clients
				e.publishMessage(event)
			case packet.ExactlyOnce: // PUBLISH ->PUBREC, PUBREL - PUBCOMP
				client.clientChan <- transport.Event{PacketType: packet.PUBREC, MessageId: event.MessageId,
					ClientId: event.ClientId}

				// record packet in delivery queue
				e.delivery[event.MessageId] = event
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
				e.clients[event.ClientId].clientChan <- transport.Event{PacketType: packet.PUBREL,
					ClientId: event.ClientId, MessageId: event.MessageId}
			} else {
				log.Printf("error process PUBREC, message %d for %s was not found\n", event.MessageId, event.ClientId)
			}
		case packet.PUBREL: // in
			event, ok := e.delivery[event.MessageId]
			if ok {
				// send answer PUBCOMP
				e.clients[event.ClientId].clientChan <- transport.Event{PacketType: packet.PUBCOMP,
					MessageId: event.MessageId, ClientId: event.ClientId}

				// send publish
				e.publishMessage(event)

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
			log.Println("engine: unexpected disconnect")

			client := e.clients[event.ClientId]
			if client != nil {
				// if will message is set for client - send this message to all clients
				if client.will != nil {
					will := transport.Event{
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

func (e *Engine) publishMessage(event transport.Event) {
	event.PacketType = packet.PUBLISH

	for _, client := range e.clients {
		if client != nil {
			if client.Contains(event.Topic.Name) {
				client.clientChan <- event

				if event.Qos > 0 {
					e.send[event.MessageId] = event
				}
			}
		}
	}
}

func (e *Engine) saveClient(id string, conn net.Conn, session bool) *Client {
	if e.clients[id] != nil {
		e.clients[id].conn = conn
	} else {
		client := NewClient(id, conn, session, e.channel)
		e.clients[id] = client
	}

	return e.clients[id]
}