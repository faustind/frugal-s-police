package models

import (
	"faustind/aubipo/pkg/config"

	"gorm.io/gorm"
)

var db *gorm.DB

func init() {
	config.Connect()
	db = config.GetDB()
	db.AutoMigrate(&User{}, &Subscription{})
}
