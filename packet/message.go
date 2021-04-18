package packet

import (
	"strconv"
)

type Message struct {
	Topic     string
	QoS       QoS
	Retain    bool
	Payload   string
	Dublicate bool // dublicate packet
	Flag      bool // will message flag
}

func (m *Message) Length() int {
	return len(m.Topic) + 2 + len(m.Payload) + 2
}

func (m *Message) Pack() []byte {
	var offset int = 0
	buf := make([]byte, m.Length())

	offset = WriteInt16(buf, offset, uint16(len(m.Topic)))
	copy(buf[offset:], m.Topic)
	offset += len(m.Topic)

	offset = WriteInt16(buf, offset, uint16(len(m.Payload)))
	copy(buf[offset:], m.Payload)
	offset += len(m.Payload)

	return buf
}

func (m *Message) String() string {
	return "Message { topic=" + m.Topic + ", QoS=" + m.QoS.ToString() + ", retain=" + strconv.FormatBool(m.Retain) +
		", payload=" + m.Payload + ", dublicate=" + strconv.FormatBool(m.Dublicate) + ", flag=" +
		strconv.FormatBool(m.Flag) + "}"
}
