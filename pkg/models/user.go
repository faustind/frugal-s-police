package models

type User struct {
	ID            string         `gorm:"primaryKey"`
	Budget        int            `json:"budget"`
	Subscriptions []Subscription `json:"subscriptions"`
}

func (u *User) CreateUser() *User {
	db.Create(&u)
	return u
}

func DeleteUser(Id string) User {
	var user User
	user.ID = Id
	// remove user and her subscription from db
	db.Select("Subscriptions").Delete(user)
	return user
}
