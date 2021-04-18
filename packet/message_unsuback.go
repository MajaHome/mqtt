package packet

import (

)

type UnSubAckPacket struct {
	Header []byte
}

func NewUnSubAck() *UnSubAckPacket {
	return &UnSubAckPacket{}
}

func CreateUnSubAck(buf []byte) *UnSubAckPacket {
	return &UnSubAckPacket{
		Header: buf,
	}
}

func (uack *UnSubAckPacket) Type() Type {
	return UNSUBACK
}

func (uack *UnSubAckPacket) Length() int {
	return 2
}

func (uack *UnSubAckPacket) Unpack(buf []byte) error {
	return nil
}

func (uack *UnSubAckPacket) Pack() []byte {
	return nil
}

func (uack *UnSubAckPacket) ToString() string {
	return "Message UnSubAck: {}"
}
