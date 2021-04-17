package packet

import (
	"strconv"
)

type PublishPacket struct {
	Id uint16
	QoS       QoS
	Retain    bool
	DUP bool
	Topic string
	Payload string
}

func Publish() *PublishPacket {
	return &PublishPacket{}
}

func (p *PublishPacket) Type() Type {
	return PUBLISH
}

func (p *PublishPacket) Length() int {
	return 2 + 2/*topic len*/ + len(p.Topic) + len(p.Payload)
}

func (p *PublishPacket) Unpack(buf []byte) error {
	var offset int = 0

	// packet type
	packetType, offset, err := ReadInt8(buf, offset)
	if err != nil {
		return err
	}

	// packet len
	packetLen, offset, err := ReadInt8(buf, offset)
	if err != nil {
		return err
	}

	p.DUP = packetType >> 3 & 0x1 == 1
	p.Retain = packetType & 0x1 == 1
	p.QoS = QoS(packetType >> 1 & 0x3)

	topicLen, offset, err := ReadInt16(buf, offset)
	if err != nil {
		return err
	}

	p.Topic, offset, err = ReadString(buf, offset, int(topicLen))
	if err != nil {
		return err
	}

	if p.QoS > 0 {
		p.Id, offset, err = ReadInt16(buf, offset)
		if err != nil {
			return err
		}
	}

	p.Payload, offset, err = ReadString(buf, offset, int(int(packetLen) - offset))
	if err != nil {
		return err
	}

	return nil
}

func (p *PublishPacket) Pack() ([]byte, error) {
	offset := 0
	buf := make([]byte, 4)

	var packType uint8 = byte(PUBLISH) << 4
	if p.Retain {
		packType |= 0x01
	}

	if p.DUP {
		packType |= 0x8
	}

	if !p.QoS.Valid() {
		return nil, ErrInvalidQos
	}
	if p.QoS == QoS(1) {
		packType |= 0x2
	}
	if p.QoS == QoS(2) {
		packType |= 0x4
	}

	offset = WriteInt8(buf, offset, packType)
	offset = WriteInt8(buf, offset, byte(p.Length()))
	offset = WriteInt16(buf, offset, uint16(len(p.Topic)))
	buf = append(buf, p.Topic...)
	buf = append(buf, p.Payload...)

	return buf, nil
}

func (p *PublishPacket) ToString() string {
	return "MessagePublish: {id=" + strconv.Itoa(int(p.Id)) + ", topic=" + p.Topic + ", payload=" +
		p.Payload + ", qos=" + p.QoS.ToString() + ", retain=" + strconv.FormatBool(p.Retain) +
		", dup=" + strconv.FormatBool(p.DUP) + "}"
}
