package broker

import (
	"log"
	"net"
)

type Server struct {
	listener net.Listener
	debug    bool
}

func NewServer(debug bool) *Server {
	l, err := net.Listen("tcp", "0.0.0.0:1883")
	if err != nil {
		return nil
	}

	log.Println("listen on address", l.Addr())

	return &Server{listener: l, debug: debug}
}

func (s *Server) Manage(broker *Broker) {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Println("error accept connection", err)
			conn.Close()
			continue
		}

		go broker.processConnect(conn)
	}
}

func (s *Server) Accept() (net.Conn, error) {
	conn, err := s.listener.Accept()
	if err != nil {
		return nil, err
	}

	if s.debug {
		log.Println("accept new connection from", conn.RemoteAddr())
	}

	return conn, nil
}

func (s *Server) Close() error {
	log.Println("close listener")
	return s.listener.Close()
}

func (s *Server) Addr() net.Addr {
	return s.listener.Addr()
}
