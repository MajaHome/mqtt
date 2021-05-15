package server

import (
	"mqtt/packet"
	"strconv"
	"strings"
)

type Event struct {
	clientId   string
	packetType packet.Type
	messageId  uint16
	topic      EventTopic
	payload    string
	qos        int
	retain     bool
	dublicate  bool
}

type EventTopic struct {
	name string
	qos  int
}

func (t *EventTopic) String() string {
	return "{ name: " + t.name + ", qos: " + strconv.Itoa(t.qos) + "}"
}

func (e Event) String() string {
	var sb strings.Builder

	sb.WriteString("Event: { ")
	sb.WriteString("clientId: \"" + e.clientId + "\", ")
	sb.WriteString("type:\"" + e.packetType.String() + "\", ")
	sb.WriteString("messageId: \"" + strconv.Itoa(int(e.messageId)) + "\", ")
	sb.WriteString("topic: " + e.topic.String() + ", ")
	sb.WriteString("payload: \"" + e.payload + "\", ")
	sb.WriteString("qos: " + strconv.Itoa(e.qos) + ", ")
	sb.WriteString("retain: " + strconv.FormatBool(e.retain) + ", ")
	sb.WriteString("dublicate: " + strconv.FormatBool(e.dublicate) + "}")

	return sb.String()
}
