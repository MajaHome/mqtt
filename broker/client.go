package broker

import (
	"fmt"
	"github.com/MajaSuite/mqtt/packet"
	"log"
	"net"
	"strings"
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
	go func() {
		for {
			if pkt, err := packet.ReadPacket(c.conn, c.debug); err != nil || pkt == nil {
				c.broker <- &packet.PacketImpl{ClientId: c.clientId}
				log.Printf("%s error read packet, disconnected: %s", c.clientId, err)
				return
			} else {
				pkt.SetSource(c.clientId)
				c.broker <- pkt

				if pkt.Type() == packet.DISCONNECT {
					return
				}
			}
		}
	}()

	for p := range c.channel {
		if c.debug {
			log.Printf("%s message to send %s", c.clientId, p)
		}

		if err := packet.WritePacket(c.conn, p, c.debug); err != nil {
			c.broker <- &packet.PacketImpl{ClientId: c.clientId} // send to toEngine unexpected disconnect
			log.Printf("%s disconnect while write to socket %s", c.clientId, err)
			return
		}
	}
	if c.debug {
		log.Printf("client %s stopped", c.clientId)
	}
}

func (c *Client) Stop() {
	close(c.channel)
	c.conn.Close()
}

func (c *Client) addSubscription(t packet.SubscribePayload) packet.QoS {
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

func (c *Client) Contains(topic string) packet.QoS {
	var ret = packet.QoS(0x80)

	if len(c.subscription) == 0 {
		return ret
	}

	t := strings.Split(topic, "/")

	// searched topic is empty
	if len(t) == 0 {
		return ret
	}

	for _, subs := range c.subscription {
		var i int = 0 // start from first level
		s := strings.Split(subs.Topic, "/")

		for {
			if len(s) <= i || len(t) <= i {
				break
			}

			// subscribed to any topic
			if s[i] == "#" {
				ret = subs.QoS
				break
			}

			// match at this level
			if s[i] == "*" || s[i] == t[i] {
				if len(t) == i+1 {
					if len(t) == len(s) {
						ret = subs.QoS
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

		if ret != packet.QoS(0x80) {
			return ret
		}
	}

	return ret
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
