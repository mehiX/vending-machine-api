package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

func TestHandleListProductsNoDb(t *testing.T) {

	r, err := http.NewRequest(http.MethodGet, "/product/list", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	NewApp("", nil).handleListProducts().ServeHTTP(w, r)

	sc := w.Result().StatusCode

	if sc != http.StatusInternalServerError {
		t.Errorf("wrong status code. Expected: %d, got: %d", http.StatusInternalServerError, sc)
	}
}

func TestHandleListProductsDbConnFail(t *testing.T) {

	r, err := http.NewRequest(http.MethodGet, "/product/list", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	db.Close()

	NewApp("", db).handleListProducts().ServeHTTP(w, r)

	sc := w.Result().StatusCode

	if sc != http.StatusInternalServerError {
		t.Errorf("wrong status code. Expected: %d, got: %d", http.StatusInternalServerError, sc)
	}
}

func TestHandleListProductsDbFail(t *testing.T) {

	r, err := http.NewRequest(http.MethodGet, "/product/list", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`select .* from products`).WillReturnError(errors.New("db error"))

	NewApp("", db).handleListProducts().ServeHTTP(w, r)

	sc := w.Result().StatusCode

	if sc != http.StatusInternalServerError {
		t.Errorf("wrong status code. Expected: %d, got: %d", http.StatusInternalServerError, sc)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestHandleListProductsSuccessNone(t *testing.T) {

	r, err := http.NewRequest(http.MethodGet, "/product/list", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`select .* from products`).WillReturnRows(&sqlmock.Rows{})

	NewApp("", db).handleListProducts().ServeHTTP(w, r)

	resp := w.Result()
	sc := resp.StatusCode

	if sc != http.StatusOK {
		t.Errorf("wrong status code. Expected: %d, got: %d", http.StatusOK, sc)
	}

	ct := resp.Header.Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("wrong content type. expected: %s, got: %s", "application/json", ct)
	}

	defer resp.Body.Close()

	var products []model.Product
	if err := json.NewDecoder(resp.Body).Decode(&products); err != nil {
		t.Fatal(err)
	}

	if len(products) != 0 {
		t.Errorf("wrong number of records. Expected: %d, got: %d", 0, len(products))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestHandleListProductsSuccessMany(t *testing.T) {

	r, err := http.NewRequest(http.MethodGet, "/product/list", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	cols := []string{"id", "name", "available_amount", "cost", "seller_id"}
	mock.ExpectQuery(`select .* from products`).WillReturnRows(sqlmock.NewRows(cols).
		AddRow("1", "prod 1", 10, 10, "seller 1").
		AddRow("2", "prod 2", 10, 10, "seller 1").
		AddRow("3", nil, 10, 10, "seller 1"))

	NewApp("", db).handleListProducts().ServeHTTP(w, r)

	resp := w.Result()
	sc := resp.StatusCode

	if sc != http.StatusOK {
		t.Errorf("wrong status code. Expected: %d, got: %d", http.StatusOK, sc)
	}

	ct := resp.Header.Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("wrong content type. expected: %s, got: %s", "application/json", ct)
	}

	defer resp.Body.Close()

	var products []model.Product
	if err := json.NewDecoder(resp.Body).Decode(&products); err != nil {
		t.Fatal(err)
	}

	if len(products) != 2 {
		t.Errorf("wrong number of records. Expected: %d, got: %d", 2, len(products))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestHandleProductDetailsFailNoProdInCtx(t *testing.T) {

	r, err := http.NewRequest(http.MethodGet, "/product/1234", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	NewApp("", nil).handleProductDetails().ServeHTTP(w, r)

	sc := w.Result().StatusCode

	if sc != http.StatusNotFound {
		t.Errorf("wrong status code. Expected: %d, got: %d", http.StatusNotFound, sc)
	}
}

func TestHandleProductDetailsSuccess(t *testing.T) {

	r, err := http.NewRequest(http.MethodGet, "/product/1234", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	prod := model.Product{
		ID:              "1234",
		Name:            "Prod 1234",
		AmountAvailable: 10,
		Cost:            5,
		SellerID:        "Seller 1",
	}

	ctx := context.WithValue(r.Context(), productContextKey, &prod)

	NewApp("", nil).handleProductDetails().ServeHTTP(w, r.WithContext(ctx))

	resp := w.Result()
	sc := resp.StatusCode

	if sc != http.StatusOK {
		t.Errorf("wrong status code. Expected: %d, got: %d", http.StatusOK, sc)
	}

	ct := resp.Header.Get("Content-Type")

	if ct != "application/json" {
		t.Errorf("wrong content-type. expected: %s, got: %s", "application/json", ct)
	}

	defer resp.Body.Close()

	var p model.Product
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		t.Fatal(err)
	}

	if p.ID != prod.ID {
		t.Errorf("wrong product ID. Expected: %s, got: %s", prod.ID, p.ID)
	}

	if p.Name != prod.Name {
		t.Errorf("wrong product name. Expected: %s, got: %s", prod.Name, p.Name)
	}

	if p.SellerID != prod.SellerID {
		t.Errorf("wrong seller ID. Expected: %s, got: %s", prod.SellerID, p.SellerID)
	}

	if p.AmountAvailable != prod.AmountAvailable {
		t.Errorf("wrong amount available. Expected: %d, got: %d", prod.AmountAvailable, p.AmountAvailable)
	}

	if p.Cost != prod.Cost {
		t.Errorf("wrong cost. Expected: %d, got: %d", prod.Cost, p.Cost)
	}
}

func TestHandleUpdateProductFailNoUserInCtx(t *testing.T) {

}

func TestHandleUpdateProductFailUserIsNotSeller(t *testing.T) {

}

func TestHandleUpdateProductFailNoProductInCtx(t *testing.T) {

}

func TestHandleUpdateProductFailNotSameSeller(t *testing.T) {

}

func TestHandleUpdateProductFailNoDatabase(t *testing.T) {

}

func TestHandleUpdateProductFailDatabaseError(t *testing.T) {

}

func TestHandleUpdateProductFailWrongName(t *testing.T) {

}

func TestHandleUpdateProductFailWrongCost(t *testing.T) {

}

func TestHandleUpdateProductSuccess(t *testing.T) {

}
