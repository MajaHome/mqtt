package broker

import (
	"log"
	"net"
)

type Server struct {
	listener net.Listener
	debug    bool
}

/* TODO manage ssl and websocket connections
 */
func NewListener(debug bool) *Server {
	l, err := net.Listen("tcp", "0.0.0.0:1883")
	if err != nil {
		return nil
	}

	log.Println("listen on address", l.Addr())
	return &Server{listener: l, debug: debug}
}

func (m *Server) Accept() (net.Conn, error) {
	conn, err := m.listener.Accept()
	if err != nil {
		return nil, err
	}

	if m.debug {
		log.Println("accept new connection from", conn.RemoteAddr())
	}
	return conn, nil
}

func (m *Server) Close() error {
	log.Println("close listener")
	return m.listener.Close()
}

func (m *Server) Addr() net.Addr {
	return m.listener.Addr()
}
