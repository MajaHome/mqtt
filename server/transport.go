package server

import (
	"crypto/tls"
	"fmt"
	"net"
)

type Server struct {
	listener net.Listener
}

func NewMqtt(l net.Listener) *Server {
	return &Server{
		listener: l,
	}
}

func Run() (*Server, error) {
	l, err := net.Listen("tcp", "0.0.0.0:1883")
	if err != nil {
		return nil, err
	}

	return NewMqtt(l), nil
}

func RunSecured(config *tls.Config) (*Server, error) {
	l, err := tls.Listen("tcp", "0.0.0.0:8883", config)
	if err != nil {
		return nil, err
	}

	return NewMqtt(l), nil
}

func RunWs() (*Server, error) {
	return nil, nil
}

func RunSecuredWs() (*Server, error) {
	return nil, nil
}

// Accept will return the next available connection or block until a connection becomes available.
func (m *Server) Accept() (net.Conn, error) {
	// wait next connection
	conn, err := m.listener.Accept()
	if err != nil {
		return nil, err
	}

	fmt.Println("Accept new connection from", conn.RemoteAddr())
	return conn, nil
}

// Close will close the underlying listener and cleanup resources.
func (m *Server) Close() error {
	return m.listener.Close()
}

// Addr returns the server's network address.
func (m *Server) Addr() net.Addr {
	return m.listener.Addr()
}
