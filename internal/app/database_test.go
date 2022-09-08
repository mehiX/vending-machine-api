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
