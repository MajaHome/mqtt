package broker

import (
	"github.com/MajaSuite/mqtt/db"
	"github.com/MajaSuite/mqtt/packet"
	"github.com/MajaSuite/mqtt/transport"
	"log"
	"net"
)

func (e *Engine) processConnect(conn net.Conn) {
	pkt, err := packet.ReadPacket(conn, debug)
	if err != nil {
		log.Println("err read packet", err.Error())
		conn.Close()
		return
	}

	if pkt.Type() == packet.CONNECT {
		res := packet.NewConnAck()

		connPacket := pkt.(*packet.ConnPacket)
		id := connPacket.ClientID

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
			err := db.CheckAuth(connPacket.Username, connPacket.Password)
			if err != nil {
				log.Printf("auth fail: %s", err)
				res.ReturnCode = uint8(packet.ConnectBadUserPass)
			} else {
				id = connPacket.Username
			}
		}

		var client *Client
		if res.ReturnCode == uint8(packet.ConnectAccepted) {
			res.Session = !connPacket.CleanSession
			client = e.saveClient(id, conn, res.Session)
		}

		err = packet.WritePacket(conn, res, debug)
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
		go client.Start()

		if res.Session {
			if connPacket.Will != nil {
				client.will = connPacket.Will
			}

			// restore subscription
			if subs, err := db.FetchSubcription(id); err == nil {
				for topic, qos := range subs {
					e.channel <- transport.Event{
						PacketType: packet.SUBSCRIBE,
						ClientId:   id,
						Topic:      transport.EventTopic{Name: topic, Qos: qos},
						Restore:    true,
					}
				}
			}
		}
	} else {
		log.Println("wrong packet. expect CONNECT")
		conn.Close()
		return
	}
}
