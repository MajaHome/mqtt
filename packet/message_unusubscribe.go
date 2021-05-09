package packet

import (
	"strconv"
	"strings"
)

type UnSubscribePacket struct {
	Header []byte
	Id     uint16
	Topics []SubscribePayload
}

func NewUnSub() *UnSubscribePacket {
	return &UnSubscribePacket{}
}

func CreateUnSubscribe(buf []byte) *UnSubscribePacket {
	return &UnSubscribePacket{
		Header: buf,
	}
}

func (u *UnSubscribePacket) Type() Type {
	return UNSUBSCRIBE
}

func (u *UnSubscribePacket) Length() int {
	var l int = 0
	for _, p := range u.Topics {
		l += p.Length()
	}
	return 2 + 2 + len(u.Topics) + l
}

func (u *UnSubscribePacket) Unpack(buf []byte) error {
	var offset int = 0

	id, offset, err := ReadInt16(buf, offset)
	if err != nil {
		return err
	}
	u.Id = id

	var read uint8 = 2
	for u.Header[1] > read {
		topicLen, offset, err := ReadInt16(buf, offset)
		if err != nil {
			return err
		}

		topic, offset, err := ReadString(buf, offset, int(topicLen))
		if err != nil {
			return err
		}

		qos, offset, err := ReadInt8(buf, offset)

		u.Topics = append(u.Topics, SubscribePayload{Topic: topic, QoS: QoS(qos)})

		read += uint8(2 + len(topic) + 1)
	}

	return nil
}

func (u *UnSubscribePacket) Pack() []byte {
	var offset int = 0
	buf := make([]byte, 4)
	offset = WriteInt8(buf, offset, byte(UNSUBSCRIBE)<<4)
	offset = WriteInt8(buf, offset, byte(u.Length()))
	offset = WriteInt16(buf, offset, u.Id)

	for _, t := range u.Topics {
		buf = append(buf, t.Pack()...)
	}

	return buf
}

func (u *UnSubscribePacket) String() string {
	var sb strings.Builder

	sb.WriteString("Message Unsubscribe: {id:")
	sb.WriteString(strconv.Itoa(int(u.Id)))
	for _, t := range u.Topics {
		sb.WriteString(", payload: ")
		sb.WriteString(t.String())
	}
	sb.WriteString("}")

	return sb.String()
}
