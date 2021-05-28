package broker

import (
	"log"
	"net"
)

type Server struct {
	listener net.Listener
}

func RunMqtt() (*Server, error) {
	l, err := net.Listen("tcp", "0.0.0.0:1883")
	if err != nil {
		return nil, err
	}

	log.Println("Listen on address", l.Addr())

	return &Server{listener: l}, nil
}

func (m *Server) Accept() (net.Conn, error) {
	conn, err := m.listener.Accept()
	if err != nil {
		return nil, err
	}

	log.Println("Accept new connection from", conn.RemoteAddr())
	return conn, nil
}

func (m *Server) Close() error {
	return m.listener.Close()
}

func (m *Server) Addr() net.Addr {
	return m.listener.Addr()
}
