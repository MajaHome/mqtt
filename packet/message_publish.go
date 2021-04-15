package packet

import (

)

type PublishPacket struct {
}

func Publish() *PublishPacket {
	return &PublishPacket{}
}

func (p *PublishPacket) Type() Type {
	return PUBLISH
}

func (p *PublishPacket) Length() int {
	var l = 0

	return 2 + l
}

func (p *PublishPacket) Unpack(buf []byte) error {
	return nil
}

func (p *PublishPacket) Pack() ([]byte, error) {
	return nil, nil
}

func (p *PublishPacket) ToString() string {
	return "publish: {}"
}
