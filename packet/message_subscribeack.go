package packet

import (

)

type SubscribeAckPacket struct {
}

func SubscribeAck() *SubscribeAckPacket {
	return &SubscribeAckPacket{}
}

func (sack *SubscribeAckPacket) Type() Type {
	return SUBSCRIBEACK
}

func (sack *SubscribeAckPacket) Length() int {
	var l = 0

	return 2 + l
}

func (sack *SubscribeAckPacket) Unpack(buf []byte) error {
	return nil
}

func (sack *SubscribeAckPacket) Pack() ([]byte, error) {
	return nil, nil
}

func (sack *SubscribeAckPacket) ToString() string {
	return "subscribeack: {}"
}
