package packet

type PingPacket struct {}

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

func (p *PingPacket) Pack() ([]byte, error) {
	buf := make([]byte, 2)

	buf[0] = byte(PING) << 4
	buf[1] = byte(p.Length()) // Size

	return buf, nil
}

func (p *PingPacket) ToString() string {
	return "ping: {}"
}
