package packet

import (

)

type PubCompPacket struct {
	Header []byte
}

func NewPubComp() *PubCompPacket {
	return &PubCompPacket{}
}

func CreatePubComp(buf []byte) *PubCompPacket {
	return &PubCompPacket{
		Header: buf,
	}
}

func (p *PubCompPacket) Type() Type {
	return PUBCOMP
}

func (p *PubCompPacket) Length() int {
	return 0
}

func (p *PubCompPacket) Unpack(buf []byte) error {
	return nil
}

func (p *PubCompPacket) Pack() []byte {
	return nil
}

func (p *PubCompPacket) ToString() string {
	return "Message PubComp: {}"
}
