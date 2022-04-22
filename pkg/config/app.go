package config

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db *gorm.DB
)

func Connect() {
	d, err := gorm.Open(sqlite.Open("aubipo.db"), &gorm.Config{
		Logger: logger.Default,
	})
	if err != nil {
		log.Fatalf("failed to connect database: %s", err)
	}

	db = d
}

func GetDB() *gorm.DB {
	return db
}
