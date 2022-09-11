package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mehiX/vending-machine-api/internal/app/model"

	_ "github.com/go-sql-driver/mysql"
)

// ConnectDB tries to establish a database connection.
// Retries periodically to check that the connection is still available.
// Should be run in a separate goroutine.
func (a *app) ConnectDB(done context.Context, connStr string, pingDelay time.Duration) {

	test := func(db *sql.DB) error {
		ctx, cancel := context.WithTimeout(done, 2*time.Second)
		defer cancel()
		return db.PingContext(ctx)
	}

	// don't fill the logs if connection is OK
	var printConnOK bool = true

	tkr := time.NewTicker(pingDelay)
	for {
		select {
		case <-done.Done():
			if a.Db != nil {
				a.Db.Close()
			}
			fmt.Println("DB connection closed")
			return
		case <-tkr.C:
			if a.Db == nil {
				// try to connect
				fmt.Println("DB: connecting...")
				db, err := sql.Open("mysql", connStr)
				if err != nil {
					fmt.Printf("DB: %v\n", err.Error())
				} else {
					db.SetConnMaxLifetime(0)
					db.SetMaxIdleConns(50)
					db.SetMaxOpenConns(50)

					if err := test(db); err == nil {
						a.Db = db
					}
				}
			} else {
				// check if server still available
				if err := test(a.Db); err != nil {
					fmt.Printf("DB: Ping %v\n", err.Error())
					a.Db = nil
					printConnOK = true
				} else {
					if printConnOK {
						fmt.Println("DB: connection OK")
						printConnOK = false
					}
				}
			}
		}
	}
}

// dbCreateUser receives sanitized data and tries to create a new database record
func (a *app) dbCreateUser(ctx context.Context, username, encPasswd string, deposit int64, role string) (err error) {

	if a.Db == nil {
		return errors.New("no database configured")
	}

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

	qryProd := "insert into products (id, name, available_amount, cost, seller_id) values (?, ?, ?, ?, ?)"

	_, err = tx.ExecContext(ctx, qryProd, uuid.New().String(), name, amountAvailable, cost, sellerID)
	if err != nil {
		return err
	}

	return

}

func (a *app) dbFindUserByID(ctx context.Context, userID string) (*model.User, error) {

	if a.Db == nil {
		return nil, errors.New("no database configured")
	}

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

func (a *app) dbDeleteProduct(ctx context.Context, productID, sellerID string) (err error) {

	tx, err := a.Db.BeginTx(ctx, nil)
	if err != nil {
		return
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	qryDelProd := `delete from products where id=? and seller_id=?`

	res, err := tx.ExecContext(ctx, qryDelProd, productID, sellerID)
	if err != nil {
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return
	}

	if rowsAffected < 1 {
		err = errors.New("no rows deleted")
	}

	return
}

func (a *app) dbListProducts(ctx context.Context) ([]model.Product, error) {
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

func (a *app) dbUpdateProduct(ctx context.Context, p model.Product) (err error) {

	tx, err := a.Db.BeginTx(ctx, nil)
	if err != nil {
		return
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	qryUpdProd := `update products set name=?, cost=? where id=? and seller_id=?`

	_, err = tx.ExecContext(ctx, qryUpdProd, p.Name, p.Cost, p.ID, p.SellerID)

	return
}

func (a *app) dbUserUpdateDeposit(ctx context.Context, userID string, newDeposit int64) (err error) {

	if a.Db == nil {
		return errors.New("no database configured")
	}

	tx, err := a.Db.BeginTx(ctx, nil)
	if err != nil {
		return
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	qryUpdateDeposit := `update users set deposit=? where id=?`

	_, err = tx.ExecContext(ctx, qryUpdateDeposit, newDeposit, userID)

	return

}

// dbBuy implements the buy logic at the database level
// this would be better implemented in a stored procedure
func (a *app) dbBuy(ctx context.Context, userID, prodID string, amount int, totalCost int64) (err error) {

	if a.Db == nil {
		return errors.New("no database configured")
	}

	tx, err := a.Db.BeginTx(ctx, nil)
	if err != nil {
		return
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `update products set available_amount = available_amount - ? where id=?`, amount, prodID); err != nil {
		return
	}

	if _, err = tx.ExecContext(ctx, `update users set deposit = deposit - ? where id=?`, totalCost, userID); err != nil {
		return
	}

	// TODO: extra checks if resting available_amount < 0 or deposit < 0
	// the way these situations are handled in a concurrent environment
	// depend a lot on the database as well (how transactions work, isolation, snapshots, etc)

	return

}
