package packet

type PongPacket struct {}

func Pong() *PongPacket {
	return &PongPacket{}
}

func (p *PongPacket) Type() Type {
	return PONG
}

func (p *PongPacket) Length() int {
	return 0
}

func (p *PongPacket) Unpack(buf []byte) error {
	var offset int = 0

	// packet type
	_, offset, err := ReadInt8(buf, offset)
	if err != nil {
		return err
	}

	// packet len
	_, offset, err = ReadInt8(buf, offset)
	if err != nil {
		return err
	}

	return nil
}

func (p *PongPacket) Pack() ([]byte, error) {
	buf := make([]byte, 2)

	buf[0] = byte(PONG) << 4
	buf[1] = byte(p.Length())	// Size

	return buf, nil
}

func (p *PongPacket) ToString() string {
	return "pong: {}"
}
