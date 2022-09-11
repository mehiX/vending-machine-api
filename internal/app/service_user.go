package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/mehiX/vending-machine-api/internal/app/model"
	"golang.org/x/crypto/bcrypt"
)

const password_min_length = 8

func (a *App) CreateUser(ctx context.Context, username, password string, deposit int64, role model.TypeRole) (err error) {

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

func (a *App) FindUserByCredentials(ctx context.Context, username, password string) (*model.User, error) {

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

func (a *App) FindUserByID(ctx context.Context, id string) (*model.User, error) {
	return a.dbFindUserByID(ctx, id)
}

func (a *App) ResetDeposit(ctx context.Context, usr *model.User) error {
	return a.dbUserUpdateDeposit(ctx, usr.ID, 0)
}

func (a *App) UserDepositCoin(ctx context.Context, usr *model.User, coin int) error {

	if err := validateDepositCoin(coin); err != nil {
		return errors.New("coin value not allowed")
	}

	if err := a.dbUserUpdateDeposit(ctx, usr.ID, usr.Deposit+int64(coin)); err != nil {
		fmt.Println("deposit error", err)
		return errors.New("deposit failed")
	}

	return nil
}

func (a *App) Buy(ctx context.Context, user *model.User, prod *model.Product, amount int) error {
	if amount > int(prod.AmountAvailable) {
		return errors.New("no availability")
	}

	if user.Deposit < int64(amount)*prod.Cost {
		return errors.New("not enough deposit")
	}

	return a.dbBuy(ctx, user.ID, prod.ID, amount, int64(amount)*prod.Cost)
}

// getChange splits the amount `n` in coins of 5, 10, 20, 50, 100
// It assumes that the vending machine has all types of coins available at any time, in sufficient amounts
// The amount `n` should be a multiple of 5.
func getChange(n int64) [5]int64 {

	values := []int64{5, 10, 20, 50, 100}
	coins := [5]int64{}

	remaining := n
	for i := 4; i >= 0; i-- {
		coins[i] = remaining / values[i]
		remaining -= coins[i] * values[i]
	}

	return coins
}
