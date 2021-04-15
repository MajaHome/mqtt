package packet

import (
	"strconv"
)

type Message struct {
	Topic   string
	QoS     QoS
	Retain  bool
	Payload []byte
	Dublicate bool			// dublicate packet
	DisconnectFlag bool		// will message flag
}

func (m *Message) String() string {
	return "Message { topic=" + m.Topic + ", QoS=" + m.QoS.ToString() + ", retain=" + strconv.FormatBool(m.Retain) +
		", payload=" + string(m.Payload) + "}"
}
