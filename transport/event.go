package transport

import (
	"fmt"
	"github.com/MajaSuite/mqtt/packet"
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
	Restore    bool
}

type EventTopic struct {
	Name string
	Qos  int
}

func (t *EventTopic) String() string {
	return fmt.Sprintf(`{"name":"%s","qos":%d}`, t.Name, t.Qos)
}

func (e Event) String() string {
	return fmt.Sprintf(`{"clientid","%s","type","%s","messageid":%d,"topic":%v,"payload","%s","qos":%d,"retain":%v,"dup":%v}`,
		e.ClientId, e.PacketType, e.MessageId, e.Topic.String(), e.Payload, e.Qos, e.Retain, e.Dublicate)
}
