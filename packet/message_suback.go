package packet

import (
	"encoding/binary"
	"strings"
)

type SubAckPacket struct {
	Header []byte
	Id uint16
	ReturnCodes []QoS
}

func NewSubAck() *SubAckPacket {
	return &SubAckPacket{}
}

func CreateSubAck(buf []byte) *SubAckPacket {
	return &SubAckPacket{
		Header: buf,
	}
}

func (sack *SubAckPacket) Type() Type {
	return SUBACK
}

func (sack *SubAckPacket) Length() int {
	return 2 + len(sack.ReturnCodes)
}

func (sack *SubAckPacket) Unpack(buf []byte) error {
	var offset int = 0

	id, offset, err := ReadInt16(buf, offset)
	if err != nil {
		return err
	}
	sack.Id = id

	var read uint8 = 2
	for sack.Header[1] > read {
		var qos byte
		qos, offset, err = ReadInt8(buf, offset)
		sack.ReturnCodes = append(sack.ReturnCodes, QoS(qos))
		read++
	}

	return nil
}

func (sack *SubAckPacket) Pack() []byte {
	buf := make([]byte, 4)

	buf[0] = byte(SUBACK) << 4
	buf[1] = byte(sack.Length())	// Size
	binary.BigEndian.PutUint16(buf[2:], sack.Id)

	for _, rc := range sack.ReturnCodes {
		buf = append(buf, byte(rc))
	}

	return buf
}

func (sack *SubAckPacket) ToString() string {
	var sb strings.Builder

	sb.WriteString("Message SubAck: [")
	for _, rc := range sack.ReturnCodes {
		sb.WriteString(", returnCode: ")
		sb.WriteString(rc.ToString())
	}
	sb.WriteString("]")

	return sb.String()
}
