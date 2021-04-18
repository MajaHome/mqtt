package packet

import (

)

type UnSubscriibePacket struct {
	Header []byte
}

func NewUnSub() *UnSubscriibePacket {
	return &UnSubscriibePacket{}
}

func CreateUnSubscribe(buf []byte) *UnSubscriibePacket {
	return &UnSubscriibePacket{
		Header: buf,
	}
}

func (u *UnSubscriibePacket) Type() Type {
	return UNSUBSCRIBE
}

func (u *UnSubscriibePacket) Length() int {
	return 2
}

func (u *UnSubscriibePacket) Unpack(buf []byte) error {
	return nil
}

func (u *UnSubscriibePacket) Pack() []byte {
	return nil
}

func (u *UnSubscriibePacket) ToString() string {
	return "Message Unsubscribe: {}"
}
