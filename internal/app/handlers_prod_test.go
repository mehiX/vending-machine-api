package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mehiX/vending-machine-api/internal/app/model"
)

func TestHandleCreateProductFailNoSeller(t *testing.T) {

	r, err := http.NewRequest(http.MethodPost, "/product", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	NewApp("", nil).handleCreateProduct().ServeHTTP(w, r)

	sc := w.Result().StatusCode
	if sc != http.StatusUnauthorized {
		t.Error("should have a seller logged in")
	}
}

func TestHandleCreateProductFailNoData(t *testing.T) {

	r, err := http.NewRequest(http.MethodPost, "/product", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	ctx := context.WithValue(r.Context(), userContextKey, &model.User{Role: model.ROLE_SELLER})

	NewApp("", nil).handleCreateProduct().ServeHTTP(w, r.WithContext(ctx))

	sc := w.Result().StatusCode
	if sc != http.StatusBadRequest {
		t.Error("should error if no product data in body")
	}
}

func TestHandleCreateProductFailWrongData(t *testing.T) {

	var buf bytes.Buffer
	buf.WriteString("blah blah")

	r, err := http.NewRequest(http.MethodPost, "/product", &buf)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	ctx := context.WithValue(r.Context(), userContextKey, &model.User{Role: model.ROLE_SELLER})

	NewApp("", nil).handleCreateProduct().ServeHTTP(w, r.WithContext(ctx))

	sc := w.Result().StatusCode
	if sc != http.StatusBadRequest {
		t.Error("should error if it receives wrong data")
	}
}

func TestHandleCreateProductSuccess(t *testing.T) {

	var buf bytes.Buffer
	data := createProductRequest{
		AmountAvailable: 100,
		Cost:            10,
		Name:            "Product 1",
	}
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		t.Fatal(err)
	}

	r, err := http.NewRequest(http.MethodPost, "/product", &buf)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	usr := model.User{
		ID:   "id123",
		Role: model.ROLE_SELLER,
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`insert into products`).WithArgs(sqlmock.AnyArg(), "Product 1", 100, 10, "id123").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	ctx := context.WithValue(context.Background(), userContextKey, &usr)

	NewApp("", db).handleCreateProduct().ServeHTTP(w, r.WithContext(ctx))

	sc := w.Result().StatusCode
	if sc != http.StatusCreated {
		t.Errorf("should succeed. Status code expected: %d, got: %d", http.StatusCreated, sc)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestHandleDeleteProductFailNoUser(t *testing.T) {

	r, err := http.NewRequest(http.MethodDelete, "/product", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	NewApp("", nil).handleDeleteProduct().ServeHTTP(w, r)

	sc := w.Result().StatusCode
	if sc != http.StatusUnauthorized {
		t.Error("should have a user logged in")
	}
}

func TestHandleDeleteProductFailNoData(t *testing.T) {

	r, err := http.NewRequest(http.MethodDelete, "/product", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	ctx := context.WithValue(r.Context(), userContextKey, &model.User{Role: model.ROLE_SELLER})

	NewApp("", nil).handleDeleteProduct().ServeHTTP(w, r.WithContext(ctx))

	sc := w.Result().StatusCode
	if sc != http.StatusBadRequest {
		t.Error("should error if no product data in body")
	}
}
