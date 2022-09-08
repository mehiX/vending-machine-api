package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/mehiX/vending-machine-api/internal/app/model"
)

// dbCreateUser receives sanitized data and tries to create a new database record
func (a *app) dbCreateUser(ctx context.Context, username, encPasswd string, deposit int64, role string) (err error) {
	tx, err := a.Db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	qryUser := "insert into users (id, username, password, deposit, role) values (?, ?, ?, ?, ?)"

	_, err = tx.ExecContext(ctx, qryUser, uuid.New().String(), username, encPasswd, deposit, role)
	if err != nil {
		return err
	}

	return

}

func (a *app) dbCreateProduct(ctx context.Context, sellerID string, amountAvailable int64, cost int64, name string) (err error) {
	tx, err := a.Db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	qryProd := "insert into products (id, name, amountAvailable, cost, seller_id) values (?, ?, ?, ?, ?)"

	_, err = tx.ExecContext(ctx, qryProd, uuid.New().String(), name, amountAvailable, cost, sellerID)
	if err != nil {
		return err
	}

	return

}

func (a *app) dbFindUserByID(ctx context.Context, userID string) (*model.User, error) {

	conn, err := a.Db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var usr model.User

	row := conn.QueryRowContext(ctx, `select id, username, password, deposit, role from users where id=?`, userID)
	if err := row.Scan(&usr.ID, &usr.Username, &usr.Password, &usr.Deposit, &usr.Role); err != nil {
		return nil, err
	}

	return &usr, nil
}

func (a *app) dbFindUserByUsername(ctx context.Context, username string) (*model.User, error) {

	conn, err := a.Db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var usr model.User

	row := conn.QueryRowContext(ctx, `select id, username, password, deposit, role from users where username=?`, username)
	if err := row.Scan(&usr.ID, &usr.Username, &usr.Password, &usr.Deposit, &usr.Role); err != nil {
		return nil, err
	}

	return &usr, nil
}

func (a *app) dbFindProductByID(ctx context.Context, productID string) (*model.Product, error) {

	conn, err := a.Db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var prod model.Product

	qryProdByID := `select id, name, available_amount, count, seller_id from products where id=?`

	row := conn.QueryRowContext(ctx, qryProdByID, productID)
	if err := row.Scan(&prod.ID, &prod.Name, &prod.AmountAvailable, &prod.Cost, &prod.SellerID); err != nil {
		return nil, err
	}

	return &prod, nil
}
