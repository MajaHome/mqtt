package server

import (
	"encoding/hex"
	"fmt"
	"io"
	"mqtt/packet"
	"net"
	"time"
)

type Engine struct {
	backend *Backend
}

func GetEngine(backend *Backend) *Engine {
	return &Engine{
		backend: backend,
	}
}

func (e Engine) Manage(server *Server) {
	fmt.Println("Listen on address", server.Addr())

	for {
		// accept next connection
		conn, err := server.Accept()
		if err != nil {
			fmt.Println("err accept", err.Error())
			continue
		}
		go e.Accept(conn)

	}
}

func (e Engine) Accept(conn net.Conn) error {
	conn.SetDeadline(time.Now().Add(time.Minute*5))

	for {
		err := e.Serve(conn)
		if err != nil {
			if err != io.EOF {
				fmt.Println("err serve: ", err.Error())
			}
			conn.Close()
			return err
		}
	}
}

func (e Engine) Close() {
	e.backend.Close()
}

func (e Engine) Serve(conn net.Conn) error {
	fHeader := make([]byte, 2)
	n, err := conn.Read(fHeader)
	if n < 2 || err != nil {
		return io.ErrUnexpectedEOF
	}

	fmt.Println("request")
	fmt.Println(hex.Dump(fHeader))

	pkt, packetLength, err := packet.Create(fHeader)

	if packetLength != 0 {
		vHeader := make([]byte, packetLength)
		n, err = conn.Read(vHeader)
		if n < int(packetLength) || err != nil {
			return io.ErrUnexpectedEOF
		}

		fmt.Println(hex.Dump(vHeader))
		pkt.Unpack(vHeader)
	}

	fmt.Println("packet", pkt.ToString(), "\n")

	var res packet.Packet
	switch pkt.Type() {
	case packet.CONNECT:
		res = packet.NewConnAck()
		res.(*packet.ConnAckPacket).Session = !pkt.(*packet.ConnPacket).CleanSession
		res.(*packet.ConnAckPacket).ReturnCode = uint8(packet.ConnectAccepted)
	case packet.DISCONNECT:
		return io.EOF
	case packet.SUBSCRIBE:
		res = packet.NewSubAck()

		// send SUBACK
		res.(*packet.SubAckPacket).Id = pkt.(*packet.SubscribePacket).Id
		var qos []packet.QoS
		for _, q := range pkt.(*packet.SubscribePacket).Topics {
			qos = append(qos, q.QoS)
		}
		res.(*packet.SubAckPacket).ReturnCodes = qos
	case packet.UNSUBSCRIBE:
		res = packet.NewUnSubAck()
	case packet.PUBLISH:
		res = pkt
		
		// if payload is empty - unsubscribe to the topic
	case packet.PUBACK:
		res = pkt
	case packet.PUBCOMP:
		;
	case packet.PUBREC:
		;
	case packet.PUBREL:
		;
	case packet.PING:
		res = packet.NewPong()
	default:
		return packet.ErrUnknownPacket
	}

	r := res.Pack()

	fmt.Println("response")
	fmt.Println(hex.Dump(r))

	n, err = conn.Write(r)
	if err != nil {
		fmt.Println("err write")
	}

	return nil
}
