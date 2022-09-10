package app

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mehiX/vending-machine-api/internal/app/model"
)

func TestCreateUserShouldSucceed(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`insert into users \(id, username, password, deposit, role\) values`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	vm := NewApp("", db)
	if err := vm.dbCreateUser(context.Background(), "mihaiuser", "strong23Pass*", 100, model.ROLE_BUYER); err != nil {
		t.Fatal(err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there are unfulfilled expectations: %s", err)
	}
}

func TestDbCreateUserShouldRollbackOnError(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`insert into users`).WillReturnError(fmt.Errorf("some db error"))
	mock.ExpectRollback()

	vm := NewApp("", db)
	if err := vm.dbCreateUser(context.Background(), "mihaiuser", "strong23Pass*", 100, model.ROLE_BUYER); err == nil {
		t.Fatal(errors.New("database error should be returned by the function"))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there are unfulfilled expectations: %s", err)
	}
}

func TestDbFindUserByUsernameSuccess(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	columns := []string{"id", "username", "password", "deposit", "role"}

	mock.ExpectQuery("select id, username, password, deposit, role from users where username=").
		WithArgs("mihaiusr").WillReturnRows(sqlmock.NewRows(columns).
		AddRow("id1", "mihaiusr", "bcryptksdafjsafkj", 100, "BUYER"))

	vm := NewApp("", db)
	usr, err := vm.dbFindUserByUsername(context.Background(), "mihaiusr")
	if err != nil {
		t.Fatal(err)
	}

	if usr.ID != "id1" {
		t.Errorf("wrong id. Expected: %s, got: %s", "id1", usr.ID)
	}

	if usr.Username != "mihaiusr" {
		t.Errorf("wrong username. Expected: %s, got: %s", "mihaiusr", usr.Username)
	}

	if usr.Password != "bcryptksdafjsafkj" {
		t.Errorf("wrong encrypted password. Expected: %s, got: %s", "bcryptksdafjsafkj", usr.Password)
	}

	if usr.Deposit != 100 {
		t.Errorf("wrong deposit. Expected: %d, got: %d", 100, usr.Deposit)
	}

	if usr.Role != model.ROLE_BUYER {
		t.Errorf("wrong role. Expected: %s, got: %s", model.ROLE_BUYER, usr.Role)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there are unfulfilled expectations: %s", err)
	}
}

func TestDbFindUserByUsernameNoMatch(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery("select id, username, password, deposit, role from users where username=").
		WithArgs("mihaiusr").WillReturnError(errors.New("no records found"))

	vm := NewApp("", db)
	_, err = vm.dbFindUserByUsername(context.Background(), "mihaiusr")
	if err == nil {
		t.Fatal("no user found for username and no error returned")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there are unfulfilled expectations: %s", err)
	}
}

func TestDbFindUserByIDSuccess(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	columns := []string{"id", "username", "password", "deposit", "role"}

	mock.ExpectQuery("select id, username, password, deposit, role from users where id=").
		WithArgs("id1").WillReturnRows(sqlmock.NewRows(columns).
		AddRow("id1", "mihaiusr", "bcryptksdafjsafkj", 100, "BUYER"))

	vm := NewApp("", db)
	usr, err := vm.dbFindUserByID(context.Background(), "id1")
	if err != nil {
		t.Fatal(err)
	}

	if usr.ID != "id1" {
		t.Errorf("wrong id. Expected: %s, got: %s", "id1", usr.ID)
	}

	if usr.Username != "mihaiusr" {
		t.Errorf("wrong username. Expected: %s, got: %s", "mihaiusr", usr.Username)
	}

	if usr.Password != "bcryptksdafjsafkj" {
		t.Errorf("wrong encrypted password. Expected: %s, got: %s", "bcryptksdafjsafkj", usr.Password)
	}

	if usr.Deposit != 100 {
		t.Errorf("wrong deposit. Expected: %d, got: %d", 100, usr.Deposit)
	}

	if usr.Role != model.ROLE_BUYER {
		t.Errorf("wrong role. Expected: %s, got: %s", model.ROLE_BUYER, usr.Role)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there are unfulfilled expectations: %s", err)
	}
}

