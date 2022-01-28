package transport

import (
	"io"
	"log"
	"net"
	"time"

	"github.com/MajaSuite/mqtt/packet"
)

type ClientConnection struct {
	debug      bool
	addr       string
	connect    bool
	connPacket *packet.ConnPacket
	Receive    chan packet.Packet // messages from Broker
	Send       chan packet.Packet // messages to Broker
	conn       net.Conn
}

func Connect(addr string, clientId string, keepAlive uint16, session bool, login string, pass string, debug bool) *ClientConnection {
	conn, err := net.Dial("tcp4", addr)
	if err != nil {
		return nil
	}

	cp := packet.NewConnect()
	cp.Version = 4
	cp.VersionName = "MQTT"
	cp.ClientID = clientId
	cp.KeepAlive = keepAlive
	cp.Username = login
	cp.Password = pass
	cp.CleanSession = !session

	if err := packet.WritePacket(conn, cp, debug); err != nil {
		return nil
	}

	cc := &ClientConnection{
		debug:      debug,
		addr:       addr,
		connPacket: cp,
		Receive:    make(chan packet.Packet),
		Send:       make(chan packet.Packet),
		conn:       conn,
	}

	go cc.manage()

	return cc
}

func (cc *ClientConnection) pinger(quit chan bool) {
	var keepAlive = time.Second*time.Duration(cc.connPacket.KeepAlive) - 1
	var nextPing = time.Now().Add(keepAlive)

	for {
		select {
		case <-quit:
			if cc.debug {
				log.Println("pinger receive stop")
			}
			return
		default:
		}

		if time.Now().After(nextPing) {
			nextPing = time.Now().Add(keepAlive)
			cc.Send <- packet.NewPing()
		}
	}
}

func (cc *ClientConnection) sendout(quit chan bool) {
	pinger := make(chan bool)
	go cc.pinger(pinger)

	if cc.debug {
		log.Println("start client")
	}

	for {
		for pkt := range cc.Send {
		resend:
			select {
			case <-quit:
				if cc.debug {
					log.Println("sendout receive stop")
				}
				return
			default:
			}

			var err error
			if cc.conn != nil {
				err = packet.WritePacket(cc.conn, pkt, cc.debug)
				if pkt.Type() == packet.DISCONNECT {
					if cc.debug {
						log.Println("receive disconnect packet. stop connection")
					}
					pinger <- true
					cc.conn.Close()
					return
				}
			} else {
				err = io.ErrClosedPipe
			}

			if err != nil {
				if cc.conn != nil {
					pinger <- true // stop pinger
				}
				cc.conn, err = net.Dial("tcp4", cc.addr)
				if err != nil {
					time.Sleep(time.Second)
					goto resend
				}

				cc.connect = false
				if err := packet.WritePacket(cc.conn, cc.connPacket, cc.debug); err != nil {
					cc.conn.Close()
					goto resend
				}

				go cc.pinger(pinger)
			}
		}
	}
}

func (cc *ClientConnection) manage() {
	sendout := make(chan bool)
	cc.connect = false

	for {
		if cc.conn == nil {
			time.Sleep(time.Second)
			continue
		}
		pkt, err := packet.ReadPacket(cc.conn, cc.debug)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}

		// check for connect ack
		if !cc.connect {
			if pkt.(*packet.ConnAckPacket).ReturnCode == 0 {
				if cc.debug {
					log.Println("connected")
				}
				go cc.sendout(sendout)
				cc.connect = true
			} else {
				cc.Send <- packet.NewDisconnect()
				return
			}
		}

		// manage received packet
		switch pkt.Type() {
		case packet.PUBLISH:
			cc.Receive <- pkt
			switch pkt.(*packet.PublishPacket).QoS {
			case packet.AtLeastOnce: // PUBLISH -> PUBACK
				p := packet.NewPubAck()
				p.Id = pkt.(*packet.PublishPacket).Id
				cc.Send <- p
			case packet.ExactlyOnce: // PUBLISH ->PUBREC, PUBREL - PUBCOMP
				p := packet.NewPubRec()
				p.Id = pkt.(*packet.PublishPacket).Id
				cc.Send <- p
			}
		case packet.PUBACK:
			// answer on our publish
			// todo remove remove publish from queue
		case packet.PUBREC:
			// answer on our PUBLISH with qos2. publish was succesfull
			// todo send PUBREL? and remove publish from queue
		case packet.PUBREL:
			// answer on PUBREC
			p := packet.NewPubComp()
			p.Id = pkt.(*packet.PubRelPacket).Id
			cc.Send <- p
		case packet.PUBCOMP:
			// answer on our PUBREL
			// todo remove pubrel from queue
		case packet.DISCONNECT:
			sendout <- true
			cc.Send <- packet.NewDisconnect()
		}
	}
}
