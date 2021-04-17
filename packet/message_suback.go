package packet

import (
	"encoding/binary"
	"strings"
)

type SubscribeAckPacket struct {
	Id uint16
	ReturnCodes []QoS
}

func SubscribeAck() *SubscribeAckPacket {
	return &SubscribeAckPacket{}
}

func (sack *SubscribeAckPacket) Type() Type {
	return SUBACK
}

func (sack *SubscribeAckPacket) Length() int {
	return 2 + len(sack.ReturnCodes)
}

func (sack *SubscribeAckPacket) Unpack(buf []byte) error {
	var offset int = 0

	// packet type
	_, offset, err := ReadInt8(buf, offset)
	if err != nil {
		return err
	}

	// packet len
	bufLen, offset, err := ReadInt8(buf, offset)
	if err != nil {
		return err
	}

	sack.Id, offset, err = ReadInt16(buf, offset)
	if err != nil {
		return err
	}

	var read uint8 = 2
	for bufLen > read {
		var qos byte
		qos, offset, err = ReadInt8(buf, offset)
		sack.ReturnCodes = append(sack.ReturnCodes, QoS(qos))
		read++
	}

	return nil
}

func (sack *SubscribeAckPacket) Pack() ([]byte, error) {
	buf := make([]byte, 4)

	buf[0] = byte(SUBACK) << 4
	buf[1] = byte(sack.Length())	// Size
	binary.BigEndian.PutUint16(buf[2:], sack.Id)

	for _, rc := range sack.ReturnCodes {
		buf = append(buf, byte(rc))
	}

	return buf, nil
}

func (sack *SubscribeAckPacket) ToString() string {
	var sb strings.Builder

	sb.WriteString("MessageSubAck: [")
	for _, rc := range sack.ReturnCodes {
		sb.WriteString(", returnCode: ")
		sb.WriteString(rc.ToString())
	}
	sb.WriteString("]")

	return sb.String()
}
