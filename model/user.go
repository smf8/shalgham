package model

import (
	"errors"

	"github.com/jinzhu/gorm"
)

var ErrUserNotFound = errors.New("user not found")

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	IsOnline bool   `json:"is_online"`
}

type UserRepo interface {
	Save(user User) error
	FindUser(username string) (*User, error)
	Disconnect(username string) error
	Connect(username string) error
}

type SQLUserRepo struct {
	DB *gorm.DB
}

func (s SQLUserRepo) Save(user User) error {
	return s.DB.Create(&user).Error
}

func (s SQLUserRepo) FindUser(username string) (*User, error) {
	user := &User{}

	if err := s.DB.Where("username = ?", username).Find(user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	return user, nil
}

func (s SQLUserRepo) Disconnect(username string) error {
	return s.DB.Model(&User{}).Where("username = ?", username).Update("is_online", false).Error
}

func (s SQLUserRepo) Connect(username string) error {
	return s.DB.Model(&User{}).Where("username = ?", username).Update("is_online", "true").Error
}
