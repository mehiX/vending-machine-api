package app

import (
	"context"
	"errors"
	"fmt"

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
		return
	}

	return a.dbCreateUser(ctx, username, string(encPasswd), deposit, role)
}

func (a *app) FindUserByCredentials(ctx context.Context, username, password string) (*model.User, error) {

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

func (a *app) FindUserByID(ctx context.Context, id string) (*model.User, error) {
	return a.dbFindUserByID(ctx, id)
}

func (a *app) ResetDeposit(ctx context.Context, usr *model.User) error {
	return a.dbUserUpdateDeposit(ctx, usr.ID, 0)
}

func (a *app) UserDepositCoin(ctx context.Context, usr *model.User, coin int) error {

	if err := validateDepositCoin(coin); err != nil {
		return errors.New("coin value not allowed")
	}

	if err := a.dbUserUpdateDeposit(ctx, usr.ID, usr.Deposit+int64(coin)); err != nil {
		fmt.Println("deposit error", err)
		return errors.New("deposit failed")
	}

	return nil
}
