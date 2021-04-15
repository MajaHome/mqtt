package packet

import (

)

type UnSubscribeAckPacket struct {
}

func UnSubscribeAck() *UnSubscribeAckPacket {
	return &UnSubscribeAckPacket{}
}

func (uack *UnSubscribeAckPacket) Type() Type {
	return UNSUBSCRIBEACK
}

func (uack *UnSubscribeAckPacket) Length() int {
	var l = 0

	return 2 + l
}

func (uack *UnSubscribeAckPacket) Unpack(buf []byte) error {
	return nil
}

func (uack *UnSubscribeAckPacket) Pack() ([]byte, error) {
	return nil, nil
}

func (uack *UnSubscribeAckPacket) ToString() string {
	return "unsubscribeack: {}"
}
