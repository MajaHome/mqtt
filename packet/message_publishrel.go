package packet

import (

)

type PublishRelPacket struct {
}

func PublishRel() *PublishRelPacket {
	return &PublishRelPacket{}
}

func (p *PublishRelPacket) Type() Type {
	return PUBREL
}

func (p *PublishRelPacket) Length() int {
	var l = 0

	return 2 + l
}

func (p *PublishRelPacket) Unpack(buf []byte) error {
	return nil
}

func (p *PublishRelPacket) Pack() ([]byte, error) {
	return nil, nil
}

func (p *PublishRelPacket) ToString() string {
	return "publishrel: {}"
}
