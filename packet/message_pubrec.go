package packet

import (

)

type PubRecPacket struct {
	Header []byte
}

func NewPubRec() *PubRecPacket {
	return &PubRecPacket{}
}

func CreatePubRec(buf []byte) *PubRecPacket {
	return &PubRecPacket{
		Header: buf,
	}
}

func (p *PubRecPacket) Type() Type {
	return PUBREC
}

func (p *PubRecPacket) Length() int {
	return 0
}

func (p *PubRecPacket) Unpack(buf []byte) error {
	return nil
}

func (p *PubRecPacket) Pack() []byte {
	return nil
}

func (p *PubRecPacket) ToString() string {
	return "Message PubRec: {}"
}
