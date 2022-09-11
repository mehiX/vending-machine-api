package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
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
	mock.ExpectExec(`insert into users`).WithArgs(sqlmock.AnyArg(), data.Username, sqlmock.AnyArg(), data.Role).WillReturnResult(sqlmock.NewResult(1, 1))
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

func TestHandleResetFailNoUserInContext(t *testing.T) {

	r, err := http.NewRequest(http.MethodGet, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	NewApp("", nil).handleReset().ServeHTTP(w, r)

	resp := w.Result()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("wrong status code. expected: %d, got: %d", http.StatusUnauthorized, resp.StatusCode)
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	txt := strings.TrimSpace(string(b))

	if txt != http.StatusText(http.StatusUnauthorized) {
		t.Fatalf("wrong error message. expected: %s, got: %s", http.StatusText(http.StatusUnauthorized), txt)
	}
}

func TestHandleResetFailUserIsNotBuyer(t *testing.T) {

	usr := model.User{
		Role: model.ROLE_SELLER,
	}

	r, err := http.NewRequest(http.MethodGet, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	ctx := context.WithValue(r.Context(), userContextKey, &usr)

	NewApp("", nil).handleReset().ServeHTTP(w, r.WithContext(ctx))

	resp := w.Result()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("wrong status code. expected: %d, got: %d", http.StatusUnauthorized, resp.StatusCode)
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	txt := strings.TrimSpace(string(b))

	if txt != http.StatusText(http.StatusUnauthorized) {
		t.Fatalf("wrong error message. expected: %s, got: %s", http.StatusText(http.StatusUnauthorized), txt)
	}

}

func TestHandleResetFailDatabaseErr(t *testing.T) {

	usr := model.User{
		Role: model.ROLE_BUYER,
	}

	r, err := http.NewRequest(http.MethodGet, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	ctx := context.WithValue(r.Context(), userContextKey, &usr)

	NewApp("", nil).handleReset().ServeHTTP(w, r.WithContext(ctx))

	resp := w.Result()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("wrong status code. expected: %d, got: %d", http.StatusUnauthorized, resp.StatusCode)
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	txt := strings.TrimSpace(string(b))

	if txt != "reset failed" {
		t.Fatalf("wrong error message. expected: %s, got: %s", "reset failed", txt)
	}

}

func TestHandleResetSuccess(t *testing.T) {

	usr := model.User{
		ID:      "userid",
		Role:    model.ROLE_BUYER,
		Deposit: 100,
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`update users set deposit=\? where id=\?`).WithArgs(0, "userid").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	columns := []string{"id", "username", "password", "deposit", "role"}

	mock.ExpectQuery(`select .* from users where id=\?`).WithArgs("userid").WillReturnRows(sqlmock.NewRows(columns).AddRow("userid", "", "", 0, "BUYER"))

	r, err := http.NewRequest(http.MethodGet, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	ctx := context.WithValue(r.Context(), userContextKey, &usr)

	NewApp("", db).handleReset().ServeHTTP(w, r.WithContext(ctx))

	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("wrong status code. expected: %d, got: %d", http.StatusOK, resp.StatusCode)
	}

	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&usr); err != nil {
		t.Fatal(err)
	}

	if usr.Deposit != 0 {
		t.Fatalf("wrong deposit in response. expected: %d, got: %d", 0, usr.Deposit)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestReturnUserAsJsonFailBadUserId(t *testing.T) {

	w := httptest.NewRecorder()
	NewApp("", nil).returnUserAsJson(context.Background(), w, "")

	resp := w.Result()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("wrong status code. expected: %d, got: %d", http.StatusInternalServerError, resp.StatusCode)
	}
}

func TestHandleDepositFailNoUserInContext(t *testing.T) {

	r, err := http.NewRequest(http.MethodGet, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	NewApp("", nil).handleDeposit().ServeHTTP(w, r)

	resp := w.Result()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("wrong status code. expected: %d, got: %d", http.StatusUnauthorized, resp.StatusCode)
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	txt := strings.TrimSpace(string(b))

	if txt != http.StatusText(http.StatusUnauthorized) {
		t.Fatalf("wrong error message. expected: %s, got: %s", http.StatusText(http.StatusUnauthorized), txt)
	}
}

func TestHandleDepositFailNoCoinValueInContext(t *testing.T) {

	r, err := http.NewRequest(http.MethodGet, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	ctx := context.WithValue(r.Context(), userContextKey, &model.User{Role: model.ROLE_BUYER})

	NewApp("", nil).handleDeposit().ServeHTTP(w, r.WithContext(ctx))

	resp := w.Result()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("wrong status code. expected: %d, got: %d", http.StatusBadRequest, resp.StatusCode)
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	txt := strings.TrimSpace(string(b))

	if txt != "missing coin value" {
		t.Fatalf("wrong error message. expected: %s, got: %s", "missing coin value", txt)
	}
}

func TestHandleDepositFailDbError(t *testing.T) {

	r, err := http.NewRequest(http.MethodGet, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	ctx := context.WithValue(r.Context(), userContextKey, &model.User{Role: model.ROLE_BUYER})
	var val int = 10
	ctx = context.WithValue(ctx, coinValueContextKey, &val)

	NewApp("", nil).handleDeposit().ServeHTTP(w, r.WithContext(ctx))

	resp := w.Result()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("wrong status code. expected: %d, got: %d", http.StatusInternalServerError, resp.StatusCode)
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	txt := strings.TrimSpace(string(b))

	if txt != "deposit failed" {
		t.Fatalf("wrong error message. expected: %s, got: %s", "deposit failed", txt)
	}
}

func TestHandleDepositSuccess(t *testing.T) {

	usr := model.User{
		ID:      "userid",
		Role:    model.ROLE_BUYER,
		Deposit: 100,
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`update users set deposit=\? where id=\?`).WithArgs(110, "userid").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	columns := []string{"id", "username", "password", "deposit", "role"}

	mock.ExpectQuery(`select .* from users where id=\?`).WithArgs("userid").WillReturnRows(sqlmock.NewRows(columns).AddRow("userid", "", "", 110, "BUYER"))

	r, err := http.NewRequest(http.MethodGet, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	ctx := context.WithValue(r.Context(), userContextKey, &usr)
	var val int = 10
	ctx = context.WithValue(ctx, coinValueContextKey, &val)

	NewApp("", db).handleDeposit().ServeHTTP(w, r.WithContext(ctx))

	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("wrong status code. expected: %d, got: %d", http.StatusOK, resp.StatusCode)
	}

	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&usr); err != nil {
		t.Fatal(err)
	}

	if usr.Deposit != 110 {
		t.Fatalf("wrong deposit in response. expected: %d, got: %d", 110, usr.Deposit)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestReturnAsJSONFail(t *testing.T) {

	w := httptest.NewRecorder()

	invalidValue := make(chan int)

	returnAsJSON(context.Background(), w, invalidValue)

	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("wrong status code. expected: %d, got: %d", http.StatusInternalServerError, resp.StatusCode)
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	txt := strings.TrimSpace(string(b))
	if txt != http.StatusText(resp.StatusCode) {
		t.Fatalf("wrong error message. expected: %s, got: %s", http.StatusText(resp.StatusCode), txt)
	}
}

func TestOnlyBuyersCanBuy(t *testing.T) {
	type scenario struct {
		name       string
		user       *model.User
		product    *model.Product
		amount     int
		seller     *model.User
		statusCode int
		response   buyResponse
	}

	scenarios := []scenario{
		{name: "user is admin", user: &model.User{Role: model.ROLE_ADMIN}, statusCode: http.StatusUnauthorized},
		{name: "user is seller", user: &model.User{Role: model.ROLE_SELLER}, statusCode: http.StatusUnauthorized},
		{name: "buyer, no product", user: &model.User{Role: model.ROLE_BUYER}, statusCode: http.StatusNotFound},
		{name: "buyer, product, no amount", user: &model.User{Role: model.ROLE_BUYER}, product: &model.Product{ID: "productid"}, statusCode: http.StatusBadRequest},
		{name: "buyer, product, amount, no seller info", user: &model.User{Role: model.ROLE_BUYER}, product: &model.Product{ID: "productid"}, amount: 5, statusCode: http.StatusNotFound},
		{name: "buyer, product, amount, no seller info", user: &model.User{Role: model.ROLE_BUYER}, product: &model.Product{ID: "productid", SellerID: "seller-id"}, amount: 5, statusCode: http.StatusNotFound},
		{name: "no availability", user: &model.User{Role: model.ROLE_BUYER}, product: &model.Product{ID: "productid", SellerID: "seller-id", AmountAvailable: 3}, amount: 5, seller: &model.User{ID: "seller-id", Role: model.ROLE_SELLER}, statusCode: http.StatusBadRequest},
		{name: "not enough deposit", user: &model.User{Role: model.ROLE_BUYER, Deposit: 15}, product: &model.Product{ID: "productid", SellerID: "seller-id", AmountAvailable: 10, Cost: 5}, amount: 5, seller: &model.User{ID: "seller-id", Role: model.ROLE_SELLER}, statusCode: http.StatusBadRequest},
		{name: "no database", user: &model.User{Role: model.ROLE_BUYER, Deposit: 30}, product: &model.Product{ID: "productid", SellerID: "seller-id", AmountAvailable: 10, Cost: 5}, amount: 5, seller: &model.User{ID: "seller-id", Role: model.ROLE_SELLER}, statusCode: http.StatusBadRequest},
		{
			name:       "all good",
			user:       &model.User{Role: model.ROLE_BUYER, Deposit: 30},
			product:    &model.Product{ID: "productid", Name: "product name", SellerID: "seller-id", AmountAvailable: 10, Cost: 5},
			amount:     5,
			seller:     &model.User{ID: "seller-id", Username: "best-seller", Role: model.ROLE_SELLER},
			statusCode: http.StatusOK,
			response: buyResponse{
				Amount:     5,
				TotalSpent: 25,
				Change:     [5]int64{1, 0, 0, 0, 0},
				Product: prodBuyerInfo{
					Name:       "product name",
					Cost:       5,
					SellerName: "best-seller",
				},
			},
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			r, err := http.NewRequest(http.MethodPost, "/buy", nil)
			if err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()

			ctx := context.WithValue(r.Context(), userContextKey, s.user)
			if s.product != nil {
				ctx = context.WithValue(ctx, productContextKey, s.product)
			}
			if s.amount != 0 {
				ctx = context.WithValue(ctx, amountValueContextKey, &s.amount)
			}
			if s.seller != nil {
				ctx = context.WithValue(ctx, sellerContextKey, s.seller)
			}

			if s.name != "all good" {
				NewApp("", nil).handleBuy().ServeHTTP(w, r.WithContext(ctx))
			} else {
				// the happy scenario
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatal(err)
				}
				defer db.Close()

				mock.ExpectBegin()
				mock.ExpectExec(`update products set available_amount`).WithArgs(s.amount, s.product.ID).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(`update users set deposit `).WithArgs(int64(s.amount)*s.product.Cost, s.user.ID).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()

				NewApp("", db).handleBuy().ServeHTTP(w, r.WithContext(ctx))

				if err := mock.ExpectationsWereMet(); err != nil {
					t.Error(err)
				}

				resp := w.Result()
				defer resp.Body.Close()

				var data buyResponse
				if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
					t.Fatal(err)
				}

				// validate returned data
				if data.Amount != s.response.Amount {
					t.Errorf("wrong amount. expected: %d, got: %d", s.response.Amount, data.Amount)
				}
				if data.TotalSpent != s.response.TotalSpent {
					t.Errorf("wrong total spent. expected: %d, got: %d", s.response.TotalSpent, data.TotalSpent)
				}
				if data.Change != s.response.Change {
					t.Errorf("wrong change. expected: %v, got: %v", s.response.Change, data.Change)
				}
				if !reflect.DeepEqual(data.Product, s.response.Product) {
					t.Errorf("wrong product info. expected: %v, got: %v", s.response.Product, data.Product)
				}
			}

			if w.Result().StatusCode != s.statusCode {
				t.Errorf("wrong status code. expected: %d, got: %d", s.statusCode, w.Result().StatusCode)
			}
		})
	}
}
