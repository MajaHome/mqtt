package packet

type DisconnectPacket struct {
	Header []byte
}

func NewDisconnect() *DisconnectPacket {
	return &DisconnectPacket{}
}
func CreateDisconnect(buf []byte) *DisconnectPacket {
	return &DisconnectPacket{
		Header: buf,
	}
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

func (cack *DisconnectPacket) Pack() []byte {
	buf := make([]byte, 2)

	buf[0] = byte(DISCONNECT) << 4
	buf[1] = byte(cack.Length()) // Size

	return buf
}

func (cack *DisconnectPacket) ToString() string {
	return "Message Disconnect: {}"
}
