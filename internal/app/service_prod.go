package app

import (
	"context"
	"errors"

	"github.com/mehiX/vending-machine-api/internal/app/model"
)

func (a *app) CreateProduct(ctx context.Context, seller *model.User, amountAvailable int64, cost int64, name string) (err error) {

	return errors.New("not implemented")
}
