package model

import (
	"fmt"
	"gorm.io/gorm"
)

type RetainMessage struct {
	gorm.Model
	MessageId uint16 `gorm:"uniqueIndex"`
	Topic     string
	Payload   string
	Qos       int
}

func (m RetainMessage) String() string {
	return fmt.Sprintf("retain message {id: %d, topic:%s, payload:%s, qos:%d}",
		m.MessageId, m.Topic, m.Payload, m.Qos)
}
