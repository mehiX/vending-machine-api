package app

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mehiX/vending-machine-api/internal/app/model"
)

func TestUsernameValidate(t *testing.T) {

	type scenario struct {
		input   string
		isValid bool
	}

	scenarios := []scenario{
		{"skjdfs", false},
		{"12345678", true},
		{"ksjsksjdsk!", false},
		{"verylongusername", true},
		{"mix@valid.name-", true},
	}

	for _, s := range scenarios {
		err := validateUsername(s.input)
		if s.isValid && err != nil {
			t.Errorf("username should be valid: %s", s.input)
		}

		if !s.isValid && err == nil {
			t.Errorf("username should not pass validation: %s", s.input)
		}
	}
}

func TestPasswordValidate(t *testing.T) {
	type scenario struct {
		input   string
		isValid bool
	}

	scenarios := []scenario{
		{"12short", false},
		{"kasdfjadfjdasfkjakdsf", false},
		{"1233484744", false},
		{"AAAHAHAHAHAA", false},
		{"A678", false},
		{"hajas^&^8*&", false},
		{"123HDhhasdKJJJM", false},
		{"mhGP*&UksdfLK", false},
		{"mhG2P*&UksdfLK", true},
		{"kjadf SKS k& k7(*  ", true},
		{"    ksjE#2j    ", false},
	}

	for _, s := range scenarios {
		err := validatePassword(s.input)
		if s.isValid && err != nil {
			t.Errorf("password should be valid: %s", s.input)
		}

		if !s.isValid && err == nil {
			t.Errorf("password should not pass validation: %s", s.input)
		}
	}
}

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
	if err := vm.CreateUser(context.Background(), "mihaiuser", "strong23Pass*", 100, model.ROLE_USER); err != nil {
		t.Fatal(err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there are unfulfilled expectations: %s", err)
	}
}
