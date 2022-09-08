package app

import (
	"context"
	"errors"

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

	_, err = tx.ExecContext(ctx, "insert into users (id, username, password, deposit, role) values (?, ?, ?, ?, ?)", uuid.New().String(), username, encPasswd, deposit, role)
	if err != nil {
		return err
	}

	return

}

func (a *app) dbFindOneByUsername(ctx context.Context, username string) (*model.User, error) {

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
	// TODO: implement
	return nil, errors.New("not implemented")
}
