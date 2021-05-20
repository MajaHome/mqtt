package model

import "github.com/jinzhu/gorm"

type Subscription struct {
	gorm.Model


	// todo
	UserName string `gorm:"uniqueIndex"`
	Password string
}
