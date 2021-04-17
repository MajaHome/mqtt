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

func (m *Message) String() string {
	return "Message { topic=" + m.Topic + ", QoS=" + m.QoS.ToString() + ", retain=" + strconv.FormatBool(m.Retain) +
		", payload=" + m.Payload + ", dublicate=" + strconv.FormatBool(m.Dublicate) + ", flag=" +
		strconv.FormatBool(m.Flag) + "}"
}
