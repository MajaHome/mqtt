package packet

import "strconv"

type PubCompPacket struct {
	Header []byte
	Id     uint16
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
	return 2
}

func (p *PubCompPacket) Unpack(buf []byte) error {
	id, _, err := ReadInt16(buf, 0)
	if err != nil {
		return err
	}
	p.Id = id

	return nil
}

func (p *PubCompPacket) Pack() []byte {
	offset := 0
	buf := make([]byte, 4)

	offset = WriteInt8(buf, offset, byte(PUBCOMP)<<4)
	offset = WriteInt8(buf, offset, byte(p.Length()))
	offset = WriteInt16(buf, offset, p.Id)

	return buf
}

func (p *PubCompPacket) String() string {
	return "Message PubComp: {id=" + strconv.Itoa(int(p.Id)) + " }"
}
