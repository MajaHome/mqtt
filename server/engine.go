package server

import (
	"crypto/rand"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"log"
	"mqtt/model"
	"mqtt/packet"
	"net"
)

type Engine struct {
	db      *gorm.DB
	clients map[string]*Client
	channel chan *Event
}

func NewEngine() *Engine {
	log.Println("initialize database")
	conn, err := gorm.Open("sqlite3", "mqtt.db")
	if err != nil {
		panic(err)
	}

	log.Println("migrate database")
	conn.AutoMigrate(&model.User{})

	e := &Engine{
		db:      conn,
		clients: make(map[string]*Client),
		channel: make(chan *Event),
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
			// event := &Event{clientId: c.clientId, packetType: pkt.Type(), topic: t}
			// TODO send retain message if exists for this topic
			break
		case packet.PUBLISH: // in
			// TODO if RETAIN set - save message to retain queue (and push on subscribe)

			switch packet.QoS(event.qos) {
			case packet.AtMostOnce:
				e.publishMessage(event)
			case packet.AtLeastOnce: // PUBLISH -> PUBACK
				res := &Event{packetType: packet.PUBACK, messageId: event.messageId, clientId: event.clientId}
				e.clients[event.clientId].clientChan <- res

				// send published message to all subscribed clients
				e.publishMessage(event)
			case packet.ExactlyOnce: // PUBLISH ->PUBREC, PUBREL - PUBCOMP
				res := &Event{packetType: packet.PUBREC, messageId: event.messageId, clientId: event.clientId}
				e.clients[event.clientId].clientChan <- res

				// todo: record packet in delivery queue
			}
		case packet.PUBACK: // out
			// we receive answer on publish command,
			// TODO remove from send queue
			break
		case packet.PUBREC: // out
			// we receive answer on PUUBLISH with qos2,

			// TODO send PUBREL

			break
		case packet.PUBREL: // in
			// TODO remove from delivery queue

			// send answer PUBCOMP
			res := &Event{packetType: packet.PUBCOMP, messageId: event.messageId, clientId: event.clientId}
			e.clients[event.clientId].clientChan <- res

			// TODO send publish
			//todo				e.publishMessage(storedEvent)
		case packet.PUBCOMP: // out
			// we receive answer on PUUBREL

			// TODO remove from sent queue with qos2
			break
		default:
			log.Println("engineChan: unexpected disconnect")

			client := e.clients[event.clientId]
			if client != nil {
				// todo is not clean session - save subscription

				// TODO if will message set for client - send this message to all clients

				if !client.session {
					delete(e.clients, event.clientId)
				}
			}
		}
	}
}

func (e *Engine) publishMessage(event *Event) {
	for _, client := range e.clients {
		if client != nil && client.isSubscribed(event.topic.name) {
			res := &Event{
				packetType: packet.PUBLISH,
				clientId:   event.clientId,
				messageId:  event.messageId,
				topic:      event.topic,
				payload:    event.payload,
				qos:        event.qos,
				retain:     event.retain,
				dublicate:  event.dublicate,
			}
			client.clientChan <- res
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
