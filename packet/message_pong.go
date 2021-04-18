package packet

type PongPacket struct {
	Header []byte
}

func NewPong() *PongPacket {
	return &PongPacket{}
}

func CreatePong(buf []byte) *PongPacket {
	return &PongPacket{
		Header: buf,
	}
}

func (p *PongPacket) Type() Type {
	return PONG
}

func (p *PongPacket) Length() int {
	return 0
}

func (p *PongPacket) Unpack(buf []byte) error {
	return nil
}

func (p *PongPacket) Pack() []byte {
	buf := make([]byte, 2)

	buf[0] = byte(PONG) << 4
	buf[1] = byte(p.Length())	// Size

	return buf
}

func (p *PongPacket) ToString() string {
	return "Message Pong: {}"
}
