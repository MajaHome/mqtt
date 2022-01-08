package transport

import (
	"log"
	"net"
	"time"

	"github.com/MajaSuite/mqtt/packet"
)

type Stage byte

const (
	CONNECTED Stage = iota
	DISCONNECTED
	RECONNECT
)

type MqttClient struct {
	debug     bool
	conn      net.Conn
	clientId  string
	keepalive uint16
	address   string
	login     string
	password  string
	Broker    chan packet.Packet // messages from Broker
	Sendout   chan packet.Packet // messages to Broker
	stage     Stage
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
			debug:     debug,
			conn:      c,
			address:   addr,
			clientId:  clientId,
			keepalive: keepAlive,
			login:     login,
			password:  pass,
			Broker:    make(chan packet.Packet),
			Sendout:   make(chan packet.Packet),
			stage:     CONNECTED,
		}

		return mqttClient, nil
	}

	return nil, packet.ErrConnect
}

func (c *MqttClient) pinger(keepAlive time.Duration) {
	var nextPing = time.Now().Add(time.Second*keepAlive - 1)
	for {
		if c.stage == DISCONNECTED {
			log.Println("stop mqtt pinger")
			return
		}
		time.Sleep(time.Second)

		if time.Now().After(nextPing) {
			nextPing = time.Now().Add(time.Second * keepAlive)
			c.Sendout <- packet.NewPing()
		}
	}
}

func (c *MqttClient) reconnect() error {
	if c.stage == DISCONNECTED {
		return packet.ErrConnect
	}
	log.Println("try reconnect to mqtt server")

	if c.stage == RECONNECT {
		for c.stage == RECONNECT {
			time.Sleep(time.Second / 10)
		}
		if c.stage == DISCONNECTED {
			return packet.ErrConnect
		}
		return nil
	}

	c.stage = RECONNECT
	c.conn.Close()
	cp := packet.NewConnect()
	cp.Version = 4
	cp.VersionName = "MQTT"
	cp.ClientID = c.clientId
	cp.KeepAlive = c.keepalive
	cp.Username = c.login
	cp.Password = c.password

	for {
		if conn, err := net.Dial("tcp4", c.address); err != nil {
			continue
		} else {
			c.conn = conn
		}

		if err := packet.WritePacket(c.conn, cp, c.debug); err != nil {
			c.conn.Close()
			continue
		}

		resp, err := packet.ReadPacket(c.conn, c.debug)
		if err != nil {
			c.conn.Close()
			continue
		}

		if resp == nil || resp.Type() != packet.CONNACK {
			c.stage = DISCONNECTED
			return packet.ErrConnect
		}
		if resp.(*packet.ConnAckPacket).ReturnCode != 0 {
			c.stage = DISCONNECTED
			return packet.ErrConnect
		} else {
			break
		}
	}

	c.stage = CONNECTED

	return nil
}

func (c *MqttClient) Start() {
	// send message to server
	go func() {
		for pkt := range c.Sendout {
			if c.stage == DISCONNECTED {
				return
			}
			if pkt.Type() == packet.DISCONNECT {
				log.Println("disconnect from broker")
				packet.WritePacket(c.conn, pkt, c.debug)
				c.stage = DISCONNECTED
				c.conn.Close()
				return
			} else {
				log.Println("send to server ", pkt.String())
				if err := packet.WritePacket(c.conn, pkt, c.debug); err != nil {
					if c.reconnect() == packet.ErrConnect {
						return
					}
					packet.WritePacket(c.conn, pkt, c.debug)
				}
			}
		}
	}()

	go c.pinger(time.Duration(c.keepalive))

	// read from network
	for {
		if c.stage == DISCONNECTED {
			log.Println("stop mqtt client")
			close(c.Broker)
			close(c.Sendout)
			c.conn.Close()
			return
		}

		pkt, err := packet.ReadPacket(c.conn, c.debug)
		if err != nil {
			if c.reconnect() == packet.ErrConnect {
				return
			}
			pkt, err = packet.ReadPacket(c.conn, c.debug)
		}

		// manage received packet
		switch pkt.Type() {
		case packet.PUBLISH:
			c.Broker <- pkt
			switch pkt.(*packet.PublishPacket).QoS {
			case packet.AtLeastOnce: // PUBLISH -> PUBACK
				p := packet.NewPubAck()
				p.Id = pkt.(*packet.PublishPacket).Id
				c.Sendout <- p
			case packet.ExactlyOnce: // PUBLISH ->PUBREC, PUBREL - PUBCOMP
				p := packet.NewPubRec()
				p.Id = pkt.(*packet.PublishPacket).Id
				c.Sendout <- p
			}
		case packet.PUBACK:
			// answer on our publish
		case packet.PUBREC:
			// answer on our PUBLISH with qos2
		case packet.PUBREL:
			// answer on PUBREC
			p := packet.NewPubComp()
			p.Id = pkt.(*packet.PubRelPacket).Id
			c.Sendout <- p
		case packet.PUBCOMP:
			// answer on our PUBREL
		case packet.DISCONNECT:
			c.stage = DISCONNECTED
		}
	}
}
