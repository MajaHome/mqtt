package broker

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/MajaSuite/mqtt/packet"
)

type Client struct {
	debug        bool
	conn         net.Conn
	messageId    uint16
	clientId     string
	session      bool                      // persisted session (true) or clean (false)
	subscription []packet.SubscribePayload // subscribed topics
	ack          map[string]packet.Packet
	will         *packet.WillMessage
	channel      chan packet.Packet // channel to send message to client over connection
	broker       chan packet.Packet // channel to send message to broker
}

func NewClient(conn net.Conn, id string, session bool, broker chan packet.Packet, debug bool) *Client {
	return &Client{
		debug:        debug,
		conn:         conn,
		messageId:    1,
		clientId:     id,
		session:      session,
		subscription: []packet.SubscribePayload{},
		ack:          make(map[string]packet.Packet),
		channel:      make(chan packet.Packet),
		broker:       broker,
	}
}

func (c *Client) Start() {
	var err error
	var quit = make(chan bool)

	for {
		// send messages to client
		go func(quit chan bool) {
			for {
				select {
				case <-quit:
					c.conn.Close()
					return
				case pkt := <-c.channel:
					if c.debug {
						log.Printf("%s message to send %s", c.clientId, pkt)
					}

					if err := packet.WritePacket(c.conn, pkt, c.debug); err != nil {
						c.broker <- &packet.PacketImpl{ClientId: c.clientId} // send to toEngine unexpected disconnect
						log.Printf("%s disconnect while write to socket %s", c.clientId, err)
						c.Stop()
						return
					}
				}
			}
		}(quit)

		var pkt packet.Packet
		if pkt, err = packet.ReadPacket(c.conn, c.debug); err != nil {
			c.broker <- &packet.PacketImpl{ClientId: c.clientId} // send to toEngine unexpected disconnect
			log.Printf("%s error read packet, disconnected: %s", c.clientId, err)
			quit <- true
			return
		}

		pkt.SetSource(c.clientId)

		switch pkt.Type() {
		case packet.DISCONNECT:
			c.broker <- pkt
			quit <- true
			return
		case packet.PING:
			c.channel <- packet.NewPong()
		case packet.SUBSCRIBE:
			res := packet.NewSubAck()
			res.Id = pkt.(*packet.SubscribePacket).Id
			for _, payload := range pkt.(*packet.SubscribePacket).Topics {
				res.ReturnCodes = append(res.ReturnCodes, c.addSubscription(payload))
			}
			c.broker <- pkt
			c.channel <- res
		case packet.UNSUBSCRIBE:
			res := packet.NewUnSubAck()
			res.Id = pkt.(*packet.UnSubscribePacket).Id
			for _, topic := range pkt.(*packet.UnSubscribePacket).Topics {
				if c.removeSubscription(topic) {
					c.channel <- res
				}
			}
		case packet.PUBLISH:
			c.broker <- pkt
		case packet.PUBREC:
			c.broker <- pkt
		case packet.PUBREL:
			c.broker <- pkt
		case packet.PUBACK:
			c.broker <- pkt
		case packet.PUBCOMP:
			c.broker <- pkt
		default:
			err = packet.ErrUnknownPacket
		}

		if err != nil {
			log.Printf("%s unable to manage request, disconnect (%s)", c.clientId, err)
			c.broker <- &packet.PacketImpl{ClientId: c.clientId} // send to toEngine unexpected disconnect
			quit <- true
			return
		}
	}
}

func (c *Client) Stop() {
	c.conn.Close()
}

func (c *Client) addSubscription(t packet.SubscribePayload) packet.QoS {
	// check for duplicate
	for _, v := range c.subscription {
		if v.Topic == t.Topic {
			if v.QoS != t.QoS {
				v.QoS = t.QoS
			}
			return t.QoS
		}
	}
	c.subscription = append(c.subscription, t)
	return t.QoS
}

func (c *Client) removeSubscription(t packet.SubscribePayload) bool {
	if len(c.subscription) == 0 {
		return false
	}

	for i, v := range c.subscription {
		if v.Topic == t.Topic && v.QoS == t.QoS {
			if len(c.subscription) > i+1 {
				c.subscription[i] = c.subscription[len(c.subscription)-1]
			}
			c.subscription = c.subscription[:len(c.subscription)-1]
			return true
		}
	}
	return false
}

func (c *Client) Contains(topic string) bool {
	if len(c.subscription) == 0 {
		return false
	}

	t := strings.Split(topic, "/")

	// searched topic is empty
	if len(t) == 0 {
		return false
	}

	for _, subs := range c.subscription {
		var i int = 0 // start from first level
		s := strings.Split(subs.Topic, "/")

		var found bool
		for {
			if len(s) <= i || len(t) <= i {
				break
			}

			// subscribed to any topic
			if s[i] == "#" {
				found = true
				break
			}

			// match at this level
			if s[i] == "*" || s[i] == t[i] {
				if len(t) == i+1 {
					if len(t) == len(s) {
						found = true
						break
					}
					break
				}

				// try next level
				i++
				continue
			}

			// doesn't match
			break
		}

		if found {
			return true
		}
	}

	return false
}

func (c *Client) String() string {
	var will string
	if c.will != nil {
		will = fmt.Sprintf(" will: %s,", c.will.String())
	}

	var subs string
	for _, v := range c.subscription {
		subs += v.String() + ", "
	}

	return fmt.Sprintf("client {clientId: %s, session: %v,%s subscription: [%s]}",
		c.clientId, c.session, will, subs)
}
