package config

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db  *gorm.DB
	dsn string
	err error
)

func Connect() {
	if os.Getenv("ENV") == "PROD" {
		dsn := os.Getenv("DATABASE_URL")
		db, err = gorm.Open(postgres.New(postgres.Config{DSN: dsn}), &gorm.Config{
			Logger: logger.Default,
		})
	} else {
		dsn = "host=db user=admin password=admin dbname=aubipo port=5432 sslmode=disable TimeZone=Asia/Tokyo"
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default,
		})
	}
	if err != nil {
		log.Fatalf("failed to connect database: %s", err)
	}

}

func GetDB() *gorm.DB {
	return db
}
