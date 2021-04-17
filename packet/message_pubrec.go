package packet

import (

)

type PublishRecPacket struct {
}

func PublishRec() *PublishRecPacket {
	return &PublishRecPacket{}
}

func (p *PublishRecPacket) Type() Type {
	return PUBREC
}

func (p *PublishRecPacket) Length() int {
	var l = 0

	return 2 + l
}

func (p *PublishRecPacket) Unpack(buf []byte) error {
	return nil
}

func (p *PublishRecPacket) Pack() ([]byte, error) {
	return nil, nil
}

func (p *PublishRecPacket) ToString() string {
	return "publishrec: {}"
}
