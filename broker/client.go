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
	debug        bool
	conn         net.Conn
	messageId    uint16
	toEngine     chan transport.Event
	toSendOut    chan transport.Event
	clientId     string
	session      bool                   // clean or persisted session
	subscription []transport.EventTopic // subscribed topics
	will         *packet.Message
	stop         bool // flag to stop
	started      bool
}

func NewClient(debug bool, id string, conn net.Conn, session bool, channel chan transport.Event) *Client {
	return &Client{
		debug:     debug,
		conn:      conn,
		messageId: 1,
		toEngine:  channel,
		toSendOut: make(chan transport.Event),
		clientId:  id,
		session:   session,
		started:   false,
	}
}

func (c *Client) Start() {
	if c.started {
		return
	}
	c.stop = false
	c.started = true

	var pkt packet.Packet
	var err error
	for {
		if c.stop {
			log.Printf("client %s disconnected\n", c.clientId)
			return
		}

		// send messages to client
		go func() {
			for event := range c.toSendOut {
				if c.debug {
					log.Println("message to send to client", event)
				}

				var res packet.Packet

				switch event.PacketType {
				case packet.PUBLISH:
					publish := packet.NewPublish()
					publish.Id = event.MessageId
					publish.Topic = event.Topic.Name
					publish.QoS = packet.QoS(event.Qos)
					publish.Payload = event.Payload
					publish.Retain = event.Retain
					publish.DUP = event.Dublicate
					res = publish
				case packet.PUBACK:
					ack := packet.NewPubAck()
					ack.Id = event.MessageId
					res = ack
				case packet.PUBREC:
					rec := packet.NewPubRec()
					rec.Id = event.MessageId
					res = rec
				case packet.PUBCOMP:
					comp := packet.NewPubComp()
					comp.Id = event.MessageId
					res = comp
				default:
					log.Println("wrong packet from engine")
					continue
				}

				if err := packet.WritePacket(c.conn, res, c.debug); err != nil {
					c.toEngine <- transport.Event{ClientId: c.clientId}
					log.Println("client disconnect while write to socket", err)
					c.Stop()
					return
				}
			}
		}()

		if pkt, err = packet.ReadPacket(c.conn, c.debug); err != nil {
			log.Println("err read packet, disconnected: ", err.Error())
			c.Stop()

			// send to toEngine unexpected disconnect
			c.toEngine <- transport.Event{ClientId: c.clientId}
			return
		}

		switch pkt.Type() {
		case packet.DISCONNECT:
			c.toEngine <- transport.Event{ClientId: c.clientId, PacketType: pkt.Type()}
			err = packet.WritePacket(c.conn, packet.NewDisconnect(), c.debug)
			c.Stop()
			return
		case packet.PING:
			err = packet.WritePacket(c.conn, packet.NewPong(), c.debug)
		case packet.SUBSCRIBE:
			var qos []packet.QoS
			req := pkt.(*packet.SubscribePacket)

			for _, topic := range req.Topics {
				t := transport.EventTopic{Name: strings.Trim(topic.Topic, "/"), Qos: topic.QoS.Int()}
				qos = append(qos, c.addSubscription(t))
				c.toEngine <- transport.Event{MessageId: req.Id, ClientId: c.clientId, PacketType: pkt.Type(), Topic: t}
			}

			res := packet.NewSubAck()
			res.Id = req.Id
			res.ReturnCodes = qos
			err = packet.WritePacket(c.conn, res, c.debug)
		case packet.UNSUBSCRIBE:
			req := pkt.(*packet.UnSubscribePacket)
			res := packet.NewUnSubAck()
			res.Id = req.Id

			for _, topic := range req.Topics {
				t := transport.EventTopic{Name: strings.Trim(topic.Topic, "/"), Qos: topic.QoS.Int()}
				c.removeSubscription(t)
				err = packet.WritePacket(c.conn, res, c.debug)
			}
		case packet.PUBLISH:
			req := pkt.(*packet.PublishPacket)
			c.toEngine <- transport.Event{
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
			c.toEngine <- transport.Event{ClientId: c.clientId, PacketType: pkt.Type(), MessageId: req.Id}
		case packet.PUBREL:
			req := pkt.(*packet.PubRelPacket)
			c.toEngine <- transport.Event{ClientId: c.clientId, PacketType: pkt.Type(), MessageId: req.Id}
		case packet.PUBACK:
			req := pkt.(*packet.PubAckPacket)
			c.toEngine <- transport.Event{ClientId: c.clientId, PacketType: pkt.Type(), MessageId: req.Id}
		case packet.PUBCOMP:
			req := pkt.(*packet.PubCompPacket)
			c.toEngine <- transport.Event{ClientId: c.clientId, PacketType: pkt.Type(), MessageId: req.Id}
		default:
			err = packet.ErrUnknownPacket
		}
		if err != nil {
			log.Println("unable to manage request, disconnect.", err)
			c.Stop()
			c.toEngine <- transport.Event{ClientId: c.clientId}
			return
		}
	}
}

func (c *Client) Stop() {
	c.stop = true
	c.conn.Close()
}

func (c *Client) addSubscription(t transport.EventTopic) packet.QoS {
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
		will = fmt.Sprintf(" will: %s,", c.will.String())
	}

	var subs string
	for _, v := range c.subscription {
		subs += v.String() + ", "
	}

	return fmt.Sprintf("client {clientId: %s, session: %v,%s subscription: [%s]}",
		c.clientId, c.session, will, subs)
}
