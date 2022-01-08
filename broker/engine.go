package broker

import (
	"log"
	"net"

	"github.com/MajaSuite/mqtt/packet"
	"github.com/MajaSuite/mqtt/transport"
)

var debug bool

type Engine struct {
	clients  map[string]*Client
	channel  chan transport.Event
	send     map[uint16]transport.Event
	delivery map[uint16]transport.Event
	retain   map[string]transport.Event
}

func NewEngine(_debug bool) *Engine {
	debug = _debug

	e := &Engine{
		clients:  make(map[string]*Client),
		channel:  make(chan transport.Event),
		send:     make(map[uint16]transport.Event),
		delivery: make(map[uint16]transport.Event),
		retain:   make(map[string]transport.Event),
	}

	return e
}

func (e *Engine) ManageServer(server *Server) {
	go e.clientThread()

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Println("err accept", err.Error())
			conn.Close()
			continue
		}

		go e.processConnect(conn)
	}
}

func (e *Engine) publishMessage(event transport.Event) {
	event.PacketType = packet.PUBLISH

	for _, client := range e.clients {
		if client != nil {
			if client.Contains(event.Topic.Name) {
				client.clientChan <- event

				if event.Qos > 0 {
					e.send[event.MessageId] = event
				}
			}
		}
	}
}

func (e *Engine) saveClient(id string, conn net.Conn, session bool) *Client {
	if e.clients[id] != nil {
		e.clients[id].conn = conn
	} else {
		client := NewClient(id, conn, session, e.channel)
		e.clients[id] = client
	}

	return e.clients[id]
}
