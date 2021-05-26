package transport

import (
	"github.com/MajaSuite/mqtt/packet"
	"log"
	"net"
	"time"
)

type MqttClient struct {
	conn     net.Conn
	Broker   chan *Event // messages from Broker
	sendout  chan *Event // messages to Broker
	clientId string
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
		mqttClient := &MqttClient{conn: c, Broker: make(chan *Event), sendout: make(chan *Event), clientId: cp.ClientID}
		go mqttClient.pinger(time.Duration(keepAlive))
		go mqttClient.client()

		return mqttClient, nil
	}

	return nil, packet.ErrConnect
}

func (c *MqttClient) Disconnect() {
	event := &Event{PacketType: packet.DISCONNECT}
	c.sendout <- event
}

func (c *MqttClient) Subscribe(*packet.SubscribePacket) {
	// save subscription?
	// send to c.sendout
}

func (c *MqttClient) Unsubscribe(*packet.UnSubscribePacket) {
	// send to c.sendout
}

func (c *MqttClient) Publish(*packet.PublishPacket) {
	// send to c.sendout
}

func (c *MqttClient) client() {
	var pkt packet.Packet
	var err error

	for {
		if c.quit {
			log.Println("client quit")
			close(c.Broker)
			close(c.sendout)
			return
		}

		go func() {
			for event := range c.sendout {
				log.Println("to send: ", event)
				switch event.PacketType {
				case packet.PING:
					if err := WritePacket(c.conn, packet.NewPing()); err != nil {
						log.Println("error write response, server drop connection")
						c.quit = true
					}
				case packet.DISCONNECT:
					log.Println("disconnect from server")
					WritePacket(c.conn, packet.NewDisconnect())
					c.quit = true
				case packet.SUBSCRIBE:
				case packet.UNSUBSCRIBE:
				case packet.PUBLISH:
				}
			}
		}()

		if pkt, err = ReadPacket(c.conn); err != nil {
			log.Println("err read packet, disconnected: ", err.Error())
			c.quit = true
			continue
		}

		switch pkt.Type() {
		case packet.PONG:
			log.Println("pong received")
			c.Broker <- &Event{PacketType: packet.PONG}
		case packet.PUBLISH:
		case packet.PUBACK:
		case packet.PUBREC:
		case packet.PUBCOMP:
		default:
			log.Println("wrong packet from broker")
		}
	}
}

func (c *MqttClient) pinger(keepAlive time.Duration) {
	var nextPing = time.Now().Add(time.Second*keepAlive-1)
	for {
		time.Sleep(time.Second)

		if time.Now().After(nextPing) {
			nextPing = time.Now().Add(time.Second*keepAlive)
			c.sendout <- &Event{PacketType: packet.PING}
		}

		if c.quit {
			log.Println("pinger quit")
			return
		}
	}
}