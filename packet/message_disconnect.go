package packet

type DisconnectPacket struct {}

func Disconnect() *DisconnectPacket {
	return &DisconnectPacket{}
}

func (cack *DisconnectPacket) Type() Type {
	return DISCONNECT
}

func (cack *DisconnectPacket) Length() int {
	return 0
}

func (cack *DisconnectPacket) Unpack(buf []byte) error {
	return nil
}

func (cack *DisconnectPacket) Pack() ([]byte, error) {
	buf := make([]byte, 2)

	buf[0] = byte(DISCONNECT) << 4
	buf[1] = byte(cack.Length()) // Size

	return buf, nil
}

func (cack *DisconnectPacket) ToString() string {
	return "disconnect: {}"
}
