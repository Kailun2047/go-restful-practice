package models

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/badoux/checkmail"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Action = string

const (
	Update = "update"
	Login  = "login"
)

type User struct {
	ID        uint      `json:"id"`
	Name      string    `gorm:"255;not null;unique" json:"name"`
	Email     string    `gorm:"100;not null;unique" json:"email"`
	Password  string    `gorm:"100;not null" json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func Hash(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func CheckPassword(hash []byte, password []byte) error {
	return bcrypt.CompareHashAndPassword(hash, password)
}

func (u *User) Validate(action Action) error {
	switch strings.ToLower(action) {
	case Update:
		if u.Name == "" {
			return fmt.Errorf("Name is a required field")
		}
		if u.Password == "" {
			return fmt.Errorf("Password is a required field")
		}
		if u.Email == "" {
			return fmt.Errorf("Email is a required field")
		}
		if err := checkmail.ValidateFormat(u.Email); err != nil {
			return fmt.Errorf("Invalid Email format: [%v]", err)
		}
		return nil
	case Login:
		if u.Password == "" {
			return fmt.Errorf("Password is a required field")
		}
		if u.Email == "" {
			return fmt.Errorf("Email is a required field")
		}
		if err := checkmail.ValidateFormat(u.Email); err != nil {
			return fmt.Errorf("Invalid Email format: [%v]", err)
		}
		return nil
	default:
		return fmt.Errorf("Unsupported action [%s]", action)
	}
}

func (u *User) BeforeSave() error {
	hash, err := Hash(u.Password)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return nil
}

func (u *User) SaveUser(db *gorm.DB) (*User, error) {
	err := db.Debug().Model(&User{}).Create(u).Error
	if err != nil {
		return &User{}, err
	}
	return u, nil
}

func (u *User) FindAllUsers(db *gorm.DB) (*[]User, error) {
	users := []User{}
	err := db.Debug().Model(&User{}).Find(&users).Error
	if err != nil {
		return &[]User{}, err
	}
	return &users, nil
}

func (u *User) FindUserById(db *gorm.DB, id uint) (*User, error) {
	err := db.Debug().Model(&User{}).Where("id = ?", id).First(u).Error
	if err == nil {
		return u, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &User{}, fmt.Errorf("User with ID [%d] not found", id)
	}
	return &User{}, err
}

func (u *User) UpdateUser(db *gorm.DB, id uint) (*User, error) {
	err := u.BeforeSave()
	if err != nil {
		return &User{}, err
	}

	db = db.Debug().Model(&User{}).First(&User{}).Updates(
		map[string]interface{}{
			"name":       u.Name,
			"password":   u.Password,
			"email":      u.Email,
			"updated_at": time.Now(),
		},
	)
	if err := db.Error; err != nil {
		return &User{}, err
	}
	// Retrieve the updated user.
	err = db.Debug().Where("id = ?", id).First(u).Error
	if err != nil {
		return &User{}, err
	}
	return u, err
}

func (u *User) DeleteUser(db *gorm.DB, id uint) (int64, error) {
	db = db.Debug().Model(&User{}).Where("id = ?", id).Delete(&User{})
	if err := db.Error; err != nil {
		return 0, err
	}
	return db.RowsAffected, nil
}
