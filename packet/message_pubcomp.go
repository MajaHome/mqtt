package packet

import (

)

type PublishCompPacket struct {
}

func PublishComp() *PublishCompPacket {
	return &PublishCompPacket{}
}

func (p *PublishCompPacket) Type() Type {
	return PUBCOMP
}

func (p *PublishCompPacket) Length() int {
	var l = 0

	return 2 + l
}

func (p *PublishCompPacket) Unpack(buf []byte) error {
	return nil
}

func (p *PublishCompPacket) Pack() ([]byte, error) {
	return nil, nil
}

func (p *PublishCompPacket) ToString() string {
	return "publishcomp: {}"
}
