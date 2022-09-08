package app

import (
	"context"
	"errors"

	"github.com/mehiX/vending-machine-api/internal/app/model"
	"golang.org/x/crypto/bcrypt"
)

const password_min_length = 8

func (a *app) CreateUser(ctx context.Context, username, password string, deposit int64, role model.TypeRole) (err error) {

	if err = validateUsername(username); err != nil {
		return
	}

	if err = validatePassword(password); err != nil {
		return
	}

	if err = validateDeposit(deposit); err != nil {
		return
	}

	if err = validateRole(role); err != nil {
		return
	}

	encPasswd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return err
	}

	if a.Db == nil {
		return errors.New("no database configured")
	}

	return a.dbCreateUser(ctx, username, string(encPasswd), deposit, role)
}

func (a *app) FindByCredentials(ctx context.Context, username, password string) (*model.User, error) {

	usr, err := a.dbFindUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(usr.Password), []byte(password)); err != nil {
		return nil, errors.New("credentials don't match")
	}

	usr.Password = ""

	return usr, nil
}
