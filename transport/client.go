package transport

import (
	"github.com/MajaSuite/mqtt/packet"
	"log"
	"net"
	"time"
)

type MqttClient struct {
	conn     net.Conn
	clientId string
	Broker   chan packet.Packet // messages from Broker
	Sendout  chan packet.Packet // messages to Broker
	quit     bool
}

func Connect(addr string, clientId string, keepAlive uint16, login string, pass string)  (*MqttClient, error) {
	// socket connect
	c, err := net.Dial("tcp4", addr)
	if err != nil {
		return nil, err
	}

	// send connect
	cp := packet.NewConnect()
	cp.Version = 4
	cp.ClientID = clientId
	cp.KeepAlive = keepAlive
	cp.Username = login
	cp.Password = pass
	if err := WritePacket(c, cp); err != nil {
		return nil, err
	}

	resp, err := ReadPacket(c)
	if err != nil {
		return nil, err
	}

	if resp == nil || resp.Type() != packet.CONNACK {
		return nil, packet.ErrProtocolError
	}

	if resp.(*packet.ConnAckPacket).ReturnCode == 0 {
		mqttClient := &MqttClient{
			conn:     c,
			clientId: cp.ClientID,
			Broker:   make(chan packet.Packet),
			Sendout:  make(chan packet.Packet),
		}

		go mqttClient.pinger(time.Duration(keepAlive))
		go mqttClient.client()

		return mqttClient, nil
	}

	return nil, packet.ErrConnect
}

func (c *MqttClient) Disconnect() {
	c.Sendout <- packet.NewDisconnect()
}

func (c *MqttClient) Subscribe(pkt *packet.SubscribePacket) {
	c.Sendout <- pkt
}

func (c *MqttClient) Unsubscribe(pkt *packet.UnSubscribePacket) {
	c.Sendout <- pkt
}

func (c *MqttClient) Publish(pkt *packet.PublishPacket) {
	c.Sendout <- pkt
}

func (c *MqttClient) client() {
	var pkt packet.Packet
	var err error

	for {
		if c.quit {
			log.Println("client quit")
			close(c.Broker)
			close(c.Sendout)
			return
		}

		go func() {
			for pkt := range c.Sendout {
				switch pkt.Type() {
				case packet.DISCONNECT:
					log.Println("disconnect from server")
					WritePacket(c.conn, pkt)
				case packet.PING:
					if err := WritePacket(c.conn, pkt); err != nil {
						log.Println("error send packet to broker, server drop connection")
						c.quit = true
					}
				case packet.SUBSCRIBE:
					if err := WritePacket(c.conn, pkt); err != nil {
						log.Println("error send packet to broker, server drop connection")
						c.quit = true
					}
				case packet.UNSUBSCRIBE:
					if err := WritePacket(c.conn, pkt); err != nil {
						log.Println("error send packet to broker, server drop connection")
						c.quit = true
					}
				case packet.PUBLISH:
					if err := WritePacket(c.conn, pkt); err != nil {
						log.Println("error send packet to broker, server drop connection")
						c.quit = true
					}
				}
			}
		}()

		if pkt, err = ReadPacket(c.conn); err != nil {
			log.Println("err read packet, disconnected: ", err.Error())
			c.quit = true
			continue
		}

		switch pkt.Type() {
		case packet.PUBLISH:
			c.Broker <- pkt
		case packet.PUBACK:
			;
		case packet.PUBREC:
			;
		case packet.PUBCOMP:
			;
		case packet.DISCONNECT:
			c.quit = true
		}
	}
}

func (c *MqttClient) pinger(keepAlive time.Duration) {
	var nextPing = time.Now().Add(time.Second*keepAlive-1)
	for {
		time.Sleep(time.Second)

		if time.Now().After(nextPing) {
			nextPing = time.Now().Add(time.Second*keepAlive)
			c.Sendout <- packet.NewPing()
		}

		if c.quit {
			log.Println("pinger quit")
			return
		}
	}
}