package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/MajaSuite/mqtt/packet"
)

type ClientConnection struct {
	debug         bool
	Receive       chan packet.Packet // messages from Broker
	Send          chan packet.Packet // messages to Broker
	conn          net.Conn
	connectAddr   string
	connectPacket *packet.ConnPacket
}

func Connect(addr string, clientId string, keepAlive uint16, session bool, login string, pass string, debug bool) (*ClientConnection, error) {
	connPacket := packet.NewConnect()
	connPacket.Version = 4
	connPacket.VersionName = "MQTT"
	connPacket.ClientID = clientId
	connPacket.KeepAlive = keepAlive
	connPacket.Username = login
	connPacket.Password = pass
	connPacket.CleanSession = !session

	cc := &ClientConnection{
		debug:         debug,
		Receive:       make(chan packet.Packet),
		Send:          make(chan packet.Packet, 2),
		connectAddr:   addr,
		connectPacket: connPacket,
	}

	if err := cc.connect(); err != nil {
		return nil, err
	}

	go cc.sendout()
	go cc.manage()

	return cc, nil
}

func (cc *ClientConnection) connect() error {
	log.Println("connect ...")
	conn, err := net.DialTimeout("tcp4", cc.connectAddr, time.Second)
	if err != nil {
		return err
	}

	if err := packet.WritePacket(conn, cc.connectPacket, cc.debug); err != nil {
		return err
	}

	pkt, err := packet.ReadPacket(conn, cc.debug)
	if err != nil {
		cc.conn.Close()
		return err
	}

	if pkt.(*packet.ConnAckPacket).ReturnCode == 0 {
		if cc.debug {
			log.Println("connected")
		}

		cc.conn = conn
		return nil
	} else {
		packet.WritePacket(conn, packet.NewDisconnect(), cc.debug)
		cc.conn.Close()
		return fmt.Errorf("wrong response from server, %d", pkt.(*packet.ConnAckPacket).ReturnCode)
	}
}

func (cc *ClientConnection) pinger(quit chan bool) {
	var keepAlive = time.Second*time.Duration(cc.connectPacket.KeepAlive) - 1
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
		time.Sleep(time.Second)
		if time.Now().After(nextPing) {
			nextPing = time.Now().Add(keepAlive)
			cc.Send <- packet.NewPing()
		}
	}
}

func (cc *ClientConnection) sendout() {
	pinger := make(chan bool)
	go cc.pinger(pinger)

	for pkt := range cc.Send {
		if pkt.Type() == packet.DISCONNECT {
			packet.WritePacket(cc.conn, packet.NewDisconnect(), cc.debug)
			if cc.debug {
				log.Println("receive disconnect packet. stop connection")
			}
			pinger <- true
			cc.conn.Close()
			return
		}

		err := packet.WritePacket(cc.conn, pkt, cc.debug)
		if err != nil {
			pinger <- true
			cc.conn.Close()
			for {
				if err := cc.connect(); err == nil {
					break
				}
				time.Sleep(time.Second * 3)
			}

			go cc.pinger(pinger)
		}
	}
}

func (cc *ClientConnection) manage() {
	quit := make(chan bool)

	for {
		pkt, err := packet.ReadPacket(cc.conn, cc.debug)
		if err != nil {
			time.Sleep(time.Second)
			continue
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
			quit <- true
			cc.Send <- packet.NewDisconnect()
		}
	}
}
