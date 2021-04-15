package packet

import (
	"strconv"
)

type ConnectPacket struct {
	ClientID     string
	KeepAlive    uint16
	Username     string
	Password     string
	CleanSession bool
	Will         *Message
	Version      byte
}

func Connect() *ConnectPacket {
	return &ConnectPacket{}
}

func (c *ConnectPacket) Type() Type{
	return CONNECT
}

func (c *ConnectPacket) Length() int {
	var l int = 0
	l += 2/*packet type + size/ + 2 /*hdr len*/ + 6 /*hdr name*/ + 1 /*version*/ + 1 /*flag*/
	if c.Will != nil {
		l += 2 /*len*/ + len(c.Will.Topic) + 2 /*len*/ + len(c.Will.Payload)
	}
	l += 2 /*len*/ + len(c.Username) + 2/*len*/ + len(c.Password)

	return l
}

func (c *ConnectPacket) Unpack(buf []byte) error {
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

	hdrLen, offset, err := ReadInt16(buf, offset)
	if err != nil {
		return err
	}

	hdr, offset, err := ReadString(buf, offset, int(hdrLen))
	if err != nil {
		return err
	}

	c.Version, offset, err = ReadInt8(buf, offset)
	if err != nil {
		return err
	}
	if c.Version == 4 && hdr != "MQTT" {
		return ErrProtocolError
	}

	if c.Version != byte(4) {
		return ErrUnsupportedVersion
	}

	flag, offset, err := ReadInt8(buf, offset)
	if err != nil {
		return err
	}

	if flag & 0x01 != 0 {
		return ErrUnknownPacket
	}

	usernameFlag := ((flag >> 7) & 0x1) == 1
	passwordFlag := ((flag >> 6) & 0x1) == 1
	if !usernameFlag && passwordFlag {
		return ErrUnknownPacket
	}

	willFlag := ((flag >> 2) & 0x1) == 1
	willRetain := ((flag >> 5) & 0x1) == 1
	willQoS := QoS((flag >> 3) & 0x3)
	if !willQoS.Valid() {
		return ErrUnknownPacket
	}
	if !willFlag && (willRetain || willQoS != 0) {
		return ErrUnknownPacket
	}

	if willFlag {
		c.Will = &Message{
			QoS:            willQoS,
			Retain:         willRetain,
			Dublicate:      false,
			DisconnectFlag: false,
		}
	}

	c.CleanSession = ((flag >> 1) & 0x1) == 1

	c.KeepAlive, offset, err = ReadInt16(buf, offset)
	if err != nil {
		return err
	}

	clidLen, offset, err := ReadInt16(buf, offset)
	if err != nil {
		return err
	}
	if clidLen == 0 && !c.CleanSession {
		return ErrUnknownPacket
	}

	c.ClientID, offset, err = ReadString(buf, offset, int(clidLen))
	if err != nil {
		return err
	}

	if willFlag {
		tLen, offset, err := ReadInt16(buf, offset)
		if err != nil {
			return err
		}

		c.Will.Topic, offset, err = ReadString(buf, offset, int(tLen))
		if err != nil {
			return err
		}

		pLen, offset, err := ReadInt16(buf, offset)
		if err != nil {
			return err
		}

		c.Will.Payload, offset, err = ReadBytes(buf, offset, int(pLen))
		if err != nil {
			return err
		}
	}

	loginLen, offset, err := ReadInt16(buf, offset)
	if err != nil {
		return err
	}

	c.Username, offset, err = ReadString(buf, offset, int(loginLen))
	if err != nil {
		return err
	}

	passLen, offset, err := ReadInt16(buf, offset)
	if err != nil {
		return err
	}

	c.Password, offset, err = ReadString(buf, offset, int(passLen))
	if err != nil {
		return err
	}

	return nil
}

func (c *ConnectPacket) Pack() ([]byte, error) {
	buf := make([]byte, c.Length())

	buf[0] = byte(CONNECT) << 4
	buf[1] = byte(c.Length())

	// todo

	return buf, nil
}

func (c *ConnectPacket) ToString() string {
	return "MessageConnect {ver="+strconv.Itoa(int(c.Version))+", keepalive="+
		strconv.Itoa(int(c.KeepAlive))+", clientId="+c.ClientID+", login="+c.Username+", password="+c.Password+"}"
}
