package broker

import (
	"fmt"
	"github.com/MajaSuite/mqtt/packet"
	"github.com/MajaSuite/mqtt/transport"
	"log"
	"time"
)

type Mqtt struct {
	debug         bool
	brokerChannel chan transport.Event // channel to mqtt broker to manage
	clients       map[string]*Client
	sent          map[string]transport.Event
	sentWithQos   map[string]transport.Event
	retain        map[string]transport.Event
}

func NewMqtt(debug bool, listener *Server) *Mqtt {
	m := &Mqtt{
		debug:         debug,
		brokerChannel: make(chan transport.Event),
		clients:       make(map[string]*Client),
		sent:          make(map[string]transport.Event),
		sentWithQos:   make(map[string]transport.Event),
		retain:        make(map[string]transport.Event),
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Println("error accept connection", err)
				conn.Close()
				continue
			}

			go m.processConnect(conn)
		}
	}()
	go m.broker()
	go m.rescan()

	return m
}

func (e *Mqtt) publishMessage(event transport.Event) {
	event.PacketType = packet.PUBLISH

	if event.ClientId == "" {
		// send to all subscribed clients
		for _, client := range e.clients {
			if client != nil && client.conn != nil {
				if client.Contains(event.Topic.Name) {
					event.MessageId = client.messageId
					client.messageId++
					if event.Qos > 0 {
						e.sent[fmt.Sprintf("%s/%d", event.ClientId, event.MessageId)] = event
					}
					client.toSendOut <- event
				}
			}
		}
	} else {
		// send to only one client
		event.MessageId = e.clients[event.ClientId].messageId
		e.clients[event.ClientId].messageId++
		if event.Qos > 0 {
			e.sent[fmt.Sprintf("%s/%d", event.ClientId, event.MessageId)] = event
		}
		e.clients[event.ClientId].toSendOut <- event
	}
}

func (e *Mqtt) sendWill(client *Client) {
	if client != nil {
		if client.will != nil {
			will := transport.Event{
				Topic:   transport.EventTopic{Name: client.will.Topic},
				Payload: client.will.Payload,
				Qos:     client.will.QoS.Int(),
				Retain:  client.will.Retain,
			}

			for _, c := range e.clients {
				if c != nil && c.conn != nil {
					client.toSendOut <- will
				}
			}
		}

		// delete clean session
		if !client.session {
			delete(e.clients, client.clientId)
		}
	}
}

func (e *Mqtt) rescan() {
	for {
		if e.debug {
			log.Println("rescan queue for ack")
		}

		//for id, event := range e.sent {
		//	// TODO try to send messages without ack again
		//}

		time.Sleep(time.Second * 10)
	}
}
