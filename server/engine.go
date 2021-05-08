package server

import (
	"crypto/rand"
	"log"
	"mqtt/model"
	"mqtt/packet"
	"net"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type Engine struct {
	db      *gorm.DB
	clients map[string]*Client
	channel chan string
}

func NewEngine() *Engine {
	log.Println("initialize database")
	conn, err := gorm.Open("sqlite3", "mqtt.db")
	if err != nil {
		panic(err)
	}

	log.Println("migrate database")
	conn.AutoMigrate(&model.User{})

	return &Engine{
		db:      conn,
		clients: make(map[string]*Client),
		channel: make(chan string),
	}
}

func (e Engine) Process(server *Server) {
	go e.manageChannel()

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Println("err accept", err.Error())
			conn.Close()
			continue
		}

		go func() {
			pkt, err := server.ReadPacket(conn)
			if err != nil {
				log.Println("err read packet", err.Error())
				conn.Close()
				return
			}

			if pkt.Type() == packet.CONNECT {
				res := packet.NewConnAck()

				cp := pkt.(*packet.ConnPacket)

				// check version, now only 4 (3.11)
				if cp.Version != 4 {
					res.ReturnCode = uint8(packet.ConnectUnacceptableProtocol)
					return
				}

				if len(cp.ClientID) == 0 && !cp.CleanSession {
					res.ReturnCode = uint8(packet.ConnectIndentifierRejected)
					return
				}

				res.ReturnCode = uint8(packet.ConnectAccepted)

				// check authorization
				if len(cp.Username) > 0 {
					var user model.User
					result := e.db.Take(&user, "user_name = ? and password = ?", cp.Username, cp.Password)
					if result.Error != nil {
						log.Println("username/password is wrong")
						res.ReturnCode = uint8(packet.ConnectBadUserPass)
					}
				}

				var client *Client
				if res.ReturnCode == uint8(packet.ConnectAccepted) {
					if !cp.CleanSession && e.clients[cp.ClientID] != nil {
						res.Session = !cp.CleanSession
						client = e.saveClient(cp.ClientID, conn)
						if cp.Will != nil {
							client.SaveWill(cp.Will.QoS, cp.Will.Retain)
						}
					} else {
						res.Session = false
						client = e.saveClient("", conn)
					}
				}

				err := server.WritePacket(conn, res)
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
				client.Start(server)
			} else {
				log.Println("wrong packet. expect CONNECT")
				conn.Close()
				return
			}
		}()
	}
}

func (e *Engine) manageChannel() {
	for msg := range e.channel {
		log.Println("engine " + msg)
	}

	log.Println("closed communication channel")
}

func (e *Engine) saveClient(id string, conn net.Conn) *Client {
	var cid string
	if id == "" {
		// generate temporary ident
		ident := make([]byte, 8)
		rand.Read(ident)
		cid = string(cid)
	} else {
		cid = id
	}

	if e.clients[id] != nil {
		e.clients[id].conn = conn
	} else {
		client := NewClient(cid, conn, e.channel)
		e.clients[id] = client
	}

	return e.clients[id]
}
