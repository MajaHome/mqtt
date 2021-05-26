package packet

import "strconv"

type PubRelPacket struct {
	Header []byte
	Id     uint16
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
	return 2
}

func (p *PubRelPacket) Unpack(buf []byte) error {
	id, _, err := ReadInt16(buf, 0)
	if err != nil {
		return err
	}
	p.Id = id

	return nil
}

func (p *PubRelPacket) Pack() []byte {
	offset := 0
	buf := make([]byte, 4)

	offset = WriteInt8(buf, offset, byte(PUBREL)<<4)
	offset = WriteInt8(buf, offset, byte(p.Length()))
	offset = WriteInt16(buf, offset, p.Id)

	return buf
}

func (p *PubRelPacket) String() string {
	return "Message PubRel: {id=" + strconv.Itoa(int(p.Id)) + " }"
}
