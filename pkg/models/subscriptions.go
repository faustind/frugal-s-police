package models

import (
	"database/sql"
)

type Subscription struct {
	UserID       string `gorm:"PrimaryKey"`
	Name         string `gorm:"PrimaryKey"`
	Cost         int
	DueDay       int
	LastPayMonth int
}

func (s *Subscription) CreateSubscription() (*Subscription, error) {
	err := db.Create(&s).Error
	return s, err
}

func GetSubscription(userId, serviceName string) (*Subscription, error) {
	var subscription *Subscription
	err := db.Where("user_id = @id AND name = @name", sql.Named("name", serviceName), sql.Named("id", userId)).Find(&subscription).Error
	return subscription, err
}

func GetAllSubscriptionsByUser(userId string) ([]Subscription, error) {
	var subscriptions []Subscription
	err := db.Where("user_id = @id", sql.Named("id", userId)).Find(&subscriptions).Error
	return subscriptions, err
}

func (s *Subscription) UpdateSubscription() (*Subscription, error) {
	//err := db.Where("user_id = @id AND name = @name", sql.Named("name", s.Name), sql.Named("id", s.UserID)).Save(&s).Error
	err := db.Save(s).Error
	return s, err
}

func DeleteSubscription(userId, serviceName string) (Subscription, error) {
	var s Subscription
	err := db.Where("user_id = @id AND name = @name", sql.Named("name", serviceName), sql.Named("id", userId)).Delete(s).Error
	return s, err
}
