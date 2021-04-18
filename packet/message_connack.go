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

type ConnAckPacket struct {
	Header []byte
	Session    bool
	ReturnCode uint8
}

func NewConnAck() *ConnAckPacket {
	return &ConnAckPacket{}
}

func CreateConnAck(buf []byte) *ConnAckPacket {
	return &ConnAckPacket{
		Header: buf,
		Session: false,
	}
}

func (cack *ConnAckPacket) Type() Type {
	return CONNACK
}

func (cack *ConnAckPacket) Length() int {
	return 2
}

func (cack *ConnAckPacket) Unpack(buf []byte) error {
	var offset int = 0

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

func (cack *ConnAckPacket) Pack() ([]byte, error) {
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

func (cack *ConnAckPacket) ToString() string {
	return "Message ConnAck: { session: " + strconv.FormatBool(cack.Session) + ", code:" +
		strconv.Itoa(int(cack.ReturnCode)) + "}"
}
