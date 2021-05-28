package packet

import (
	"errors"
)

const (
	RESERVED Type = iota
	CONNECT
	CONNACK
	PUBLISH
	PUBACK
	PUBREC
	PUBREL
	PUBCOMP
	SUBSCRIBE
	SUBACK
	UNSUBSCRIBE
	UNSUBACK
	PING
	PONG
	DISCONNECT
)

var (
	ErrInvalidPacketType = errors.New("invalid packet type")
	ErrProtocolError = errors.New("protocol error (not supported)")
	ErrInvalidPacketLength = errors.New("invalid packet Len")
	ErrUnknownPacket = errors.New("unknown packet type")
	ErrReadFromBuf = errors.New("error read data from buffer")
	ErrUnsupportedVersion = errors.New("unsupported mqtt version")
	ErrConnect = errors.New("error connect to broker")
)

type Packet interface {
	Type() Type
	Length() int
	Unpack(buf []byte) error
	Pack() []byte
	String() string
}

type Type byte

func (t Type) String() string {
	switch t {
	case CONNECT:
		return "Connect"
	case CONNACK:
		return "ConnAck"
	case PUBLISH:
		return "Publish"
	case PUBACK:
		return "PubAck"
	case PUBREC:
		return "PubRec"
	case PUBREL:
		return "PubRel"
	case PUBCOMP:
		return "PubComp"
	case SUBSCRIBE:
		return "Subscribe"
	case SUBACK:
		return "SubAck"
	case UNSUBSCRIBE:
		return "Unsubscribe"
	case UNSUBACK:
		return "UnsubAck"
	case PING:
		return "Ping"
	case PONG:
		return "Pong"
	case DISCONNECT:
		return "Disconnect"
	}

	return "Unknown"
}

func Create(buf []byte) Packet {
	t := Type(buf[0] >> 4)

	switch t {
	case CONNECT:
		return CreateConnect(buf)
	case CONNACK:
		return CreateConnAck(buf)
	case PUBLISH:
		return CreatePublish(buf)
	case PUBACK:
		return CreatePubAck(buf)
	case PUBREC:
		return CreatePubRec(buf)
	case PUBREL:
		return CreatePubRel(buf)
	case PUBCOMP:
		return CreatePubComp(buf)
	case SUBSCRIBE:
		return CreateSubscribe(buf)
	case SUBACK:
		return CreateSubAck(buf)
	case UNSUBSCRIBE:
		return CreateUnSubscribe(buf)
	case UNSUBACK:
		return CreateUnSubAck(buf)
	case PING:
		return CreatePing(buf)
	case PONG:
		return CreatePong(buf)
	case DISCONNECT:
		return CreateDisconnect(buf)
	}

	return nil
}
