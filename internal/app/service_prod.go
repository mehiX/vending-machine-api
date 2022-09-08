package app

import (
	"context"
	"errors"
	"strings"

	"github.com/mehiX/vending-machine-api/internal/app/model"
)

func (a *app) CreateProduct(ctx context.Context, seller *model.User, amountAvailable int64, cost int64, name string) (err error) {

	if seller == nil {
		return errors.New("missing seller")
	}

	if amountAvailable <= 0 {
		return errors.New("available amount must be positive")
	}

	if err = validateCost(cost); err != nil {
		return
	}

	if strings.TrimSpace(name) == "" {
		return errors.New("missing name for product")
	}

	if a.Db == nil {
		return errors.New("no database to save the product")
	}

	return a.dbCreateProduct(ctx, seller.ID, amountAvailable, cost, strings.TrimSpace(name))
}
