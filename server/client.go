package server

import (
	"log"
	"mqtt/packet"
	"net"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	conn         net.Conn
	engine       chan Event
	channel      chan Event
	id           string
	session      bool
	subscription []EventTopic // subscribed topics
	willQoS      packet.QoS
	willRetain   bool
}

func NewClient(id string, conn net.Conn, session bool, engine chan Event) *Client {
	channel := make(chan Event)

	return &Client{
		conn:    conn,
		engine:  engine,
		channel: channel,
		id:      id,
		session: session,
	}
}

func (c *Client) Start(server *Server) {
	for {
		c.conn.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
		pkt, _ := server.ReadPacket(c.conn)

		select {
		case e := <-c.channel:
			log.Println("client: " + e.String())

			res := packet.NewPublish()
			res.Id = e.messageId
			res.Topic = e.topic.name
			res.QoS = packet.QoS(e.topic.qos)
			res.Payload = e.payload
			res.Retain = e.retain
			res.DUP = e.dublicate

			if err := server.WritePacket(c.conn, res); err != nil {
				log.Println("client disconnect while write from engine")
				c.conn.Close()

				// send to engine unexpected disconnect
				event := Event{clientId: c.id}
				c.engine <- event

				return
			}
			continue
		default:
			break
		}

		if pkt == nil {
			continue
		}

		var err error
		switch pkt.Type() {
		case packet.DISCONNECT:
			res := packet.NewDisconnect()

			event := &Event{
				clientId:   c.id,
				packetType: pkt.Type(),
			}
			c.engine <- *event

			err = server.WritePacket(c.conn, res)
		case packet.PING:
			res := packet.NewPong()

			err = server.WritePacket(c.conn, res)
		case packet.SUBSCRIBE:
			req := pkt.(*packet.SubscribePacket)
			res := packet.NewSubAck()

			res.Id = req.Id
			var qos []packet.QoS
			for _, topic := range req.Topics {
				t := EventTopic{name: strings.Trim(topic.Topic, "/"), qos: topic.QoS.Int()}
				event := Event{
					clientId:   c.id,
					packetType: pkt.Type(),
					messageId:  req.Id,
					topic:      t,
				}
				c.engine <- event

				tqos := c.addSubscribtion(t)
				qos = append(qos, tqos)
			}
			res.ReturnCodes = qos

			err = server.WritePacket(c.conn, res)
		case packet.UNSUBSCRIBE:
			req := pkt.(*packet.UnSubscribePacket)
			res := packet.NewUnSubAck()

			res.Id = req.Id
			for _, topic := range req.Topics {
				t := EventTopic{name: strings.Trim(topic.Topic, "/"), qos: topic.QoS.Int()}
				event := &Event{
					clientId:   c.id,
					packetType: pkt.Type(),
					messageId:  req.Id,
					topic:      t,
				}
				c.engine <- *event

				c.removeSubscription(t)
				err = server.WritePacket(c.conn, res)
			}
		case packet.PUBLISH:
			req := pkt.(*packet.PublishPacket)

			// do not publish on "special" topics
			if !strings.HasPrefix(req.Topic, "$") {
				event := &Event{
					clientId:   c.id,
					packetType: pkt.Type(),
					messageId:  req.Id,
					topic:      EventTopic{name: strings.Trim(req.Topic, "/"), qos: req.QoS.Int()},
					payload:    req.Payload,
					retain:     req.Retain,
					dublicate:  req.DUP,
				}
				c.engine <- *event
			}

			switch req.QoS {
			case packet.AtLeastOnce:
				res := packet.NewPubAck()
				res.Id = req.Id
				err = server.WritePacket(c.conn, res)
			case packet.ExactlyOnce:
				res := packet.NewPubRec()
				res.Id = req.Id
				err = server.WritePacket(c.conn, res)
			}
		case packet.PUBREC:
			req := pkt.(*packet.PubRecPacket)

			event := &Event{
				clientId:   c.id,
				packetType: pkt.Type(),
				messageId:  req.Id,
			}
			c.engine <- *event

			// TODO read response from engine - if packet with id is registered
			res := packet.NewPubRel()
			res.Id = req.Id

			err = server.WritePacket(c.conn, res)
		case packet.PUBREL:
			req := pkt.(*packet.PubRelPacket)

			event := &Event{
				clientId:   c.id,
				packetType: pkt.Type(),
				messageId:  req.Id,
			}
			c.engine <- *event

			// TODO read response from engine - if packet with id is registered
			res := packet.NewPubComp()
			res.Id = req.Id

			err = server.WritePacket(c.conn, res)
		case packet.PUBACK:
			req := pkt.(*packet.PubAckPacket)

			event := &Event{
				clientId:   c.id,
				packetType: pkt.Type(),
				messageId:  req.Id,
			}
			c.engine <- *event
		case packet.PUBCOMP:
			req := pkt.(*packet.PubCompPacket)

			event := &Event{
				clientId:   c.id,
				packetType: pkt.Type(),
				messageId:  req.Id,
			}
			c.engine <- *event
		default:
			err = packet.ErrUnknownPacket
		}

		if err != nil {
			log.Println("err serve, disconnected: ", err.Error())
			c.conn.Close()

			// send to engine unexpected disconnect
			event := Event{clientId: c.id}
			c.engine <- event

			return
		}
	}
}

func (c *Client) saveWill(qos packet.QoS, retain bool) {
	c.willQoS = qos
	c.willRetain = retain
}

func (c *Client) addSubscribtion(t EventTopic) packet.QoS {
	// check for duplicate
	for _, v := range c.subscription {
		if v.name == t.name {
			if v.qos != t.qos {
				v.qos = t.qos
			}
			return packet.QoS(v.qos)
		}
	}
	c.subscription = append(c.subscription, t)
	return packet.QoS(t.qos)
}

func (c *Client) removeSubscription(t EventTopic) bool {
	if len(c.subscription) == 0 {
		return false
	}
	if len(c.subscription) == 1 {
		c.subscription = []EventTopic{}
		return true
	}
	for i, v := range c.subscription {
		if v.name == t.name {
			if len(c.subscription) > i+1 {
				c.subscription[i] = c.subscription[len(c.subscription)-1]
			}
			c.subscription = c.subscription[:len(c.subscription)-1]
			return true
		}
	}
	return false
}

func (c *Client) isSubscribed(topic string) bool {
	for _, v := range c.subscription {
		if v.name == topic {
			return true
		}
	}
	return true
}

func (c *Client) String() string {
	var sb strings.Builder
	sb.WriteString("client {")
	sb.WriteString("id: \"" + c.id + "\", ")
	sb.WriteString("session: " + strconv.FormatBool(c.session) + ", ")
	sb.WriteString("willQoS: " + strconv.Itoa(int(c.willQoS)) + ", ")
	sb.WriteString("willRetain: " + strconv.FormatBool(c.willRetain) + ", ")
	sb.WriteString("subscription: [")
	for _, v := range c.subscription {
		sb.WriteString("{ name: \"" + v.name + "\", qos: " + strconv.Itoa(int(v.qos)) + "}")
	}
	sb.WriteString("]")
	sb.WriteString("}")
	return sb.String()
}
