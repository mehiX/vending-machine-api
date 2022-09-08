package app

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mehiX/vending-machine-api/internal/app/model"
)

func TestCreateProductFailNoSeller(t *testing.T) {
	if err := NewApp("", nil).CreateProduct(context.Background(), nil, 0, 0, ""); err == nil {
		t.Fatal("should fail if no seller")
	}
}

func TestCreateProductFaiInvalidAmount(t *testing.T) {
	if err := NewApp("", nil).CreateProduct(context.Background(), &model.User{ID: "id1"}, -1, 0, ""); err == nil {
		t.Fatal("should fail if invalid amount")
	}
}

func TestCreateProductFailInvalidCost(t *testing.T) {
	if err := NewApp("", nil).CreateProduct(context.Background(), &model.User{ID: "id1"}, 1, 0, ""); err == nil {
		t.Fatal("should fail if invalid cost")
	}
}

func TestCreateProductFailInvalidName(t *testing.T) {
	if err := NewApp("", nil).CreateProduct(context.Background(), &model.User{ID: "id1"}, 1, 10, ""); err == nil {
		t.Fatal("should fail if invalid name")
	}
}

func TestCreateProductFailIfNoDatabase(t *testing.T) {
	if err := NewApp("", nil).CreateProduct(context.Background(), &model.User{ID: "id1"}, 1, 10, "kasjdfja"); err == nil {
		t.Fatal("should fail if no database connection")
	}
}

func TestCreateProductSuccess(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`insert into products`).WithArgs(sqlmock.AnyArg(), "kasjdfja", 1, 5, "id1").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := NewApp("", db).CreateProduct(context.Background(), &model.User{ID: "id1"}, 1, 5, "kasjdfja"); err != nil {
		t.Fatalf("should create product. Received: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCreateProductFailDatabaseError(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`insert into products`).WithArgs(sqlmock.AnyArg(), "kasjdfja", 1, 0, "id1").WillReturnError(errors.New("duplicate name"))
	mock.ExpectRollback()

	if err := NewApp("", db).CreateProduct(context.Background(), &model.User{ID: "id1"}, 1, 0, "kasjdfja"); err == nil {
		t.Fatal("should fail if there are database errors")
	}
}
