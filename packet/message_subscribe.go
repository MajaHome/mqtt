package packet

import (

)

type SubscribePacket struct {
}

func Subscribe() *SubscribePacket {
	return &SubscribePacket{}
}

func (s *SubscribePacket) Type() Type {
	return SUBSCRIBE
}

func (s *SubscribePacket) Length() int {
	var l = 0

	return 2 + l
}

func (s *SubscribePacket) Unpack(buf []byte) error {
	return nil
}

func (s *SubscribePacket) Pack() ([]byte, error) {
	return nil, nil
}

func (s *SubscribePacket) ToString() string {
	return "subscribe: {}"
}
