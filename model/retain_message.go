package model

import (
	"fmt"

	"gorm.io/gorm"
)

type RetainMessage struct {
	gorm.Model
	Topic   string
	Payload string
	Qos     int
}

func (m RetainMessage) String() string {
	return fmt.Sprintf("retain message {topic:%s, payload:%s, qos:%d}", m.Topic, m.Payload, m.Qos)
}
