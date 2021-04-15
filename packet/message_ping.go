package packet

import (

)

type PingPacket struct {
}

func Ping() *PingPacket {
	return &PingPacket{}
}

func (p *PingPacket) Type() Type {
	return PING
}

func (p *PingPacket) Length() int {
	return 0
}

func (p *PingPacket) Unpack(buf []byte) error {
	// todo
	return nil
}

func (p *PingPacket) Pack() ([]byte, error) {
	// todo
	return nil, nil
}

func (p *PingPacket) ToString() string {
	return "ping: {}"
}
