package model

import (
	"github.com/jinzhu/gorm"
)

type Subscription struct {
	gorm.Model
	ClientId	string
	Topic		string
	Qos			int
}
