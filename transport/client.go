package transport

import (
	"log"
	"net"
	"time"

	"github.com/MajaSuite/mqtt/packet"
)

type MqttClient struct {
	debug    bool
	conn     net.Conn
	clientId string
	Broker   chan packet.Packet // messages from Broker
	Sendout  chan packet.Packet // messages to Broker
	quit     bool
}

func Connect(addr string, clientId string, keepAlive uint16, login string, pass string, debug bool) (*MqttClient, error) {
	// socket connect
	c, err := net.Dial("tcp4", addr)
	if err != nil {
		return nil, err
	}

	// send connect
	cp := packet.NewConnect()
	cp.Version = 4
	cp.VersionName = "MQTT"
	cp.ClientID = clientId
	cp.KeepAlive = keepAlive
	cp.Username = login
	cp.Password = pass

	if err := packet.WritePacket(c, cp, debug); err != nil {
		return nil, err
	}

	resp, err := packet.ReadPacket(c, debug)
	if err != nil {
		return nil, err
	}

	if resp == nil || resp.Type() != packet.CONNACK {
		return nil, packet.ErrProtocolError
	}

	if resp.(*packet.ConnAckPacket).ReturnCode == 0 {
		mqttClient := &MqttClient{
			debug:    debug,
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
	if !c.quit {
		c.Sendout <- packet.NewDisconnect()
		c.quit = true
	}
}

func (c *MqttClient) Subscribe(pkt *packet.SubscribePacket) {
	if !c.quit {
		c.Sendout <- pkt
	}
}

func (c *MqttClient) Unsubscribe(pkt *packet.UnSubscribePacket) {
	if !c.quit {
		c.Sendout <- pkt
	}
}

func (c *MqttClient) Publish(pkt *packet.PublishPacket) {
	if !c.quit {
		c.Sendout <- pkt
	}
}

func (c *MqttClient) Send(pkt packet.Packet) {
	if !c.quit {
		c.Sendout <- pkt
	}
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
					log.Println("disconnect from broker")
					packet.WritePacket(c.conn, pkt, c.debug)
				case packet.PING:
					if err := packet.WritePacket(c.conn, pkt, c.debug); err != nil {
						log.Println("error send packet to broker, broker drop connection")
						c.quit = true
					}
				case packet.SUBSCRIBE:
					if err := packet.WritePacket(c.conn, pkt, c.debug); err != nil {
						log.Println("error send packet to broker, broker drop connection")
						c.quit = true
					}
				case packet.UNSUBSCRIBE:
					if err := packet.WritePacket(c.conn, pkt, c.debug); err != nil {
						log.Println("error send packet to broker, broker drop connection")
						c.quit = true
					}
				case packet.PUBLISH:
					if err := packet.WritePacket(c.conn, pkt, c.debug); err != nil {
						log.Println("error send packet to broker, broker drop connection")
						c.quit = true
					}
				}
			}
		}()

		if pkt, err = packet.ReadPacket(c.conn, c.debug); err != nil {
			log.Println("err read packet, disconnected: ", err.Error())
			c.quit = true
			continue
		}

		switch pkt.Type() {
		case packet.PUBLISH:
			c.Broker <- pkt
		case packet.PUBACK:

		case packet.PUBREC:

		case packet.PUBCOMP:

		case packet.DISCONNECT:
			c.quit = true
		}
	}
}

func (c *MqttClient) pinger(keepAlive time.Duration) {
	var nextPing = time.Now().Add(time.Second*keepAlive - 1)
	for {
		time.Sleep(time.Second)

		if time.Now().After(nextPing) {
			nextPing = time.Now().Add(time.Second * keepAlive)

			if c.quit {
				log.Println("pinger quit")
				return
			}

			c.Sendout <- packet.NewPing()
		}
	}
}
