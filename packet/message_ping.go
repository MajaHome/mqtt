package packet

type PingPacket struct {
	PacketImpl
	Header byte
}

func NewPing() *PingPacket {
	return &PingPacket{}
}
func CreatePing(buf byte) *PingPacket {
	return &PingPacket{
		Header: buf,
	}
}

func (p *PingPacket) Type() Type {
	return PING
}

func (p *PingPacket) Length() int {
	return 0
}

func (p *PingPacket) Unpack(buf []byte) error {
	return nil
}

func (p *PingPacket) Pack() []byte {
	buf := make([]byte, 2)

	buf[0] = byte(PING) << 4
	buf[1] = byte(p.Length())

	return buf
}

func (p *PingPacket) String() string {
	return "Ping: {}"
}
