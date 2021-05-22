package server

import (
	"log"
	"mqtt/packet"
	"net"
	"strconv"
	"strings"
)

type Client struct {
	conn         net.Conn
	engineChan   chan *Event
	clientChan   chan *Event
	clientId     string
	session      bool         // clean or persisten session
	subscription []EventTopic // subscribed topics
	will         *packet.Message
	stop         bool // flag to stop
}

func NewClient(id string, conn net.Conn, session bool, channel chan *Event) *Client {
	return &Client{
		conn:       conn,
		engineChan: channel,
		clientChan: make(chan *Event),
		clientId:   id,
		session:    session,
	}
}

func (c *Client) Start(server *Server) {
	var pkt packet.Packet
	var err error

	c.stop = false
	for {
		if c.stop {
			log.Println("stop " + c.clientId)
			return
		}

		// read from channel and/or from network
		go func() {
			for e := range c.clientChan {
				log.Println("client receive message: " + e.String())

				var res packet.Packet
				switch e.packetType {
				case packet.PUBLISH:
					res = packet.NewPublish()
					res.(*packet.PublishPacket).Id = e.messageId
					res.(*packet.PublishPacket).Topic = e.topic.name
					res.(*packet.PublishPacket).QoS = packet.QoS(e.topic.qos)
					res.(*packet.PublishPacket).Payload = e.payload
					res.(*packet.PublishPacket).Retain = e.retain
					res.(*packet.PublishPacket).DUP = e.dublicate
				case packet.PUBACK:
					res = packet.NewPubAck()
					res.(*packet.PubAckPacket).Id = e.messageId
				case packet.PUBREC:
					res = packet.NewPubRec()
					res.(*packet.PubRecPacket).Id = e.messageId
				case packet.PUBCOMP:
					res = packet.NewPubComp()
					res.(*packet.PubCompPacket).Id = e.messageId
				default:
					log.Println("wrong packet from engine")
				}

				if err := server.WritePacket(c.conn, res); err != nil {
					log.Println("client disconnect while write to socket")
					c.Stop()
					c.engineChan <- &Event{clientId: c.clientId} // send to engine unexpected disconnect
					return
				}
			}
		}()

		if pkt, err = server.ReadPacket(c.conn); err != nil {
			log.Println("err read packet, disconnected: ", err.Error())
			c.Stop()

			// send to engineChan unexpected disconnect
			event := &Event{clientId: c.clientId}
			c.engineChan <- event

			return
		}

		switch pkt.Type() {
		case packet.DISCONNECT:
			event := &Event{clientId: c.clientId, packetType: pkt.Type()}
			c.engineChan <- event
			res := packet.NewDisconnect()
			err = server.WritePacket(c.conn, res)
			c.Stop()
		case packet.PING:
			res := packet.NewPong()
			err = server.WritePacket(c.conn, res)
		case packet.SUBSCRIBE:
			req := pkt.(*packet.SubscribePacket)

			var qos []packet.QoS
			for _, topic := range req.Topics {
				t := EventTopic{name: strings.Trim(topic.Topic, "/"), qos: topic.QoS.Int()}
				qos = append(qos, c.addSubscription(t))

				event := &Event{clientId: c.clientId, packetType: pkt.Type(), topic: t}
				c.engineChan <- event
			}

			res := packet.NewSubAck()
			res.Id = req.Id
			res.ReturnCodes = qos
			err = server.WritePacket(c.conn, res)
		case packet.UNSUBSCRIBE:
			req := pkt.(*packet.UnSubscribePacket)
			res := packet.NewUnSubAck()

			res.Id = req.Id
			for _, topic := range req.Topics {
				t := EventTopic{name: strings.Trim(topic.Topic, "/"), qos: topic.QoS.Int()}
				c.removeSubscription(t)
				err = server.WritePacket(c.conn, res)
			}
		case packet.PUBLISH:
			req := pkt.(*packet.PublishPacket)
			event := &Event{
				clientId:   c.clientId,
				packetType: pkt.Type(),
				messageId:  req.Id,
				topic:      EventTopic{name: strings.Trim(req.Topic, "/")},
				payload:    req.Payload,
				qos:        req.QoS.Int(),
				retain:     req.Retain,
				dublicate:  req.DUP,
			}
			c.engineChan <- event
		case packet.PUBREC:
			req := pkt.(*packet.PubRecPacket)

			event := &Event{clientId: c.clientId, packetType: pkt.Type(), messageId: req.Id}
			c.engineChan <- event
		case packet.PUBREL:
			req := pkt.(*packet.PubRelPacket)

			event := &Event{clientId: c.clientId, packetType: pkt.Type(), messageId: req.Id}
			c.engineChan <- event
		case packet.PUBACK:
			req := pkt.(*packet.PubAckPacket)

			event := &Event{clientId: c.clientId, packetType: pkt.Type(), messageId: req.Id}
			c.engineChan <- event
		case packet.PUBCOMP:
			req := pkt.(*packet.PubCompPacket)

			event := &Event{clientId: c.clientId, packetType: pkt.Type(), messageId: req.Id}
			c.engineChan <- event
		default:
			err = packet.ErrUnknownPacket
		}

		if err != nil {
			log.Println("err serve, disconnected: ", err.Error())
			c.Stop()

			// send to engineChan unexpected disconnect
			event := &Event{clientId: c.clientId}
			c.engineChan <- event

			return
		}
	}
}

func (c *Client) Stop() {
	c.stop = true
	c.conn.Close()
}

func (c *Client) addSubscription(t EventTopic) packet.QoS {
	// 0x80 clientChan qos clientChan case of error ?
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

func (c *Client) Contains(topic string) bool {
	t := strings.Split(topic, "/")

	// searched topic is empty
	if len(t) == 0 {
		return false
	}

	for _, subs := range c.subscription {
		var i int = 0	// start from first level
		s := strings.Split(subs.name, "/")

		var found bool
		for {
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

			// doen't match
			break
		}

		if found {
			return true
		}
	}

	return false
}

func (c *Client) String() string {
	var sb strings.Builder
	sb.WriteString("client {")
	sb.WriteString("clientId: \"" + c.clientId + "\", ")
	sb.WriteString("session: " + strconv.FormatBool(c.session) + ", ")
	if c.will != nil {
		sb.WriteString("will: " + c.will.String() + ", ")
	}
	sb.WriteString("subscription: [")
	for _, v := range c.subscription {
		sb.WriteString("{ name: \"" + v.name + "\", qos: " + strconv.Itoa(int(v.qos)) + "}")
	}
	sb.WriteString("]")
	sb.WriteString("}")
	return sb.String()
}
