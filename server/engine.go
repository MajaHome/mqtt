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
	channel chan Event
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
		channel: make(chan Event),
	}
}

func (e Engine) Process(server *Server) {
	go e.manageClients()

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
					} else {
						res.Session = false
					}
					client = e.saveClient(cp.ClientID, conn, res.Session)
					if cp.Will != nil && res.Session {
						client.saveWill(cp.Will.QoS, cp.Will.Retain)
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
				client.Start(server)
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
		log.Println("engine " + event.String())

		switch event.packetType {
		case packet.DISCONNECT:
			client := e.clients[event.clientId]
			if !client.session {
				delete(e.clients, event.clientId)
			}
			break
		case packet.SUBSCRIBE:
			break
		case packet.UNSUBSCRIBE:
			break
		case packet.PUBLISH:
			/*
				RETAIN – при публикации данных с установленным флагом retain, брокер сохранит его. При
				следующей подписке на этот топик брокер незамедлительно отправит сообщение с этим флагом.

				DUP – флаг дубликата устанавливается, когда клиент или MQTT брокер совершает повторную
				отправку пакета (используется в типах PUBLISH, SUBSCRIBE, UNSUBSCRIBE, PUBREL). При
				установленном флаге переменный заголовок должен содержать Message ID (идентификатор
				сообщения)

				QoS 0 At most once. На этом уровне издатель один раз отправляет сообщение брокеру и не ждет
				подтверждения от него, то есть отправил и забыл.

				QoS 1 At least once. Этот уровень гарантирует, что сообщение точно будет доставлено брокеру,
				но есть вероятность дублирования сообщений от издателя. После получения дубликата сообщения,
				брокер снова рассылает это сообщение подписчикам, а издателю снова отправляет подтверждение
				о получении сообщения. Если издатель не получил PUBACK сообщения от брокера, он повторно
				отправляет этот пакет, при этом в DUP устанавливается «1».
				PUBLISH -> PUBACK

				QoS 2 Exactly once. На этом уровне гарантируется доставка сообщений подписчику и исключается
				возможное дублирование отправленных сообщений.
				PUBLISH ->PUBREC, PUBREL - PUBCOMP

				Издатель отправляет сообщение брокеру. В этом сообщении указывается уникальный Packet ID,
				QoS=2 и DUP=0. Издатель хранит сообщение неподтвержденным пока не получит от брокера ответ
				PUBREC. Брокер отвечает сообщением PUBREC в котором содержится тот же Packet ID. После его
				получения издатель отправляет PUBREL с тем же Packet ID. До того, как брокер получит PUBREL
				он должен хранить копию сообщения у себя. После получения PUBREL он удаляет копию сообщения
				и отправляет издателю сообщение PUBCOMP о том, что транзакция завершена.
			*/
			client := e.clients[event.clientId]
			res := Event{
				packetType: packet.PUBLISH,
				messageId:  event.messageId,
				topic:      event.topic,
				payload:    event.payload,
				qos:        event.qos,
				retain:     event.retain,
				dublicate:  event.dublicate,
			}
			client.channel <- res
		case packet.PUBACK:
			break
		case packet.PUBREC:
			break
		case packet.PUBREL:
			break
		case packet.PUBCOMP:
			break
		default:
			// unexpected disconnect, send will message

			/*
				Will Flag - при установленном флаге, после того, как клиент отключится от брокера без отправки команды DISCONNECT
				(в случаях непредсказуемого обрыва связи и т.д.), брокер оповестит об этом всех подключенных к нему клиентов через
				так называемый Will Message.
			*/

			break
		}
	}

	log.Println("closed communication channel")
}

func (e *Engine) saveClient(id string, conn net.Conn, session bool) *Client {
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
		client := NewClient(cid, conn, session, e.channel)
		e.clients[id] = client
	}

	return e.clients[id]
}
