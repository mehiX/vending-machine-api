package app

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mehiX/vending-machine-api/internal/app/model"
	"golang.org/x/crypto/bcrypt"
)

func TestFindByCredentialsMissingUsername(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`select * from users`).WillReturnError(errors.New("no rows found"))

	vm := NewApp("", db)
	u, err := vm.FindUserByCredentials(context.Background(), "missingusername", "somepassword")
	if err == nil || u != nil {
		t.Error("expect error for missing username")
	}
}

func TestFindByCredentialsPasswordMissmatch(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	testUser := "mihaiusr"
	testPassword := "mh12&^KJlwekJ*"
	testRole := model.ROLE_BUYER

	columns := []string{"id", "username", "password", "deposit", "role"}

	encPasswd, _ := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.MinCost)

	mock.ExpectQuery("select id, username, password, deposit, role from users where username=").
		WithArgs(testUser).WillReturnRows(sqlmock.NewRows(columns).
		AddRow("id1", testUser, encPasswd, 100, testRole))

	vm := NewApp("", db)
	u, err := vm.FindUserByCredentials(context.Background(), testUser, "somepasswordDifferentThanTestPassword")
	if err == nil || u != nil {
		t.Error("expect error for wrong password")
	}
}

func TestCreateUserSuccess(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	testPassword := "ui&*789SDJA87&"

	mock.ExpectBegin()
	mock.ExpectExec(`insert into users`).WithArgs(sqlmock.AnyArg(), "goodusername", sqlmock.AnyArg(), 100, model.ROLE_ADMIN).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := NewApp("", db).CreateUser(context.Background(), "goodusername", testPassword, 100, model.ROLE_ADMIN); err != nil {
		t.Errorf("unexpected error when creating a user: %s", err)
	}

}
func TestCreateUserFailOnValidateUsername(t *testing.T) {

	if err := NewApp("", nil).CreateUser(context.Background(), "short", "", 100, model.ROLE_ADMIN); err == nil {
		t.Error("expect to return error for invalid username")
	}
}

func TestCreateUserFailOnValidatePassword(t *testing.T) {

	if err := NewApp("", nil).CreateUser(context.Background(), "goodusername", "wrong", 100, model.ROLE_ADMIN); err == nil {
		t.Error("expect to return error for invalid password")
	}
}

func TestCreateUserFailOnValidateDeposit(t *testing.T) {

	if err := NewApp("", nil).CreateUser(context.Background(), "goodusername", "ui&*789SDJA87&", 101, model.ROLE_ADMIN); err == nil {
		t.Error("expect to return error for invalid deposit")
	}
}

func TestCreateUserFailOnValidateRole(t *testing.T) {

	if err := NewApp("", nil).CreateUser(context.Background(), "goodusername", "ui&*789SDJA87&", 100, model.TypeRole("anything")); err == nil {
		t.Error("expect to return error for invalid role")
	}
}

func TestCreateUserFailOnNoDatabase(t *testing.T) {

	if err := NewApp("", nil).CreateUser(context.Background(), "goodusername", "ui&*789SDJA87&", 100, model.ROLE_ADMIN); err == nil {
		t.Error("expect to return error if no database defined")
	}
}

func TestUserDepositCoinFailWrongCoin(t *testing.T) {
	if err := NewApp("", nil).UserDepositCoin(context.Background(), nil, 40); err == nil {
		t.Fatal("should not accept wrong coin values")
	} else {
		if err.Error() != "coin value not allowed" {
			t.Fatalf("wrong error message. Expected: %s, got: %s", "coin value not allowed", err.Error())
		}
	}
}
