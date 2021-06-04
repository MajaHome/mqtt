package model

import "gorm.io/gorm"

type Subscription struct {
	gorm.Model
	ClientId string
	Topic    string
	Qos      int
}
