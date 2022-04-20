package models

import (
	"gorm.io/gorm"
)

type Subscription struct {
	gorm.Model
	UserID    string
	Name      string `json:"name"`
	Cost      int    `json:"cost"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

func (s *Subscription) CreateSubscription() *Subscription {
	db.Create(&s)
	return s
}
