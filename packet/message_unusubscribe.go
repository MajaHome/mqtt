package packet

import (

)

type UnSubscribePacket struct {
}

func UnSubscribe() *UnSubscribePacket {
	return &UnSubscribePacket{}
}

func (u *UnSubscribePacket) Type() Type {
	return UNSUBSCRIBE
}

func (u *UnSubscribePacket) Length() int {
	var l = 0

	return 2 + l
}

func (u *UnSubscribePacket) Unpack(buf []byte) error {
	return nil
}

func (u *UnSubscribePacket) Pack() ([]byte, error) {
	return nil, nil
}

func (u *UnSubscribePacket) ToString() string {
	return "unsubscribe: {}"
}
