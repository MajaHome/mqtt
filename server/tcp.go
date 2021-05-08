package server

import (
	"encoding/hex"
	"io"
	"log"
	"mqtt/packet"
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

func (m *Server) ReadPacket(conn net.Conn) (packet.Packet, error) {
	header := make([]byte, 2)
	n, err := conn.Read(header)
	if n < 2 || err != nil {
		return nil, io.ErrUnexpectedEOF
	}

	log.Println("packet header:\n", hex.Dump(header))

	pkt, err := packet.Create(header)
	if err != nil {
		log.Println("error create packet")
		return nil, packet.ErrUnknownPacket
	}

	packetLength := header[1]
	if packetLength != 0 {
		payload := make([]byte, packetLength)
		n, err = conn.Read(payload)
		if n < int(packetLength) || err != nil {
			return nil, io.ErrUnexpectedEOF
		}

		log.Println("packet payload:\n", hex.Dump(payload))

		pkt.Unpack(payload)
	}

	log.Println("packet:", pkt.String())

	return pkt, nil
}

func (m *Server) WritePacket(conn net.Conn, pkt packet.Packet) error {
	packed := pkt.Pack()

	log.Println("response:\n", hex.Dump(packed))

	_, err := conn.Write(packed)
	if err != nil {
		log.Println("err write response")
	}

	return err
}

func (m *Server) Close() error {
	return m.listener.Close()
}

func (m *Server) Addr() net.Addr {
	return m.listener.Addr()
}
