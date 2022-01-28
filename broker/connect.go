package broker

import (
	"github.com/MajaSuite/mqtt/db"
	"github.com/MajaSuite/mqtt/packet"
	"github.com/MajaSuite/mqtt/transport"
	"log"
	"net"
)

func (e *Mqtt) processConnect(conn net.Conn) {
	pkt, err := packet.ReadPacket(conn, e.debug)
	if err != nil {
		log.Println("error read packet", err)
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
				log.Printf("authorisation failed: %s", err)
				res.ReturnCode = uint8(packet.ConnectBadUserPass)
			}
		}

		if len(connPacket.ClientID) == 0 && !connPacket.CleanSession {
			res.ReturnCode = uint8(packet.ConnectIndentifierRejected)
		}

		if !connPacket.CleanSession {
			res.Session = true
			if len(connPacket.Username) == 0 {
				res.ReturnCode = uint8(packet.ConnectNotAuthorized)
			} else {
				connPacket.ClientID = connPacket.Username
			}
		}
		err = packet.WritePacket(conn, res, e.debug)
		if err != nil {
			log.Println("error send response packet", err)
			conn.Close()
			return
		}

		// close connection if not authorized
		if res.ReturnCode != uint8(packet.ConnectAccepted) {
			log.Println("error connection declined")
			conn.Close()
			return
		}

		client := NewClient(e.debug, connPacket.ClientID, conn, res.Session, e.brokerChannel)
		if e.clients[client.clientId] != nil {
			log.Println("override")
			// override connection. onlu one connection with user/pass allowed
			e.clients[client.clientId].conn.Close()
			e.clients[client.clientId].conn = conn
		} else {
			// restore subscription
			if res.Session {
				if subs, err := db.FetchSubcription(connPacket.ClientID); err == nil {
					for topic, qos := range subs {
						client.messageId++
						e.brokerChannel <- transport.Event{
							PacketType: packet.SUBSCRIBE,
							MessageId:  client.messageId,
							ClientId:   connPacket.ClientID,
							Topic:      transport.EventTopic{Name: topic, Qos: qos},
							Restore:    true,
						}
					}
				}
			}
			e.clients[client.clientId] = client
		}

		if connPacket.Will != nil {
			client.will = connPacket.Will
		}

		// start manage client
		go client.Start()
	} else {
		log.Println("error: wrong packet. expect CONNECT")
		conn.Close()
		return
	}
}
