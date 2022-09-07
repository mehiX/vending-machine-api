package app

import (
	"log"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var users []User

func NewUser() User {
	return User{
		ID: uuid.New().String(),
	}
}

func (a *app) UserAdd(username, password string, role TypeRole) (string, error) {
	if userExists(username) {
		return "", ErrUsernameExists
	}

	if !validPassword(password) {
		return "", ErrInvalidPassword
	}

	usr := NewUser()
	usr.Username = username
	pwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return "", err
	}
	usr.Password = string(pwd)
	usr.Role = role

	users = append(users, usr)

	return usr.ID, nil
}

func userExists(username string) bool {
	for _, u := range users {
		if u.Username == username {
			return true
		}
	}

	return false
}

func validPassword(p string) bool {
	return len(p) >= 8
}

func comparePasswords(hashedPwd string, plainPwd []byte) bool {
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPwd)
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}