func TestDbFindUserByIDNoMatch(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery("select id, username, password, deposit, role from users where id=").
		WithArgs("id1").WillReturnError(errors.New("no records found"))

	vm := NewApp("", db)
	_, err = vm.dbFindUserByID(context.Background(), "id1")
	if err == nil {
		t.Fatal("no user found for username and no error returned")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there are unfulfilled expectations: %s", err)
	}
}

func TestDbCreateUserNoDb(t *testing.T) {

	errMsg := "no database configured"

	if err := NewApp("", nil).dbCreateUser(context.Background(), "", "", 10, ""); err == nil {
		t.Fatal("should fail if DB cannot open a TX")
	} else {
		if err.Error() != errMsg {
			t.Fatalf("wrong error message. expected: %s, got: %s", errMsg, err.Error())
		}
	}
}

func TestDbCreateUserCannotStartTx(t *testing.T) {

	errMsg := "cannot open tx"

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin().WillReturnError(errors.New(errMsg))

	if err := NewApp("", db).dbCreateUser(context.Background(), "", "", 10, ""); err == nil {
		t.Fatal("should fail if DB cannot open a TX")
	} else {
		if err.Error() != errMsg {
			t.Fatalf("wrong error message. expected: %s, got: %s", errMsg, err.Error())
		}
	}
}

func TestDbCreateProductFailDbError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`insert into products \(`).WillReturnError(errors.New("duplicate entry - same name"))
	mock.ExpectRollback()

	if err := NewApp("", db).dbCreateProduct(context.Background(), "sellerid", 10, 10, "product name"); err == nil {
		t.Error("should return error if the insert failed")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDbFindUserByIDFailNoConnection(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	db.Close()

	if _, err := NewApp("", db).dbFindUserByID(context.Background(), "userid"); err == nil {
		t.Fatal("should fail if db connection already closed")
	}
}

func TestDbFindUserByUsernameFailNoConnection(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	db.Close()

	if _, err := NewApp("", db).dbFindUserByUsername(context.Background(), "username"); err == nil {
		t.Fatal("should fail if db connection already closed")
	}
}

func TestDbFindProductByIDFailNoConnection(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	db.Close()

	if _, err := NewApp("", db).dbFindProductByID(context.Background(), "productid"); err == nil {
		t.Fatal("should fail if db connection already closed")
	}
}

func TestDbFindProductByIDFailQuery(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`select .* from products where id=`).WithArgs("productid").WillReturnError(errors.New("no records found"))

	_, err = NewApp("", db).dbFindProductByID(context.Background(), "productid")
	if err == nil {
		t.Fatal("should return error if the query fails")
	}
}

func TestDbFindProductByIDSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	columns := []string{"id", "name", "available_amount", "cost", "seller_id"}

	mock.ExpectQuery(`select .* from products where id=`).WithArgs("productid").WillReturnRows(sqlmock.NewRows(columns).AddRow("productid", "product name", 10, 5, "seller 1"))

	prod, err := NewApp("", db).dbFindProductByID(context.Background(), "productid")
	if err != nil {
		t.Fatal(err)
	}

	if prod.ID != "productid" {
		t.Errorf("wrong product id. expected: %s, got: %s", "productid", prod.ID)
	}

	if prod.Name != "product name" {
		t.Errorf("wrong product name. expected: %s, got: %s", "product name", prod.Name)
	}

	if prod.AmountAvailable != 10 {
		t.Errorf("wrong amount available. expected: %d, got: %d", 10, prod.AmountAvailable)
	}

	if prod.Cost != 5 {
		t.Errorf("wrong cost. expected: %d, got: %d", 5, prod.Cost)
	}

	if prod.SellerID != "seller 1" {
		t.Errorf("wrong seller id. expected: %s, got: %s", "seller 1", prod.SellerID)
	}
}

