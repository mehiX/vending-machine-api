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
	if err := vm.dbCreateUser(context.Background(), "mihaiuser", "strong23Pass*", 100, model.ROLE_USER); err != nil {
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
	if err := vm.dbCreateUser(context.Background(), "mihaiuser", "strong23Pass*", 100, model.ROLE_USER); err == nil {
		t.Fatal(errors.New("database error should be returned by the function"))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there are unfulfilled expectations: %s", err)
	}
}
