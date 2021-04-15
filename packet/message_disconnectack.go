package packet

import (
)

type DisconnectAckPacket struct {
}

func DisconnectAck() *DisconnectAckPacket {
	return &DisconnectAckPacket{}
}

func (cack *DisconnectAckPacket) Type() Type {
	return DISCONNECT
}

func (cack *DisconnectAckPacket) Length() int {
	var l uint8 = 0

	return int(l)
}

func (cack *DisconnectAckPacket) Unpack(buf []byte) error {
	return nil
}

func (cack *DisconnectAckPacket) Pack() ([]byte, error) {
	return nil, nil
}

func (cack *DisconnectAckPacket) ToString() string {
	return "disconnectAck: {}"
}
