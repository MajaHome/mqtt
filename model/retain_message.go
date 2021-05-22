package model

import "github.com/jinzhu/gorm"

type RetainMessage struct {
	gorm.Model
	MessageId uint16 `gorm:"uniqueIndex"`
	Topic     string
	Payload   string
	Qos       int
}
