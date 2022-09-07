package app

import (
	"context"

	"github.com/google/uuid"
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
