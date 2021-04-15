package packet

import (

)

type PongPacket struct {
}

func Pong() *PongPacket {
	return &PongPacket{}
}

func (p *PongPacket) Type() Type {
	return PONG
}

func (p *PongPacket) Length() int {
	var l uint8 = 0

	return int(l)
}

func (p *PongPacket) Unpack(buf []byte) error {
	// todo
	return nil
}

func (p *PongPacket) Pack() ([]byte, error) {
	// todo
	return nil, nil
}

func (p *PongPacket) ToString() string {
	return "pong: {}"
}
