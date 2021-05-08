package packet

import (
	"strconv"
)

type PublishPacket struct {
	Header  []byte
	Id      uint16
	QoS     QoS
	Retain  bool
	DUP     bool
	Topic   string
	Payload string
}

func NewPublish() *PublishPacket {
	return &PublishPacket{}
}

func CreatePublish(buf []byte) *PublishPacket {
	return &PublishPacket{
		Header: buf,
	}
}

func (p *PublishPacket) Type() Type {
	return PUBLISH
}

func (p *PublishPacket) Length() int {
	var l = 2 /*topic len*/ + len(p.Topic) + len(p.Payload)
	if p.QoS > 0 {
		l += 2
	}
	return l
}

func (p *PublishPacket) Unpack(buf []byte) error {
	// packet type
	packetType := p.Header[0]

	// packet len
	packetLen := p.Header[1]

	var offset int = 0

	p.DUP = packetType>>3&0x1 == 1
	p.Retain = packetType&0x1 == 1
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

	p.Payload, offset, err = ReadString(buf, offset, int(int(packetLen)-offset))
	if err != nil {
		return err
	}

	return nil
}

func (p *PublishPacket) Pack() []byte {
	offset := 0
	buf := make([]byte, 4)

	var packType uint8 = byte(PUBLISH) << 4
	if p.Retain {
		packType |= 0x01
	}

	if p.DUP {
		packType |= 0x8
	}

	// do not check QoS is valid
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
	offset += len(p.Topic)
	if p.QoS > 0 {
		tmp := make([]byte, 2)
		_ = WriteInt16(tmp, 0, p.Id)
		buf = append(buf, tmp...)
	}
	buf = append(buf, p.Payload...)

	return buf
}

func (p *PublishPacket) String() string {
	return "Message Publish: {id=" + strconv.Itoa(int(p.Id)) + ", topic=" + p.Topic + ", payload=" +
		p.Payload + ", qos=" + p.QoS.ToString() + ", retain=" + strconv.FormatBool(p.Retain) +
		", dup=" + strconv.FormatBool(p.DUP) + "}"
}
