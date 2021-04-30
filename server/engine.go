package server

import (
	"io"
	"log"
	"mqtt/packet"
	"time"
)

type Engine struct {
	backend *Backend
}

func GetEngine(backend *Backend) *Engine {
	return &Engine{
		backend: backend,
	}
}

func (e Engine) Manage(server *Server) {
	for {
		// accept next connection
		conn, err := server.Accept()
		if err != nil {
			log.Println("err accept", err.Error())
			conn.Close()
			continue
		}

		go func() {
			conn.SetDeadline(time.Now().Add(time.Minute * 5))

			pkt, err := server.ReadPacket(conn)
			if err != nil {
				if DEBUG {
					log.Println("err read packet", err.Error())
				}
				conn.Close()
				return
			}

			if pkt.Type() == packet.CONNECT {
				res := packet.NewConnAck()

				func(cp *packet.ConnPacket) {
					if cp.Version != 4 {
						res.ReturnCode = uint8(packet.ConnectUnacceptableProtocol)
						return
					}

					if len(cp.ClientID) == 0 && !cp.CleanSession {
						res.ReturnCode = uint8(packet.ConnectIndentifierRejected)
						return
					}

					if e.backend.Authorize(cp.Username, cp.Password) {
						res.ReturnCode = uint8(packet.ConnectAccepted)
					} else {
						res.ReturnCode = uint8(packet.ConnectBadUserPass)
					}

					res.Session = !cp.CleanSession
				}(pkt.(*packet.ConnPacket))

				err := server.WritePacket(conn, res)
				if err != nil {
					if DEBUG {
						log.Println("err write packet", err.Error())
					}
					conn.Close()
					return
				}

				if res.ReturnCode != uint8(packet.ConnectAccepted) {
					if DEBUG {
						log.Println("err connection declined")
					}
					conn.Close()
					return
				}
			} else {
				if DEBUG {
					log.Println("wrong packet. expect CONNECT")
				}
				conn.Close()
				return
			}

			for {
				pkt, err := server.ReadPacket(conn)
				if err != nil {
					break
				}

				err = func(p packet.Packet) error {
					switch pkt.Type() {
					case packet.DISCONNECT:
						res := packet.NewDisconnect()
						server.WritePacket(conn, res)
						return io.EOF
					case packet.PING:
						res := packet.NewPong()
						return server.WritePacket(conn, res)
					case packet.SUBSCRIBE:
						r := pkt.(*packet.SubscribePacket)
						res := packet.NewSubAck()

						res.Id = r.Id
						var qos []packet.QoS
						for _, q := range r.Topics {
							qos = append(qos, q.QoS)
						}
						res.ReturnCodes = qos

						return server.WritePacket(conn, res)
					case packet.UNSUBSCRIBE:
						//r := pkt.(*packet.UnSubscriibePacket)
						res := packet.NewUnSubAck()

						return server.WritePacket(conn, res)
					case packet.PUBLISH:
						res := packet.NewPubAck()
						// if payload is empty - unsubscribe to the topic
						return server.WritePacket(conn, res)
					case packet.PUBCOMP:
						res := packet.NewPubAck()
						return server.WritePacket(conn, res)
					case packet.PUBREC:
						res := packet.NewPubAck()
						return server.WritePacket(conn, res)
					case packet.PUBREL:
						res := packet.NewPubAck()
						return server.WritePacket(conn, res)
					default:
						return packet.ErrUnknownPacket
					}
				}(pkt)

				if err != nil {
					if err != io.EOF {
						log.Println("err serve: ", err.Error())
					}
					log.Println("client disconnect")
					conn.Close()
					break
				}
			}
		}()
	}
}

func (e Engine) Close() {
	e.backend.Close()
}
