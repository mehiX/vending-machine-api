package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mehiX/vending-machine-api/internal/app/model"
)

func TestHandleAddUserNonJSONInput(t *testing.T) {

	var buf bytes.Buffer
	buf.WriteString("blah blah")

	r, err := http.NewRequest(http.MethodPost, "/", &buf)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	vm := NewApp("", nil)
	vm.handleAddUser().ServeHTTP(w, r)

	resp := w.Result()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("wrong status code. Expected: %d, got: %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestHandleAddUserFailDataError(t *testing.T) {

	// various data scenarios are already tested for CreateUser directly. here we only car for an error response, not the reason for it
	var buf bytes.Buffer
	data := addUserRequest{
		Username: "short",
		Password: "lasdjfasdf",
		Deposit:  100,
		Role:     model.ROLE_ADMIN,
	}
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		t.Fatal(err)
	}

	r, err := http.NewRequest(http.MethodPost, "/", &buf)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	vm := NewApp("", db)
	vm.handleAddUser().ServeHTTP(w, r)

	resp := w.Result()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("wrong status code. Expected: %d, got: %d", http.StatusInternalServerError, resp.StatusCode)
	}

}

func TestHandleAddUserSuccess(t *testing.T) {

	// various data scenarios are already tested for CreateUser directly. here we only car for an error response, not the reason for it
	var buf bytes.Buffer
	data := addUserRequest{
		Username: "mihaiusr",
		Password: "la&*jfaS2f",
		Deposit:  100,
		Role:     model.ROLE_ADMIN,
	}
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		t.Fatal(err)
	}

	r, err := http.NewRequest(http.MethodPost, "/", &buf)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`insert into users`).WithArgs(sqlmock.AnyArg(), data.Username, sqlmock.AnyArg(), data.Deposit, data.Role).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	vm := NewApp("", db)
	vm.handleAddUser().ServeHTTP(w, r)

	resp := w.Result()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("wrong status code. Expected: %d, got: %d", http.StatusCreated, resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}

}

func TestHandleLoginFailWrongCredentials(t *testing.T) {

	data := loginRequest{
		Username: "short",
		Password: "doesnotmatter",
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		t.Fatal(err)
	}

	r, err := http.NewRequest(http.MethodPost, "/", &buf)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`select id, username, password, deposit, role from users where username=`).WithArgs(data.Username).WillReturnError(errors.New("no rows"))

	NewApp("", db).handleLogin().ServeHTTP(w, r)

	resp := w.Result()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("wrong status code for wrong credentials. Expected: %d, got: %d", http.StatusUnauthorized, resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}
