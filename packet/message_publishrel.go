package packet

import ()

type PubRelPacket struct {
	Header []byte
}

func NewPubRel() *PubRelPacket {
	return &PubRelPacket{}
}

func CreatePubRel(buf []byte) *PubRelPacket {
	return &PubRelPacket{
		Header: buf,
	}
}

func (p *PubRelPacket) Type() Type {
	return PUBREL
}

func (p *PubRelPacket) Length() int {
	return 0
}

func (p *PubRelPacket) Unpack(buf []byte) error {
	return nil
}

func (p *PubRelPacket) Pack() []byte {
	return nil
}

func (p *PubRelPacket) String() string {
	return "Message PubRel: {}"
}
