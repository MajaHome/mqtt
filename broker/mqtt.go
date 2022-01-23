package broker

import (
	"github.com/MajaSuite/mqtt/packet"
	"github.com/MajaSuite/mqtt/transport"
	"log"
)

type Mqtt struct {
	debug         bool
	brokerChannel chan transport.Event // channel to mqtt broker to manage
	clients       map[string]*Client
	sent          map[uint16]transport.Event
	sentWithQos   map[uint16]transport.Event
	retain        map[string]transport.Event
}

func NewMqtt(debug bool, listener *Server) *Mqtt {
	m := &Mqtt{
		debug:         debug,
		clients:       make(map[string]*Client),
		brokerChannel: make(chan transport.Event),
		sent:          make(map[uint16]transport.Event),
		sentWithQos:   make(map[uint16]transport.Event),
		retain:        make(map[string]transport.Event),
	}

	go m.broker()

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

	return m
}

func (e *Mqtt) publishMessage(event transport.Event) {
	event.PacketType = packet.PUBLISH

	if event.ClientId == "" {
		// send to all subscribed clients
		for _, client := range e.clients {
			if client != nil && client.conn != nil {
				if client.Contains(event.Topic.Name) {
					if event.Qos > 0 {
						// todo incorrect key
						e.sent[event.MessageId] = event
					}
					client.toSendOut <- event
				}
			}
		}
	} else {
		// send to only one client
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

// TODO once per minute rescan e.sent and try to send again (in pinger may be?)
