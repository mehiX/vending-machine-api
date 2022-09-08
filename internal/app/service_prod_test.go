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

func TestCreateProductFailUserNotSeller(t *testing.T) {
	if err := NewApp("", nil).CreateProduct(context.Background(), &model.User{ID: "id1", Role: model.ROLE_BUYER}, 10, 5, "product name"); err == nil {
		t.Fatal("should fail if user is not a seller")
	} else {
		if err.Error() != "user is not a seller" {
			t.Fatalf("wrong error. Expected: %s, got: %s", "user is not a seller", err.Error())
		}
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

	if err := NewApp("", db).CreateProduct(context.Background(), &model.User{ID: "id1", Role: model.ROLE_SELLER}, 1, 5, "kasjdfja"); err != nil {
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

func TestDeleteProductFailNoSeller(t *testing.T) {

	if err := NewApp("", nil).DeleteProduct(context.Background(), nil, nil); err == nil {
		t.Fatal("user nil, should fail")
	} else {
		if err.Error() != "user is nil" {
			t.Fatal("wrong exit error")
		}
	}
}

func TestDeleteProductFailUserIsNotSeller(t *testing.T) {

	if err := NewApp("", nil).DeleteProduct(context.Background(), &model.User{Role: model.ROLE_BUYER}, &model.Product{}); err == nil {
		t.Fatal("user not a seller, should fail")
	} else {
		if err.Error() != "user is not a seller" {
			t.Fatalf("wrong exit error. Expected: %s, got: %s", "user is not a seller", err.Error())
		}
	}
}

func TestDeleteProductFailNoProduct(t *testing.T) {

	if err := NewApp("", nil).DeleteProduct(context.Background(), &model.User{Role: model.ROLE_SELLER}, nil); err == nil {
		t.Fatal("product nil, should fail")
	} else {
		if err.Error() != "product is nil" {
			t.Fatal("wrong exit error")
		}
	}
}

func TestDeleteProductFailNoDatabaseConn(t *testing.T) {

	if err := NewApp("", nil).DeleteProduct(context.Background(), &model.User{Role: model.ROLE_SELLER}, &model.Product{}); err == nil {
		t.Fatal("db conn nil, should fail")
	} else {
		if err.Error() != "no db conn" {
			t.Fatal("wrong exit error")
		}
	}
}

func TestDeleteProductFailSellerNotSame(t *testing.T) {

	if err := NewApp("", nil).DeleteProduct(context.Background(), &model.User{ID: "seller 1", Role: model.ROLE_SELLER}, &model.Product{SellerID: "seller 2"}); err == nil {
		t.Fatal("not the same seller id, should fail")
	} else {
		if err.Error() != "wrong seller id" {
			t.Fatal("wrong exit error")
		}
	}
}

func TestDeleteProductFailDbError(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`delete from products where id\=`).WillReturnError(errors.New("delete failed"))
	mock.ExpectRollback()

	if err := NewApp("", db).DeleteProduct(context.Background(), &model.User{ID: "seller 1", Role: model.ROLE_SELLER}, &model.Product{SellerID: "seller 1"}); err == nil {
		t.Fatal("delete statement returned error, should fail")
	} else {
		if err.Error() != "delete failed" {
			t.Fatalf("wrong exit error. Expected: %s, got: %s", "delete error", err.Error())
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteProductSuccess(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`delete from products where id\=`).WithArgs("product 1", "seller 1").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := NewApp("", db).DeleteProduct(context.Background(), &model.User{ID: "seller 1", Role: model.ROLE_SELLER}, &model.Product{ID: "product 1", SellerID: "seller 1"}); err != nil {
		t.Fatalf("delete succeeded, should not fail. Error: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
