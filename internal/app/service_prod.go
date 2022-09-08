package app

import (
	"context"
	"errors"
	"fmt"
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

func (a *app) ListProducts(ctx context.Context) ([]model.Product, error) {

	if a.Db == nil {
		return nil, errors.New("no db conn")
	}

	conn, err := a.Db.Conn(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := conn.QueryContext(ctx, `select id, name, available_amount, cost, seller_id from products`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]model.Product, 0)

	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.AmountAvailable, &p.Cost, &p.SellerID); err != nil {
			fmt.Println("product record error", err)
			continue
		}
		products = append(products, p)
	}

	return products, nil
}
