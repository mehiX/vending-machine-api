package app

import (
	"errors"
	"regexp"

	"github.com/mehiX/vending-machine-api/internal/app/model"
	"golang.org/x/crypto/bcrypt"
)

func (a *app) CreateUser(username, password string, deposit int64, role model.TypeRole) error {

	if err := validateUsername(username); err != nil {
		return err
	}

	if err := validatePassword(password); err != nil {
		return err
	}

	_, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MaxCost)
	if err != nil {
		return err
	}

	return errors.New("not implemented")
}

func validateUsername(u string) error {
	p := regexp.MustCompile(`[0-9a-zA-Z@._-]{8,}`)
	rest := p.ReplaceAllLiteralString(u, "")

	if rest != "" {
		return errors.New("username should be at least 8 characters long and may only container letters, numbers and one or more symbols @ . _ -")
	}

	return nil
}

func validatePassword(p string) error {
	return errors.New("not implemented")
}
