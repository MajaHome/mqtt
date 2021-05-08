package server

import (
	"mqtt/packet"
	"strconv"
	"strings"
)

type Event struct {
	clientId   string
	packetType packet.Type
	topics     []EventTopic
	messageId  uint16
}

type EventTopic struct {
	qos int
	topic string
}

func (e Event) String() string {
	var sb strings.Builder

	sb.WriteString("Event: { ")
	sb.WriteString("clientId: \"" + e.clientId +"\",")
	sb.WriteString("type:\"" + e.packetType.String() + "\",")
	sb.WriteString("messageId: \"" + strconv.Itoa(int(e.messageId)) + "\",")
	sb.WriteString("topics: [")
	for _, t := range e.topics {
		sb.WriteString("{ qos: " + strconv.Itoa(t.qos) + ", topic: \"" + t.topic + "\" }")
	}
	sb.WriteString("]}")

	return sb.String()
}
