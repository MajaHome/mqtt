package server

import (
	"encoding/binary"
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
	recvBuf1 := make([]byte, 2)
	n, err := conn.Read(recvBuf1)
	if n < 2 || err != nil {
		return io.ErrUnexpectedEOF
	}

	packetType := packet.Type(recvBuf1[0] >> 4)
	pkt, _ := packetType.Create()

	packetLength, _ := binary.Uvarint(recvBuf1[1:])
	if packetLength != 0 {
		recvBuf2 := make([]byte, packetLength+2)
		copy(recvBuf2, recvBuf1)
		n, err = conn.Read(recvBuf2[2:])
		if n < int(packetLength) || err != nil {
			return io.ErrUnexpectedEOF
		}

		fmt.Println(hex.Dump(recvBuf2))
		pkt.Unpack(recvBuf2)
	} else {
		fmt.Println(hex.Dump(recvBuf1))
	}

	fmt.Println("packet", pkt.ToString())

	var res packet.Packet
	switch pkt.Type() {
	case packet.CONNECT:
		res = packet.ConnectAck()
		/*
		If the Server accepts a connection with CleanSession set to 1, the Server MUST set Session Present to 0
		in the CONNACK packet in addition to setting a zero return code in the CONNACK packet [MQTT-3.2.2-1].
		If the Server accepts a connection with CleanSession set to 0, the value set in Session Present depends
		on whether the Server already has stored Session state for the supplied client ID. If the Server has
		stored Session state, it MUST set Session Present to 1 in the CONNACK packet [MQTT-3.2.2-2]. If the
		Server does not have stored Session state, it MUST set Session Present to 0 in the CONNACK packet.
		This is in addition to setting a zero return code in the CONNACK packet [MQTT-3.2.2-3].
		*/
		res.(*packet.ConnectAckPacket).Session = !pkt.(*packet.ConnectPacket).CleanSession
		res.(*packet.ConnectAckPacket).ReturnCode = uint8(packet.ConnectAccepted)
	case packet.DISCONNECT:
		return io.EOF
	case packet.SUBSCRIBE:
		res = packet.SubscribeAck()

		// send SUBACK
		res.(*packet.SubscribeAckPacket).Id = pkt.(*packet.SubscribePacket).Id
		var qos []packet.QoS
		for _, q := range pkt.(*packet.SubscribePacket).Topics {
			qos = append(qos, q.QoS)
		}
		res.(*packet.SubscribeAckPacket).ReturnCodes = qos
	case packet.UNSUBSCRIBE:
		res = packet.UnSubscribeAck()
	case packet.PUBLISH:
		res = packet.PublishAck()
	case packet.PUBCOMP:
		res = packet.PublishAck()
	case packet.PUBREC:
		res = packet.PublishAck()
	case packet.PUBREL:
		res = packet.PublishAck()
	case packet.PING:
		res = packet.Pong()
	default:
		return packet.ErrUnknownPacket
	}

	r, err := res.Pack()
	if err != nil {
		return err
	}
	fmt.Println("response", hex.Dump(r))
	n, err = conn.Write(r)
	if err != nil {
		fmt.Println("err write")
	}

	return nil
}
