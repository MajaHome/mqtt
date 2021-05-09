package server

import (
	"log"
	"mqtt/packet"
	"net"
	"time"
)

type Client struct {
	conn    net.Conn
	engine  chan Event
	channel chan Event

	id      string
	session bool
	topics  []string // subscribed topics

	willQoS    packet.QoS
	willRerail bool
}

func NewClient(id string, conn net.Conn, session bool, engine chan Event) *Client {
	channel := make(chan Event)

	return &Client{
		conn:    conn,
		engine:  engine,
		channel: channel,
		id:      id,
		session:  session,
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
			res.Topic = e.topic
			res.Payload = e.payload
			res.QoS = packet.QoS(e.qos)
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
				event := Event{
					clientId:   c.id,
					packetType: pkt.Type(),
					messageId:  req.Id,
					topic:      topic.Topic,
					qos:        topic.QoS.Int(),
				}
				c.engine <- event
				// actually it must receive answer (qos) from engine in case of success subscribtion, otherwise send 0x80 as qos
				qos = append(qos, topic.QoS)
			}
			res.ReturnCodes = qos

			err = server.WritePacket(c.conn, res)
		case packet.UNSUBSCRIBE:
			req := pkt.(*packet.UnSubscribePacket)
			res := packet.NewUnSubAck()

			res.Id = req.Id
			for _, t := range req.Topics {
				event := &Event{
					clientId:   c.id,
					packetType: pkt.Type(),
					messageId:  req.Id,
					topic:      t.Topic,
					qos:        t.QoS.Int(),
				}
				c.engine <- *event

				err = server.WritePacket(c.conn, res)
			}
		case packet.PUBLISH:
			req := pkt.(*packet.PublishPacket)

			event := &Event{
				clientId:   c.id,
				packetType: pkt.Type(),
				messageId:  req.Id,
				topic:      req.Topic,
				payload:    req.Payload,
				qos:        req.QoS.Int(),
				retain:     req.Retain,
				dublicate:  req.DUP,
			}
			c.engine <- *event

			switch req.QoS {
			case packet.AtLeastOnce:
				res := packet.NewPubAck()
				res.Id = req.Id
				err = server.WritePacket(c.conn, res)
			case packet.ExactlyOnce:
				res := packet.NewPubRec()
				res.Id = req.Id
				// read response from engine - if packet with id is registered
				err = server.WritePacket(c.conn, res)
			}
		case packet.PUBREC:
			req := pkt.(*packet.PubRecPacket)
			res := packet.NewPubRel()
			res.Id = req.Id
			// read response from engine - if packet with id is registered

			event := &Event{
				clientId:   c.id,
				packetType: pkt.Type(),
				messageId:  req.Id,
			}
			c.engine <- *event

			err = server.WritePacket(c.conn, res)
		case packet.PUBREL:
			req := pkt.(*packet.PubRelPacket)
			res := packet.NewPubComp()
			res.Id = req.Id
			// read response from engine - if packet with id is registered

			event := &Event{
				clientId:   c.id,
				packetType: pkt.Type(),
				messageId:  req.Id,
			}
			c.engine <- *event

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
	c.willRerail = retain
}
