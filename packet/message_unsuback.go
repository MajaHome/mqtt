package packet

import (
	"encoding/binary"
	"strconv"
)

type UnSubAckPacket struct {
	Header []byte
	Id     uint16
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
	var offset int = 0

	id, offset, err := ReadInt16(buf, offset)
	if err != nil {
		return err
	}
	uack.Id = id

	return nil
}

func (uack *UnSubAckPacket) Pack() []byte {
	buf := make([]byte, 4)

	buf[0] = byte(UNSUBACK) << 4
	buf[1] = byte(uack.Length())
	binary.BigEndian.PutUint16(buf[2:], uack.Id)

	return buf
}

func (uack *UnSubAckPacket) String() string {
	return "Message UnSubAck: {id=" + strconv.Itoa(int(uack.Id)) + "}"
}
