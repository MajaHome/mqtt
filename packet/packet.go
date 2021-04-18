package packet

import (
	"errors"
	"encoding/binary"
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

type Packet interface {
	Type() Type
	Length() int
	Unpack(buf []byte) error
	Pack() []byte
	ToString() string
}

func Types() []Type {
	return []Type{RESERVED, CONNECT, CONNACK,
		PUBLISH, PUBACK, PUBREC, PUBREL, PUBCOMP,
		SUBSCRIBE, SUBACK,
		UNSUBSCRIBE, UNSUBACK,
		PING, PONG,
		DISCONNECT}
}

// String returns the type as a string.
func (t Type) ToString() string {
	switch t {
	case CONNECT:
		return "CreateConnect"
	case CONNACK:
		return "CreateConnAck"
	case PUBLISH:
		return "CreatePublish"
	case PUBACK:
		return "CreatePubAck"
	case PUBREC:
		return "CreatePubRec"
	case PUBREL:
		return "CreatePubRel"
	case PUBCOMP:
		return "CreatePubComp"
	case SUBSCRIBE:
		return "CreateSubscribe"
	case SUBACK:
		return "CreateSubAck"
	case UNSUBSCRIBE:
		return "Unsubscribe"
	case UNSUBACK:
		return "UnsubscribeAck"
	case PING:
		return "CreatePing"
	case PONG:
		return "CreatePong"
	case DISCONNECT:
		return "Flag"
	}

	return "Unknown"
}

func Create(buf []byte) (Packet, uint8, error) {
	t := Type(buf[0] >> 4)
	len := buf[1]

	switch t {
	case CONNECT:
		return CreateConnect(buf), len, nil
	case CONNACK:
		return CreateConnAck(buf), len, nil
	case PUBLISH:
		return CreatePublish(buf), len, nil
	case PUBACK:
		return CreatePubAck(buf), len, nil
	case PUBREC:
		return CreatePubRec(buf), len, nil
	case PUBREL:
		return CreatePubRel(buf), len, nil
	case PUBCOMP:
		return CreatePubComp(buf), len, nil
	case SUBSCRIBE:
		return CreateSubscribe(buf), len, nil
	case SUBACK:
		return CreateSubAck(buf), len, nil
	case UNSUBSCRIBE:
		return CreateUnSubscribe(buf), len, nil
	case UNSUBACK:
		return CreateUnSubAck(buf), len, nil
	case PING:
		return CreatePing(buf), len, nil
	case PONG:
		return CreatePong(buf), len, nil
	case DISCONNECT:
		return CreateDisconnect(buf), len, nil
	}

	return nil, 0, ErrInvalidPacketType
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
		value = uint16(buf[offset+1]) | uint16(buf[offset])<<8
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

func WriteInt8(buf []byte, offset int, value uint8) int {
	buf[offset] = value
	offset++
	return offset
}

func WriteInt16(buf []byte, offset int, value uint16) int {
	binary.BigEndian.PutUint16(buf[offset:], value)
	offset += 2
	return offset
}
