package packet

import (
	"strconv"
	"strings"
)

type SubscribePayload struct {
	QoS   QoS
	Topic string
}

func (p *SubscribePayload) Length() int {
	return len(p.Topic) + 1 /*qos*/ + 2 /*len*/
}

func (p *SubscribePayload) Pack() []byte {
	var offset int = 0
	buf := make([]byte, p.Length())

	offset = WriteInt16(buf, offset, uint16(len(p.Topic)))
	copy(buf[offset:], p.Topic)
	offset += len(p.Topic)
	offset = WriteInt8(buf, offset, uint8(p.QoS))

	return buf
}

func (p *SubscribePayload) String() string {
	return "{topic:" + p.Topic + ", qos=" + p.QoS.String() + "}"
}

type SubscribePacket struct {
	Header []byte
	Id     uint16
	Topics []SubscribePayload
}

func NewSubscribe() *SubscribePacket {
	return &SubscribePacket{}
}

func CreateSubscribe(buf []byte) *SubscribePacket {
	return &SubscribePacket{Header: buf}
}

func (s *SubscribePacket) Type() Type {
	return SUBSCRIBE
}

func (s *SubscribePacket) Length() int {
	var l int = 0
	for _, p := range s.Topics {
		l += p.Length()
	}
	return 2/*id*/ + l
}

func (s *SubscribePacket) Unpack(buf []byte) error {
	var offset int = 0
	var err error

	s.Id, offset, err = ReadInt16(buf, offset)
	if err != nil {
		return err
	}

	for s.Header[1] > uint8(offset) {
		var topicLen uint16
		topicLen, offset, err = ReadInt16(buf, offset)
		if err != nil {
			return err
		}

		var topic string
		topic, offset, err = ReadString(buf, offset, int(topicLen))
		if err != nil {
			return err
		}

		var qos uint8
		qos, offset, err = ReadInt8(buf, offset)

		s.Topics = append(s.Topics, SubscribePayload{Topic: topic, QoS: QoS(qos)})
	}

	return nil
}

func (s *SubscribePacket) Pack() []byte {
	var offset int = 0
	buf := make([]byte, 4)
	offset = WriteInt8(buf, offset, byte(SUBSCRIBE)<<4)
	offset = WriteInt8(buf, offset, byte(s.Length()))
	offset = WriteInt16(buf, offset, s.Id)

	for _, t := range s.Topics {
		buf = append(buf, t.Pack()...)
	}

	return buf
}

func (s *SubscribePacket) String() string {
	var sb strings.Builder

	sb.WriteString("Message Subscribe: {id:")
	sb.WriteString(strconv.Itoa(int(s.Id)))
	for _, t := range s.Topics {
		sb.WriteString(", payload: ")
		sb.WriteString(t.String())
	}
	sb.WriteString("}")

	return sb.String()
}
