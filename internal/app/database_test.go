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

func TestCreateUserShouldRollbackOnError(t *testing.T) {

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

func TestFindUserByUsernameSuccess(t *testing.T) {

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

func TestFindUserByUsernameNoMatch(t *testing.T) {

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

func TestFindUserByIDSuccess(t *testing.T) {

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

func TestFindUserByIDNoMatch(t *testing.T) {

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

func TestDbCreateUserCannotStartTx(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin().WillReturnError(errors.New("cannot open tx"))

	if err := NewApp("", db).dbCreateUser(context.Background(), "", "", 10, ""); err == nil {
		t.Fatal("should fail if DB cannot open a TX")
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
