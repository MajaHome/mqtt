package packet

import (
	"log"
	"strconv"
	"strings"
)

type ConnPacket struct {
	Header       []byte
	ClientID     string
	KeepAlive    uint16
	Username     string
	Password     string
	CleanSession bool
	Will         *Message
	Version      byte
}

func NewConnect() *ConnPacket {
	return &ConnPacket{}
}

func CreateConnect(buf []byte) *ConnPacket {
	return &ConnPacket{
		Header: buf,
	}
}

func (c *ConnPacket) Type() Type {
	return CONNECT
}

func (c *ConnPacket) Length() int {
	var l int = 2 /*hdr len*/ + len("MQTT") /*hdr name*/ + 1 /*version*/ + 1 /*flag*/
	if c.Will != nil {
		l += c.Will.Length()
	}
	l += 2 /*keepalive*/ + 2 /*cliendidlen*/ + len(c.ClientID) +
		2 /*len*/ + len(c.Username) + 2 /*len*/ + len(c.Password)

	return l
}

func (c *ConnPacket) Unpack(buf []byte) error {
	var offset int = 0

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
	if c.Version != byte(4) {
		return ErrUnsupportedVersion
	}
	if c.Version == 4 && hdr != "MQTT" {
		return ErrProtocolError
	}

	flag, offset, err := ReadInt8(buf, offset)
	if err != nil {
		return err
	}

	if flag&0x01 != 0 {
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
		log.Println("will is set")
		var willTopicLen, willMessageLen uint16
		var willTopic, willMessage string

		willTopicLen, offset, err = ReadInt16(buf, offset)
		if err != nil {
			return err
		}
		willTopic, offset, err = ReadString(buf, offset, int(willTopicLen))

		willMessageLen, offset, err = ReadInt16(buf, offset)
		if err != nil {
			return err
		}
		willMessage, offset, err = ReadString(buf, offset, int(willMessageLen))

		c.Will = &Message{
			QoS:       willQoS,
			Retain:    willRetain,
			Topic:     willTopic,
			Payload:   willMessage,
			Dublicate: false,
			Flag:      false,
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

func (c *ConnPacket) Pack() []byte {
	offset := 0
	buf := make([]byte, c.Length()+2)

	offset = WriteInt8(buf, offset, byte(CONNECT)<<4)
	offset = WriteInt8(buf, offset, byte(c.Length()))
	offset = WriteInt16(buf, offset, uint16(len("MQTT")))

	copy(buf[offset:], []byte("MQTT"))
	offset += 4
	offset = WriteInt8(buf, offset, c.Version)

	var flag uint8 = 0x0
	if len(c.Username) > 0 {
		flag |= 128 // 1000 0000
	}

	if len(c.Password) > 0 {
		flag |= 64 // 0100 0000
	}

	if c.Will != nil {
		flag |= 0x4 // 00000100

		if c.Will.Retain {
			flag |= 32 // 00100000
		}

		flag = (flag & 231) | (byte(c.Will.QoS) << 3) // 231 = 11100111
	}

	if c.CleanSession {
		flag |= 0x2 // 00000010
	}
	offset = WriteInt8(buf, offset, flag)

	offset = WriteInt16(buf, offset, c.KeepAlive)

	if c.Will != nil {
		offset = WriteInt16(buf, offset, uint16(len(c.Will.Topic)))
		copy(buf[offset:], c.Will.Topic)

		offset = WriteInt16(buf, offset, uint16(len(c.Will.Payload)))
		copy(buf[offset:], c.Will.Payload)
	}

	offset = WriteInt16(buf, offset, uint16(len(c.ClientID)))
	copy(buf[offset:], c.ClientID)
	offset += len(c.ClientID)

	if c.Will != nil {
		copy(buf[offset:], c.Will.Pack())
		offset += c.Will.Length()
	}

	offset = WriteInt16(buf, offset, uint16(len(c.Username)))
	copy(buf[offset:], c.Username)
	offset += len(c.Username)

	offset = WriteInt16(buf, offset, uint16(len(c.Password)))
	copy(buf[offset:], c.Password)

	return buf
}

func (c *ConnPacket) String() string {
	var sb strings.Builder

	sb.WriteString("Message Connect: {")
	sb.WriteString("ver=" + strconv.Itoa(int(c.Version)) + ", ")
	sb.WriteString("keepalive=" + strconv.Itoa(int(c.KeepAlive)) + ", ")
	sb.WriteString("clientId=" + c.ClientID + ", ")
	if c.Will != nil {
		sb.WriteString("will=" + c.Will.String() + ", ")
	}
	sb.WriteString("login=" + c.Username + ", ")
	sb.WriteString("password=" + c.Password)
	sb.WriteString("}")

	return sb.String()
}
