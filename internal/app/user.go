package app

import (
	"errors"

	"github.com/mehiX/vending-machine-api/internal/app/model"
)

func (a *app) CreateUser(username, password string, deposit int64, role model.TypeRole) error {
	return errors.New("not implemented")
}
