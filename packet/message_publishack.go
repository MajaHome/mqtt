package packet

import (

)

type PublishAckPacket struct {
}

func PublishAck() *PublishAckPacket {
	return &PublishAckPacket{}
}

func (pack *PublishAckPacket) Type() Type {
	return PUBLISHACK
}

func (pack *PublishAckPacket) Length() int {
	var l = 0

	return 2 + l
}

func (pack *PublishAckPacket) Unpack(buf []byte) error {
	return nil
}

func (pack *PublishAckPacket) Pack() ([]byte, error) {
	return nil, ErrUnknownPacket
}

func (pack *PublishAckPacket) ToString() string {
	return "publishack: {}"
}
