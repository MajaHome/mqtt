package model

import "gorm.io/gorm"

type RetainMessage struct {
	gorm.Model
	MessageId uint16 `gorm:"uniqueIndex"`
	Topic     string
	Payload   string
	Qos       int
}
