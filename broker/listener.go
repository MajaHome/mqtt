package broker

import (
	"log"
	"net"
)

type listener struct {
	debug    bool
	listener net.Listener
	broker   *Broker
}

func NewListener(debug bool, key string, cert string) *listener {
	l, err := net.Listen("tcp", "0.0.0.0:1883")
	if err != nil {
		return nil
	}

	log.Println("listen on address", l.Addr())

	return &listener{debug: debug, listener: l, broker: NewBroker(debug)}
}

func (s *listener) Manage() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Println("error accept connection", err)
			conn.Close()
			continue
		}

		go s.broker.newConnection(conn)
	}
}

func (s *listener) Accept() (net.Conn, error) {
	conn, err := s.listener.Accept()
	if err != nil {
		return nil, err
	}

	if s.debug {
		log.Println("accept new connection from", conn.RemoteAddr())
	}

	return conn, nil
}

func (s *listener) Close() error {
	log.Println("close listener")
	return s.listener.Close()
}

func (s *listener) Addr() net.Addr {
	return s.listener.Addr()
}
