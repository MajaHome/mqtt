package packet

import "strconv"

const (
	ConnectAccepted int = iota
	ConnectUnacceptableProtocol
	ConnectIndentifierRejected
	ConnectServerUnavailable
	ConnectBadUserPass
	ConnectNotAuthorized
)

type ConnectAckPacket struct {
	Session    bool
	ReturnCode uint8
}

func ConnectAck() *ConnectAckPacket {
	return &ConnectAckPacket{
		Session: false,
	}
}

func (cack *ConnectAckPacket) Type() Type {
	return CONNACK
}

func (cack *ConnectAckPacket) Length() int {
	return 2
}

func (cack *ConnectAckPacket) Unpack(buf []byte) error {
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

	acknowledge, offset, err := ReadInt8(buf, offset)
	if err != nil {
		return err
	}

	if acknowledge > 1 {
		return ErrProtocolError
	}
	cack.Session = (acknowledge == 1)

	cack.ReturnCode, offset, err = ReadInt8(buf, offset)
	if err != nil {
		return err
	}

	return nil
}

func (cack *ConnectAckPacket) Pack() ([]byte, error) {
	buf := make([]byte, 4)

	buf[0] = byte(CONNACK) << 4
	buf[1] = byte(cack.Length())	// Size
	if cack.Session {
		buf[2] = 0x01
	} else {
		buf[2] = 0
	}
	buf[3] = cack.ReturnCode

	return buf, nil
}

func (cack *ConnectAckPacket) ToString() string {
	return "connectack: { session: " + strconv.FormatBool(cack.Session) + ", code:" +
		strconv.Itoa(int(cack.ReturnCode)) + "}"
}
