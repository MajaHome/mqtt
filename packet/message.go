package packet

import (
	"fmt"
)

type Message struct {
	Flag      bool
	QoS       QoS
	Retain    bool
	Dublicate bool
	Topic     string
	Payload   string
}

func (m *Message) Length() int {
	return len(m.Topic) + 2/*topicLen*/ + len(m.Payload) + 2/*payload len*/
}

func (m *Message) Pack() []byte {
	buf := make([]byte, m.Length())

	offset := WriteString(buf, 0, m.Topic)
	//offset = WriteInt16(buf, offset, uint16(len(m.Topic)))
	//copy(buf[offset:], m.Topic)
	//offset += len(m.Topic)

	offset = WriteString(buf, offset, m.Payload)
	//offset = WriteInt16(buf, offset, uint16(len(m.Payload)))
	//copy(buf[offset:], m.Payload)
	//offset += len(m.Payload)

	return buf
}

func (m *Message) String() string {
	return fmt.Sprintf("message: {topic: %s, qos: %d, retain: %v, dup: %v, flag: %v, payload: %s}",
		m.Topic, m.QoS.Int(), m.Retain, m.Dublicate, m.Flag, m.Payload)
}
