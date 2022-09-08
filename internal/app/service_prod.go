package app

import (
	"context"
	"errors"
	"strings"

	"github.com/mehiX/vending-machine-api/internal/app/model"
)

// CreateProduct validates the input parameters and calls the repository to store the data in the database
// Only users with the role `SELLER` can create objects
func (a *app) CreateProduct(ctx context.Context, seller *model.User, amountAvailable int64, cost int64, name string) (err error) {

	if seller == nil {
		return errors.New("missing seller")
	}

	if seller.Role != model.ROLE_SELLER {
		return errors.New("user is not a seller")
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

func (a *app) DeleteProduct(ctx context.Context, seller *model.User, product *model.Product) (err error) {

	if seller == nil {
		return errors.New("user is nil")
	}

	if product == nil {
		return errors.New("product is nil")
	}

	if seller.Role != model.ROLE_SELLER {
		return errors.New("user is not a seller")
	}

	if seller.ID != product.SellerID {
		return errors.New("wrong seller id")
	}

	if a.Db == nil {
		return errors.New("no db conn")
	}

	return a.dbDeleteProduct(ctx, product.ID, seller.ID)
}
