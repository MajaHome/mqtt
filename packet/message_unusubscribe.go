package packet

import (
	"fmt"
	"github.com/MajaSuite/mqtt/utils"
)

type UnSubscribePacket struct {
	PacketImpl
	Header byte
	Id     uint16
	Topics []SubscribePayload
}

func NewUnSub() *UnSubscribePacket {
	return &UnSubscribePacket{}
}

func CreateUnSubscribe(buf byte) *UnSubscribePacket {
	return &UnSubscribePacket{
		Header: buf,
		Topics: []SubscribePayload{},
	}
}

func (u *UnSubscribePacket) Type() Type {
	return UNSUBSCRIBE
}

func (u *UnSubscribePacket) Length() int {
	var l int
	for _, p := range u.Topics {
		l += p.Length()
	}
	return 2 /*id*/ + l
}

func (u *UnSubscribePacket) Unpack(buf []byte) error {
	id, offset, err := utils.ReadInt16(buf, 0)
	if err != nil {
		return err
	}
	u.Id = id

	for left := len(buf) - 2; left > 0; {
		var topicLen uint16
		var topic string
		var qos uint8

		topicLen, offset, err = utils.ReadInt16(buf, offset)
		if err != nil {
			return err
		}

		topic, offset, err = utils.ReadString(buf, offset, int(topicLen))
		if err != nil {
			return err
		}

		qos, offset, err = utils.ReadInt8(buf, offset)

		u.Topics = append(u.Topics, SubscribePayload{Topic: topic, QoS: QoS(qos)})
		left -= 2 + len(topic) + 1
	}

	return nil
}

func (u *UnSubscribePacket) Pack() []byte {
	lenBuff := WriteLength(u.Length())
	buf := make([]byte, 1+len(lenBuff)+u.Length())

	offset := utils.WriteInt8(buf, 0, byte(UNSUBSCRIBE)<<4)
	offset = utils.WriteBytes(buf, offset, lenBuff)
	offset = utils.WriteInt16(buf, offset, u.Id)

	for _, t := range u.Topics {
		data := t.Pack()
		copy(buf[offset:], data)
		offset += len(data)
	}

	return buf
}

func (u *UnSubscribePacket) String() string {
	var topics string
	for _, t := range u.Topics {
		topics += t.String() + ", "
	}

	return fmt.Sprintf("Unsubscribe: {id: %d, topics: [%s]}", u.Id, topics)
}
