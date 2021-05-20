package model

import "github.com/jinzhu/gorm"

type RetainMessage struct {
	gorm.Model

	// TODO
	UserName string `gorm:"uniqueIndex"`
	Password string
}
