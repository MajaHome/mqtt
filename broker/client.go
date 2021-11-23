package broker

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/MajaSuite/mqtt/packet"
	"github.com/MajaSuite/mqtt/transport"
)

type Client struct {
	conn         net.Conn
	engineChan   chan transport.Event
	clientChan   chan transport.Event
	clientId     string
	session      bool                   // clean or persisted session
	subscription []transport.EventTopic // subscribed topics
	will         *packet.Message
	stop         bool // flag to stop
}

func NewClient(id string, conn net.Conn, session bool, channel chan transport.Event) *Client {
	return &Client{
		conn:       conn,
		engineChan: channel,
		clientChan: make(chan transport.Event),
		clientId:   id,
		session:    session,
	}
}

func (c *Client) Start() {
	var pkt packet.Packet
	var err error

	c.stop = false
	for {
		if c.stop {
			log.Printf("client %s disconnected\n", c.clientId)
			return
		}

		go func() {
			for e := range c.clientChan {
				log.Println("client receive message: " + e.String())

				var res packet.Packet
				switch e.PacketType {
				case packet.PUBLISH:
					res = packet.NewPublish()
					res.(*packet.PublishPacket).Id = e.MessageId
					res.(*packet.PublishPacket).Topic = e.Topic.Name
					res.(*packet.PublishPacket).QoS = packet.QoS(e.Topic.Qos)
					res.(*packet.PublishPacket).Payload = e.Payload
					res.(*packet.PublishPacket).Retain = e.Retain
					res.(*packet.PublishPacket).DUP = e.Dublicate
				case packet.PUBACK:
					res = packet.NewPubAck()
					res.(*packet.PubAckPacket).Id = e.MessageId
				case packet.PUBREC:
					res = packet.NewPubRec()
					res.(*packet.PubRecPacket).Id = e.MessageId
				case packet.PUBCOMP:
					res = packet.NewPubComp()
					res.(*packet.PubCompPacket).Id = e.MessageId
				default:
					log.Println("wrong packet from engine")
				}

				if err := packet.WritePacket(c.conn, res, debug); err != nil {
					c.engineChan <- transport.Event{ClientId: c.clientId} // send to engine unexpected disconnect
					log.Println("client disconnect while write to socket")
					c.Stop()
					return
				}
			}
		}()

		if pkt, err = packet.ReadPacket(c.conn, debug); err != nil {
			log.Println("err read packet, disconnected: ", err.Error())
			c.Stop()

			// send to engineChan unexpected disconnect
			c.engineChan <- transport.Event{ClientId: c.clientId}

			return
		}

		switch pkt.Type() {
		case packet.DISCONNECT:
			c.engineChan <- transport.Event{ClientId: c.clientId, PacketType: pkt.Type()}
			err = packet.WritePacket(c.conn, packet.NewDisconnect(), debug)
			c.Stop()
		case packet.PING:
			err = packet.WritePacket(c.conn, packet.NewPong(), debug)
		case packet.SUBSCRIBE:
			req := pkt.(*packet.SubscribePacket)

			var qos []packet.QoS
			for _, topic := range req.Topics {
				t := transport.EventTopic{Name: strings.Trim(topic.Topic, "/"), Qos: topic.QoS.Int()}
				qos = append(qos, c.addSubscription(t))
				c.engineChan <- transport.Event{MessageId: req.Id, ClientId: c.clientId, PacketType: pkt.Type(), Topic: t}
			}

			res := packet.NewSubAck()
			res.Id = req.Id
			res.ReturnCodes = qos
			err = packet.WritePacket(c.conn, res, debug)
		case packet.UNSUBSCRIBE:
			req := pkt.(*packet.UnSubscribePacket)
			res := packet.NewUnSubAck()

			res.Id = req.Id
			for _, topic := range req.Topics {
				t := transport.EventTopic{Name: strings.Trim(topic.Topic, "/"), Qos: topic.QoS.Int()}
				c.removeSubscription(t)
				err = packet.WritePacket(c.conn, res, debug)
			}
		case packet.PUBLISH:
			req := pkt.(*packet.PublishPacket)
			c.engineChan <- transport.Event{
				ClientId:   c.clientId,
				PacketType: pkt.Type(),
				MessageId:  req.Id,
				Topic:      transport.EventTopic{Name: strings.Trim(req.Topic, "/")},
				Payload:    req.Payload,
				Qos:        req.QoS.Int(),
				Retain:     req.Retain,
				Dublicate:  req.DUP,
			}
		case packet.PUBREC:
			req := pkt.(*packet.PubRecPacket)
			c.engineChan <- transport.Event{ClientId: c.clientId, PacketType: pkt.Type(), MessageId: req.Id}
		case packet.PUBREL:
			req := pkt.(*packet.PubRelPacket)
			c.engineChan <- transport.Event{ClientId: c.clientId, PacketType: pkt.Type(), MessageId: req.Id}
		case packet.PUBACK:
			req := pkt.(*packet.PubAckPacket)
			c.engineChan <- transport.Event{ClientId: c.clientId, PacketType: pkt.Type(), MessageId: req.Id}
		case packet.PUBCOMP:
			req := pkt.(*packet.PubCompPacket)
			c.engineChan <- transport.Event{ClientId: c.clientId, PacketType: pkt.Type(), MessageId: req.Id}
		default:
			err = packet.ErrUnknownPacket
		}

		if err != nil {
			log.Println("err serve, disconnected: ", err.Error())
			c.Stop()

			// send to engineChan unexpected disconnect
			c.engineChan <- transport.Event{ClientId: c.clientId}

			return
		}
	}
}

func (c *Client) Stop() {
	c.stop = true
	c.conn.Close()
}

func (c *Client) addSubscription(t transport.EventTopic) packet.QoS {
	// send 0x80 in qos in case of error

	// check for duplicate
	for _, v := range c.subscription {
		if v.Name == t.Name {
			if v.Qos != t.Qos {
				v.Qos = t.Qos
			}
			return packet.QoS(v.Qos)
		}
	}
	c.subscription = append(c.subscription, t)
	return packet.QoS(t.Qos)
}

func (c *Client) removeSubscription(t transport.EventTopic) bool {
	if len(c.subscription) == 0 {
		return false
	}
	for i, v := range c.subscription {
		if v.Name == t.Name {
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
	t := strings.Split(topic, "/")

	// searched topic is empty
	if len(t) == 0 {
		return false
	}

	for _, subs := range c.subscription {
		var i int = 0 // start from first level
		s := strings.Split(subs.Name, "/")

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
		will = fmt.Sprintf(", will: %s", c.will.String())
	}

	var subs string
	for _, v := range c.subscription {
		subs += v.String() + ", "
	}

	return fmt.Sprintf("client {clientId: %s, session: %v%s, subscription: [%s]}",
		c.clientId, c.session, will, subs)
}
