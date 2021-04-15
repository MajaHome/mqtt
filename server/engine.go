package server

import (
	"fmt"
	"io"
	"net"
	"time"
	"encoding/hex"
	"encoding/binary"
	"mqtt/packet"
)

type Engine struct {
	backend *Backend
}

func GetEngine(backend *Backend) *Engine {
	return &Engine{
		backend: backend,
	}
}

func (e Engine) Accept(server *Server) error {
	for {
		// accept next connection
		conn, err := server.Accept()
		if err != nil {
			fmt.Println(err.Error())
			return err
		}

		// handle connection
		err = e.Serve(conn)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
	}
}

func (e Engine) Close() {
	e.backend.Close()
}

func (e Engine) Serve(conn net.Conn) error {
	// read command header
	conn.SetDeadline(time.Now().Add(time.Second*2))

	recvBuf1 := make([]byte, 2)
	n, err := conn.Read(recvBuf1)
	if n < 2 || err != nil {
		return io.ErrUnexpectedEOF
	}

	packetType := packet.Type(recvBuf1[0] >> 4)
	packetLength, _ := binary.Uvarint(recvBuf1[1:])
	if packetLength == 0 {
		return packet.ErrInvalidPacketLength
	}

	recvBuf2 := make([]byte, packetLength+2)
	copy(recvBuf2, recvBuf1)
	n, err = conn.Read(recvBuf2[2:])
	if n < int(packetLength) || err != nil {
		return io.ErrUnexpectedEOF
	}
	fmt.Println("dump", hex.Dump(recvBuf2))

	pkt, _ := packetType.Create()
	pkt.Unpack(recvBuf2)

	fmt.Println("packet", pkt.ToString())

	var res packet.Packet
	switch pkt.Type() {
	case packet.CONNECT:
		res = packet.ConnectAck()
		res.(*packet.ConnectAckPacket).Session = true
		res.(*packet.ConnectAckPacket).ReturnCode = uint8(packet.ConnectAccepted)
	case packet.DISCONNECT:
		res = packet.DisconnectAck()
	case packet.SUBSCRIBE:
		res = packet.SubscribeAck()
	case packet.UNSUBSCRIBE:
		res = packet.UnSubscribeAck()
	case packet.PUBLISH:
		res = packet.PublishAck()
	case packet.PUBLISHCOMP:
		res = packet.PublishAck()
	case packet.PUBLISHREC:
		res = packet.PublishAck()
	case packet.PUBLISHREL:
		res = packet.PublishAck()
	case packet.PING:
		res = packet.Pong()
	default:
		return packet.ErrUnknownPacket
	}

	r, err := res.Pack()
	if err == nil {
		fmt.Println("response", hex.Dump(r))
		conn.Write(r)
	}

	return nil
}
