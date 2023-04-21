package broker

import (
	"log"
	"net"

	"github.com/MajaSuite/mqtt/db"
	"github.com/MajaSuite/mqtt/packet"
)

func (b *Broker) newConnection(conn net.Conn) {
	pkt, err := packet.ReadPacket(conn, b.debug)
	if err != nil {
		log.Println("new connection: error read packet", err)
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

		res.ReturnCode = uint8(packet.ConnectAccepted)

		// check authorization
		if len(connPacket.Username) > 0 {
			err := db.CheckAuth(connPacket.Username, connPacket.Password)
			if err != nil {
				log.Printf("new connection: authorisation failed: %s", err)
				res.ReturnCode = uint8(packet.ConnectBadUserPass)
			}
		}

		if !connPacket.CleanSession {
			if len(connPacket.ClientID) == 0 {
				res.ReturnCode = uint8(packet.ConnectIndentifierRejected)
			} else {
				if len(connPacket.Username) == 0 {
					res.ReturnCode = uint8(packet.ConnectNotAuthorized)
				} else {
					res.Session = true
					connPacket.ClientID = connPacket.Username
				}
			}
		}

		err = packet.WritePacket(conn, res, b.debug)
		if err != nil {
			log.Println("new connection: error send response packet", err)
			conn.Close()
			return
		}

		// close connection if not authorized
		if res.ReturnCode != uint8(packet.ConnectAccepted) {
			log.Println("new connection: error connection declined")
			conn.Close()
			return
		}

		if res.Session { // statefull session
			if b.clients[connPacket.ClientID] != nil {
				// session already exists in the broker memory, use it for current connection
				b.clients[connPacket.ClientID].conn.Close()
				b.clients[connPacket.ClientID].conn = conn
			} else {
				// the new one
				b.clients[connPacket.ClientID] = NewClient(conn, connPacket.ClientID, res.Session, b.channel, b.debug)

				// and restore subscription
				if subs, err := db.FetchSubcription(connPacket.ClientID); err == nil {
					subPayload := []packet.SubscribePayload{}
					for topic, qos := range subs {
						subPayload = append(subPayload, packet.SubscribePayload{Topic: topic, QoS: packet.QoS(qos)})
					}
					b.clients[connPacket.ClientID].messageId++

					subpkt := packet.NewSubscribe()
					subpkt.ClientId = connPacket.ClientId
					subpkt.Id = b.clients[connPacket.ClientID].messageId
					subpkt.Topics = subPayload
					b.channel <- subpkt
				}

				// save will packet
				if connPacket.Will != nil {
					b.clients[connPacket.ClientID].will = connPacket.Will
				}
			}
		} else { // new clean session
			// remove all (if we have)
			if b.clients[connPacket.ClientID] != nil {
				b.clients[connPacket.ClientID].conn.Close()
				delete(b.clients, connPacket.ClientID)
			}

			// and create new one
			b.clients[connPacket.ClientID] = NewClient(conn, connPacket.ClientID, res.Session, b.channel, b.debug)
		}

		if connPacket.Will != nil {
			b.clients[connPacket.ClientID].will = connPacket.Will
		}

		// start manage client
		b.clients[connPacket.ClientID].Start()
		return
	}

	log.Println("new connection: wrong packet. expect CONNECT")
	conn.Close()
	return
}
