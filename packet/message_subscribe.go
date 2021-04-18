package packet

import (
	"strconv"
	"strings"
)

type SubscribePayload struct {
	QoS QoS
	Topic string
}

func (p *SubscribePayload) ToString() string {
	return "{topic:" + p.Topic + ", qos=" + p.QoS.ToString() + "}"
}

type SubscribePacket struct {
	Header []byte
	Id uint16
	Topics []SubscribePayload
}

func NewSubscribe() *SubscribePacket {
	return &SubscribePacket{}
}

func CreateSubscribe(buf []byte) *SubscribePacket {
	return &SubscribePacket{
		Header: buf,
	}
}

func (s *SubscribePacket) Type() Type {
	return SUBSCRIBE
}

func (s *SubscribePacket) Length() int {
	return 2 + 1 + len(s.Topics)
}

func (s *SubscribePacket) Unpack(buf []byte) error {
	var offset int = 0

	id, offset, err := ReadInt16(buf, offset)
	if err != nil {
		return err
	}
	s.Id = id

	var read uint8 = 2
	for s.Header[1] > read {
		topicLen, offset, err := ReadInt16(buf, offset)
		if err != nil {
			return err
		}

		topic, offset, err := ReadString(buf, offset, int(topicLen))
		if err != nil {
			return err
		}

		qos, offset, err := ReadInt8(buf, offset)

		s.Topics = append(s.Topics, SubscribePayload{Topic: topic, QoS: QoS(qos)})

		read += uint8(2 + len(topic) + 1)
	}

	return nil
}

func (s *SubscribePacket) Pack() ([]byte, error) {
	// todo

	return nil, nil
}

func (s *SubscribePacket) ToString() string {
	var sb strings.Builder

	sb.WriteString("Message Subscribe: {id:")
	sb.WriteString(strconv.Itoa(int(s.Id)))
	for _, t := range s.Topics {
		sb.WriteString(", payload: ")
		sb.WriteString(t.ToString())
	}
	sb.WriteString("}")

	return sb.String()
}