func TestDbDeleteProductTxError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin().WillReturnError(errors.New("tx could not be created"))

	if err := NewApp("", db).dbDeleteProduct(context.Background(), "", ""); err == nil {
		t.Fatal("should exit since tx begin failed")
	} else {
		if err.Error() != "tx could not be created" {
			t.Fatal("wrong error text")
		}
	}
}

func TestDbDeleteProductNoRowsDeleted(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`delete from products`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	if err := NewApp("", db).dbDeleteProduct(context.Background(), "", ""); err == nil {
		t.Fatal("should exit with error since no rows were deleted")
	} else {
		if err.Error() != "no rows deleted" {
			t.Fatal("wrong error text")
		}
	}
}

func TestDbUpdateProductFailNoTx(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin().WillReturnError(errors.New("no transaction"))

	if err := NewApp("", db).dbUpdateProduct(context.Background(), model.Product{}); err == nil {
		t.Fatal("should fail if no TX")
	} else {
		if err.Error() != "no transaction" {
			t.Fatal("should return the database error")
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDbUpdateProductFailQueryError(t *testing.T) {

	prod := model.Product{
		ID:              "id",
		Name:            "name",
		AmountAvailable: 10,
		Cost:            5,
		SellerID:        "sellerid",
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`update products set name`).WithArgs(prod.Name, prod.Cost, prod.ID, prod.SellerID).WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	if err := NewApp("", db).dbUpdateProduct(context.Background(), prod); err == nil {
		t.Fatal("should fail if the query fails")
	} else {
		if err.Error() != "db error" {
			t.Fatal("should return the database error")
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}

}

func TestDbUpdateProductFailSuccess(t *testing.T) {
	prod := model.Product{
		ID:              "id",
		Name:            "name",
		AmountAvailable: 10,
		Cost:            5,
		SellerID:        "sellerid",
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`update products set name`).WithArgs(prod.Name, prod.Cost, prod.ID, prod.SellerID).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := NewApp("", db).dbUpdateProduct(context.Background(), prod); err != nil {
		t.Errorf("product updated but received error: %s", err.Error())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}

}

func TestDbUserUpdateDepositFailNoDb(t *testing.T) {
	if err := NewApp("", nil).dbUserUpdateDeposit(context.Background(), "", 0); err == nil {
		t.Fatal("should fail if no database connection")
	} else {
		errMsg := "no database configured"
		if err.Error() != errMsg {
			t.Fatalf("wrong error message. expected: %s, got: %s", errMsg, err.Error())
		}
	}
}

func TestDbUserUpdateDepositFailNoTx(t *testing.T) {

	errMsg := "no tx"

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin().WillReturnError(errors.New(errMsg))

	if err := NewApp("", db).dbUserUpdateDeposit(context.Background(), "", 0); err == nil {
		t.Fatal("should fail if couldn't open a transaction")
	} else {
		if err.Error() != errMsg {
			t.Fatalf("wrong error message. expected: %s, got: %s", errMsg, err.Error())
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDbUserUpdateDepositFailQueryError(t *testing.T) {

	errMsg := "update error"

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`update users set deposit=\? where id=\?`).WithArgs(10, "userid").WillReturnError(errors.New(errMsg))
	mock.ExpectRollback()

	if err := NewApp("", db).dbUserUpdateDeposit(context.Background(), "userid", 10); err == nil {
		t.Fatal("should fail if update failed")
	} else {
		if err.Error() != errMsg {
			t.Fatalf("wrong error message. expected: %s, got: %s", errMsg, err.Error())
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDbUserUpdateDepositSuccess(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`update users set deposit=\? where id=\?`).WithArgs(10, "userid").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := NewApp("", db).dbUserUpdateDeposit(context.Background(), "userid", 10); err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
