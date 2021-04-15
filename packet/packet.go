package packet

import (
	"errors"
)

var ErrInvalidPacketType = errors.New("invalid packet type")
var ErrProtocolError = errors.New("protocol error (not supported)")
var ErrInvalidPacketLength = errors.New("invalid packet Len")
var ErrUnknownPacket = errors.New("unknown packet type")
var ErrReadFromBuf = errors.New("error read data from buffer")
var ErrUnsupportedVersion = errors.New("unsupported mqtt version")

type Type byte

const (
	RESERVED Type = iota
	CONNECT
	CONNECTACK
	PUBLISH
	PUBLISHACK
	PUBLISHREC
	PUBLISHREL
	PUBLISHCOMP
	SUBSCRIBE
	SUBSCRIBEACK
	UNSUBSCRIBE
	UNSUBSCRIBEACK
	PING
	PONG
	DISCONNECT
)

type Packet interface {
	Type() Type
	Length() int
	Unpack(buf []byte) error
	Pack() ([]byte, error)
	ToString() string
}

func Types() []Type {
	return []Type{RESERVED, CONNECT, CONNECTACK,
		PUBLISH, PUBLISHACK, PUBLISHREC, PUBLISHREL, PUBLISHCOMP,
		SUBSCRIBE, SUBSCRIBEACK, UNSUBSCRIBE, UNSUBSCRIBEACK,
		PING, PONG, DISCONNECT}
}

// String returns the type as a string.
func (t Type) ToString() string {
	switch t {
	case CONNECT:
		return "Connect"
	case CONNECTACK:
		return "ConnectAck"
	case PUBLISH:
		return "Publish"
	case PUBLISHACK:
		return "PublishAck"
	case PUBLISHREC:
		return "PublishRec"
	case PUBLISHREL:
		return "PublishRel"
	case PUBLISHCOMP:
		return "PublishComp"
	case SUBSCRIBE:
		return "Subscribe"
	case SUBSCRIBEACK:
		return "SubscribeAck"
	case UNSUBSCRIBE:
		return "Unsubscribe"
	case UNSUBSCRIBEACK:
		return "UnsubscribeAck"
	case PING:
		return "Ping"
	case PONG:
		return "Pong"
	case DISCONNECT:
		return "Disconnect"
	}

	return "Unknown"
}

func (t Type) Create() (Packet, error) {
	switch t {
	case CONNECT:
		return Connect(), nil
	case CONNECTACK:
		return ConnectAck(), nil
	case PUBLISH:
		return Publish(), nil
	case PUBLISHACK:
		return PublishAck(), nil
	case PUBLISHREC:
		return PublishRec(), nil
	case PUBLISHREL:
		return PublishRel(), nil
	case PUBLISHCOMP:
		return PublishComp(), nil
	case SUBSCRIBE:
		return Subscribe(), nil
	case SUBSCRIBEACK:
		return SubscribeAck(), nil
	case UNSUBSCRIBE:
		return UnSubscribe(), nil
	case UNSUBSCRIBEACK:
		return UnSubscribeAck(), nil
	case PING:
		return Ping(), nil
	case PONG:
		return Pong(), nil
	case DISCONNECT:
		return Disconnect(), nil
	}

	return nil, ErrInvalidPacketType
}

func ReadInt8(buf []byte, offset int) (uint8, int, error) {
	var value uint8

	if len(buf) >= (offset+1) {
		value = buf[offset]
		return value, offset + 1, nil
	}

	return 0, offset, ErrReadFromBuf
}

func ReadInt16(buf []byte, offset int) (uint16, int, error) {
	var value uint16

	if len(buf) >= (offset+2) {
		value = uint16(buf[offset] << 4 + buf[offset+1])
		return value, offset + 2, nil
	}

	return 0, offset, ErrReadFromBuf
}

func ReadBytes(buf []byte, offset int, length int) ([]byte, int, error) {
	var value []byte

	if len(buf) >= (offset+length) {
		value = buf[offset:offset+length]
		return value, offset + length, nil
	}

	return nil, offset, ErrReadFromBuf
}

func ReadString(buf []byte, offset int, length int) (string, int, error) {
	s, i, e := ReadBytes(buf, offset, length)
	return string(s), i, e
}