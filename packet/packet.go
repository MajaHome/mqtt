package packet

import (
	"encoding/binary"
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
	String() string
}

func Types() []Type {
	return []Type{RESERVED, CONNECT, CONNACK, PUBLISH, PUBACK, PUBREC, PUBREL, PUBCOMP, SUBSCRIBE, SUBACK,
		UNSUBSCRIBE, UNSUBACK, PING, PONG, DISCONNECT}
}

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

func Create(buf []byte) (Packet, error) {
	t := Type(buf[0] >> 4)

	switch t {
	case CONNECT:
		return CreateConnect(buf), nil
	case CONNACK:
		return CreateConnAck(buf), nil
	case PUBLISH:
		return CreatePublish(buf), nil
	case PUBACK:
		return CreatePubAck(buf), nil
	case PUBREC:
		return CreatePubRec(buf), nil
	case PUBREL:
		return CreatePubRel(buf), nil
	case PUBCOMP:
		return CreatePubComp(buf), nil
	case SUBSCRIBE:
		return CreateSubscribe(buf), nil
	case SUBACK:
		return CreateSubAck(buf), nil
	case UNSUBSCRIBE:
		return CreateUnSubscribe(buf), nil
	case UNSUBACK:
		return CreateUnSubAck(buf), nil
	case PING:
		return CreatePing(buf), nil
	case PONG:
		return CreatePong(buf), nil
	case DISCONNECT:
		return CreateDisconnect(buf), nil
	}

	return nil, ErrInvalidPacketType
}

func NewPacket(t Type) Packet {
	switch t {
	case CONNECT:
		return NewConnect()
	case CONNACK:
		return NewConnAck()
	case PUBLISH:
		return NewPublish()
	case PUBACK:
		return NewPubAck()
	case PUBREC:
		return NewPubRec()
	case PUBREL:
		return NewPubRel()
	case PUBCOMP:
		return NewPubComp()
	case SUBSCRIBE:
		return NewSubscribe()
	case SUBACK:
		return NewSubAck()
	case UNSUBSCRIBE:
		return NewUnSub()
	case UNSUBACK:
		return NewUnSubAck()
	case PING:
		return NewPing()
	case PONG:
		return NewPong()
	case DISCONNECT:
		return NewDisconnect()
	}

	return nil
}

func ReadInt8(buf []byte, offset int) (uint8, int, error) {
	var value uint8

	if len(buf) >= (offset + 1) {
		value = buf[offset]
		return value, offset + 1, nil
	}

	return 0, offset, ErrReadFromBuf
}

func ReadInt16(buf []byte, offset int) (uint16, int, error) {
	var value uint16

	if len(buf) >= (offset + 2) {
		value = uint16(buf[offset+1]) | uint16(buf[offset])<<8
		return value, offset + 2, nil
	}

	return 0, offset, ErrReadFromBuf
}

func ReadBytes(buf []byte, offset int, length int) ([]byte, int, error) {
	var value []byte

	if len(buf) >= (offset + length) {
		value = buf[offset : offset+length]
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
