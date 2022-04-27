package models

type User struct {
	ID            string `gorm:"primaryKey"`
	Budget        int
	Subscriptions []Subscription
}

func (u *User) CreateUser() (*User, error) {
	err := db.Create(&u).Error
	return u, err
}

func GetAllUsers() ([]User, error) {
	var users []User
	err := db.Find(&users).Error
	return users, err
}

func GetUserById(Id string) (*User, error) {
	var user *User
	err := db.Where("id=?", Id).Find(&user).Error
	return user, err
}

func DeleteUser(Id string) (User, error) {
	var user User
	// remove user and her subscription from db
	err := db.Select("Subscriptions").Where("id = ?", Id).Delete(user).Error
	return user, err
}

func (u *User) UpdateUser() (*User, error) {
	err := db.Save(u).Error
	return u, err
}
