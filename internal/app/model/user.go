package model

import (
	"errors"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"golang.org/x/crypto/bcrypt"
)

// User ...
type User struct {
	ID                int    `json:"id"`
	Username          string `json:"username"`
	EncryptedPassword string `json:"-"`
	Password          string `json:"-"`
}

// EncryptPassword ...
func (user *User) EncryptPassword() error {
	if len(user.Password) == 0 {
		return errors.New("password must be set")
	}
	encrypedPassword, err := encryptString(user.Password)
	if err != nil {
		return err
	}
	user.EncryptedPassword = encrypedPassword
	return nil
}

// Validate ...
func (user *User) Validate() error {
	return validation.ValidateStruct(user,
		validation.Field(&user.Username, validation.Required, is.Alpha),
		validation.Field(&user.Password,
			validation.By(requiredIf(user.EncryptedPassword != "")),
			validation.Length(6, 120)))
}

func encryptString(plainString string) (string, error) {
	encryptedString, err := bcrypt.GenerateFromPassword([]byte(plainString), bcrypt.MinCost)
	if err != nil {
		return "", err
	}
	return string(encryptedString), nil
}

// CheckPassword ...
func (user *User) CheckPassword(password string) bool {
	return bcrypt.CompareHashAndPassword(
		[]byte(user.EncryptedPassword),
		[]byte(password)) == nil
}
