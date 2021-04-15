package packet

import (

)

type DisconnectPacket struct {
	Session bool
	ReturnCode uint8
}

func Disconnect() *DisconnectPacket {
	return &DisconnectPacket{}
}

func (cack *DisconnectPacket) Type() Type {
	return DISCONNECT
}

func (cack *DisconnectPacket) Length() int {
	var l uint8 = 0

	return int(l)
}

func (cack *DisconnectPacket) Unpack(buf []byte) error {
	return nil
}

func (cack *DisconnectPacket) Pack() ([]byte, error) {
	return nil, nil
}

func (cack *DisconnectPacket) ToString() string {
	return "disconnect: {}"
}
