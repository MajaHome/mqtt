package transport

import (
	"github.com/MajaSuite/mqtt/packet"
	"strconv"
	"strings"
)

type Event struct {
	ClientId   string
	PacketType packet.Type
	MessageId  uint16
	Topic      EventTopic
	Payload    string
	Qos        int
	Retain     bool
	Dublicate  bool
}

type EventTopic struct {
	Name string
	Qos  int
}

func (t *EventTopic) String() string {
	return "{ name: " + t.Name + ", qos: " + strconv.Itoa(t.Qos) + "}"
}

func (e Event) String() string {
	var sb strings.Builder

	sb.WriteString("Event: { ")
	sb.WriteString("clientId: \"" + e.ClientId + "\", ")
	sb.WriteString("type:\"" + e.PacketType.String() + "\", ")
	sb.WriteString("messageId: \"" + strconv.Itoa(int(e.MessageId)) + "\", ")
	sb.WriteString("topic: " + e.Topic.String() + ", ")
	sb.WriteString("payload: \"" + e.Payload + "\", ")
	sb.WriteString("qos: " + strconv.Itoa(e.Qos) + ", ")
	sb.WriteString("retain: " + strconv.FormatBool(e.Retain) + ", ")
	sb.WriteString("dublicate: " + strconv.FormatBool(e.Dublicate) + "}")

	return sb.String()
}
