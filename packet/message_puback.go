package packet

import (
	"strconv"
)

type PublishAckPacket struct {
	Id uint16
}

func PublishAck() *PublishAckPacket {
	return &PublishAckPacket{}
}

func (pack *PublishAckPacket) Type() Type {
	return PUBACK
}

func (pack *PublishAckPacket) Length() int {
	return 2 + 2
}

func (pack *PublishAckPacket) Unpack(buf []byte) error {
	var offset int = 0

	// packet type
	_, offset, err := ReadInt8(buf, offset)
	if err != nil {
		return err
	}

	// packet len
	_, offset, err = ReadInt8(buf, offset)
	if err != nil {
		return err
	}

	pack.Id, offset, err = ReadInt16(buf, offset)
	if err != nil {
		return err
	}

	return nil
}

func (pack *PublishAckPacket) Pack() ([]byte, error) {
	offset := 0
	buf := make([]byte, 4)

	offset = WriteInt8(buf, offset, byte(PUBACK) << 4)
	offset = WriteInt8(buf, offset, byte(pack.Length()))
	offset = WriteInt16(buf, offset, pack.Id)

	return buf, nil
}

func (pack *PublishAckPacket) ToString() string {
	return "Message PubAck: { id=" + strconv.Itoa(int(pack.Id)) + "}"
}
