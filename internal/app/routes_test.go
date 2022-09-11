package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/mehiX/vending-machine-api/internal/app/model"
	"golang.org/x/crypto/bcrypt"
)

func TestHealthNoDb(t *testing.T) {

	r, err := http.NewRequest(http.MethodGet, "/health", nil)
	if err != nil {
		t.Error(err)
	}
	w := httptest.NewRecorder()

	NewApp("", nil).Router.ServeHTTP(w, r)

	resp := w.Result()

	if resp.StatusCode != http.StatusFailedDependency {
		t.Errorf("/health. Expected: %d, got: %d", http.StatusFailedDependency, resp.StatusCode)
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}

	if string(b) != "NO DB" {
		t.Errorf("/health body. Expected: NO DB, got: %s", string(b))
	}
}

func TestHealthOK(t *testing.T) {

	r, err := http.NewRequest(http.MethodGet, "/health", nil)
	if err != nil {
		t.Error(err)
	}
	w := httptest.NewRecorder()

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	NewApp("", db).Router.ServeHTTP(w, r)

	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("/health. Expected: %d, got: %d", http.StatusOK, resp.StatusCode)
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}

	if string(b) != "OK" {
		t.Errorf("/health body. Expected: OK, got: %s", string(b))
	}
}

func TestValidateExists(t *testing.T) {

	r, err := http.NewRequest(http.MethodGet, "/user", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	NewApp("", nil).Router.ServeHTTP(w, r)

	resp := w.Result()

	if resp.StatusCode == http.StatusNotFound {
		t.Error("GET /user should be defined")
	}
}

func TestProtectedRoutes(t *testing.T) {
	type route struct {
		method string
		path   string
	}

	routes := []route{
		{
			method: http.MethodGet,
			path:   "/user",
		},
	}

	router := NewApp("", nil).Router

	for _, r := range routes {
		req, err := http.NewRequest(r.method, r.path, nil)
		if err != nil {
			t.Error(err)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		resp := w.Result()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("%s %s wrong status code. Expected: %d, got: %d", r.method, r.path, http.StatusUnauthorized, resp.StatusCode)
		}
	}

}

func TestLoginWrongRequestBody(t *testing.T) {

	router := NewApp("", nil).Router

	var buf bytes.Buffer

	buf.WriteString("some non json data")

	r, err := http.NewRequest(http.MethodPost, "/login", &buf)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	resp := w.Result()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("POST /login wrong status code. Expected: %d, got: %d", http.StatusBadRequest, resp.StatusCode)
	}

}

func TestLoginSuccess(t *testing.T) {

	os.Setenv("JWT_ALG", "HS256")
	os.Setenv("JWT_SIGNKEY", "somekey")

	testID := "id123"
	testUser := "mihaiusr"
	testPassword := "mh12&^KJlwekJ*"
	testRole := model.ROLE_BUYER

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	columns := []string{"id", "username", "password", "deposit", "role"}

	encPasswd, _ := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.MinCost)

	mock.ExpectQuery("select id, username, password, deposit, role from users where username=").
		WithArgs(testUser).WillReturnRows(sqlmock.NewRows(columns).
		AddRow(testID, testUser, encPasswd, 100, testRole))

	vm := NewApp("", db)
	router := vm.Router

	type reqBody struct {
		Username string
		Password string
	}

	body := reqBody{testUser, testPassword}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		t.Fatal(err)
	}

	r, err := http.NewRequest(http.MethodPost, "/login", &buf)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	resp := w.Result()

	defer resp.Body.Close()

	tknString, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	tkn, err := vm.JwtAuth.Decode(string(tknString))
	if err != nil {
		t.Fatal(err)
	}

	claims := tkn.PrivateClaims()
	if usr, ok := claims[jwtUsernameKey]; !ok || usr != testUser {
		t.Error("Wrong or missing claim 'username'")
	}
	if usrID, ok := claims[jwtUserIdKey]; !ok || usrID != testID {
		t.Errorf("Wrong or missing claim 'userID'. Expected: %s, got: %s", testID, usrID)
	}
}

func TestGetCurrentUserData(t *testing.T) {

	r, err := http.NewRequest(http.MethodGet, "/user", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	os.Setenv("JWT_SIGNKEY", "some key")
	os.Setenv("JWT_ALG", "HS256")

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	testUserID := "123"
	testUser := "mihaiusr"
	testPassword := "mh12&^KJlwekJ*"
	testRole := model.ROLE_BUYER

	columns := []string{"id", "username", "password", "deposit", "role"}

	encPasswd, _ := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.MinCost)

	mock.ExpectQuery("select id, username, password, deposit, role from users where id=").
		WithArgs(testUserID).WillReturnRows(sqlmock.NewRows(columns).
		AddRow(testUserID, testUser, encPasswd, 100, testRole))

	vm := NewApp("", db)

	tknStr, err := vm.getEncTokenString(testUserID, testUser)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Authorization", "BEARER "+tknStr)

	vm.Router.ServeHTTP(w, r)

	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /user wrong response code. Expected: %d, got: %d", http.StatusOK, resp.StatusCode)
	}

	defer resp.Body.Close()

	type response struct {
		Username string
		Role     string
	}

	var respData response
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		t.Fatal(err)
	}

	if respData.Username != testUser {
		t.Errorf("GET /user wrong username in response. Expected: %s, got: %s", testUser, respData.Username)
	}

	if respData.Role != testRole {
		t.Errorf("GET /user wrong role in response. Expected: %s, got: %s", testRole, respData.Role)
	}
}

func TestSellersCanCreateProducts(t *testing.T) {

	var buf bytes.Buffer
	data := createProductRequest{
		AmountAvailable: 100,
		Cost:            50,
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

	os.Setenv("JWT_SIGNKEY", "some key")
	os.Setenv("JWT_ALG", "HS256")

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	testUserID := "123"
	testUser := "mihaiusr"
	testPassword := "mh12&^KJlwekJ*"
	testRole := model.ROLE_SELLER

	columns := []string{"id", "username", "password", "deposit", "role"}

	encPasswd, _ := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.MinCost)

	mock.ExpectQuery("select id, username, password, deposit, role from users where id=").
		WithArgs(testUserID).WillReturnRows(sqlmock.NewRows(columns).
		AddRow(testUserID, testUser, encPasswd, 100, testRole))

	vm := NewApp("", db)

	tknStr, err := vm.getEncTokenString(testUserID, testUser)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Authorization", "BEARER "+tknStr)

	vm.Router.ServeHTTP(w, r)

	sc := w.Result().StatusCode

	if sc == http.StatusUnauthorized {
		t.Errorf("Seller account cannot create products")
	}

}

func TestNonSellersCannotCreateProducts(t *testing.T) {

	accounttypes := []model.TypeRole{model.ROLE_ADMIN, model.ROLE_BUYER}

	for _, a := range accounttypes {
		t.Run(a, func(t *testing.T) {
			t.Parallel()
			r, err := http.NewRequest(http.MethodPost, "/product", nil)
			if err != nil {
				t.Fatal(err)
			}

			w := httptest.NewRecorder()

			os.Setenv("JWT_SIGNKEY", "some key")
			os.Setenv("JWT_ALG", "HS256")

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()

			testUserID := "123"
			testUser := "mihaiusr"
			testPassword := "mh12&^KJlwekJ*"
			testRole := a

			columns := []string{"id", "username", "password", "deposit", "role"}

			encPasswd, _ := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.MinCost)

			mock.ExpectQuery("select id, username, password, deposit, role from users where id=").
				WithArgs(testUserID).WillReturnRows(sqlmock.NewRows(columns).
				AddRow(testUserID, testUser, encPasswd, 100, testRole))

			vm := NewApp("", db)

			tknStr, err := vm.getEncTokenString(testUserID, testUser)
			if err != nil {
				t.Fatal(err)
			}
			r.Header.Set("Authorization", "BEARER "+tknStr)

			vm.Router.ServeHTTP(w, r)

			sc := w.Result().StatusCode

			if sc != http.StatusUnauthorized {
				t.Errorf("Buyer accounts can create products. Status code expected: %d, got: %d", http.StatusUnauthorized, sc)
			}

		})
	}
}

func TestUserCtxFailJwtError(t *testing.T) {

	f := func(w http.ResponseWriter, r *http.Request) {

	}

	r, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	// simulate an auth error in the JWT flow
	ctx := context.WithValue(context.Background(), jwtauth.ErrorCtxKey, errors.New("some jwt error"))
	NewApp("", nil).UserCtx(http.HandlerFunc(f)).ServeHTTP(w, r.WithContext(ctx))

	resp := w.Result()
	sc := resp.StatusCode

	if sc != http.StatusUnauthorized {
		t.Error("should fail if no claims")
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.EqualFold(strings.TrimSpace(string(b)), "authentication error (no claims)") {
		t.Errorf("wrong error body. Expected: %s, got: %s", "authentication error (no claims)", string(b))
	}
}

func TestUserCtxFailNoUserIdInClaims(t *testing.T) {

	os.Setenv("JWT_SIGNKEY", "some key")
	os.Setenv("JWT_ALG", "HS256")

	f := func(w http.ResponseWriter, r *http.Request) {

	}

	tkn := jwt.New()
	tkn.Set(jwt.ExpirationKey, time.Now().Add(10*time.Minute))
	tkn.Set(jwt.NotBeforeKey, time.Now())

	claims, _ := tkn.AsMap(context.Background())

	r, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	token, _, err := NewApp("", nil).JwtAuth.Encode(claims)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.WithValue(context.Background(), jwtauth.TokenCtxKey, &token)

	NewApp("", nil).UserCtx(http.HandlerFunc(f)).ServeHTTP(w, r.WithContext(ctx))

	resp := w.Result()
	sc := resp.StatusCode

	if sc != http.StatusUnauthorized {
		t.Fatal("should fail if no claims")
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.EqualFold(strings.TrimSpace(string(b)), "authentication error (no user id)") {
		t.Errorf("wrong error body. Expected: %s, got: %s", "authentication error (no user id)", string(b))
	}
}

func TestUserCtxFailUserIdInClaimsMIssingFromDb(t *testing.T) {

	os.Setenv("JWT_SIGNKEY", "some key")
	os.Setenv("JWT_ALG", "HS256")

	f := func(w http.ResponseWriter, r *http.Request) {

	}

	tkn := jwt.New()
	tkn.Set(jwt.ExpirationKey, time.Now().Add(10*time.Minute))
	tkn.Set(jwt.NotBeforeKey, time.Now())
	tkn.Set(jwtUserIdKey, "user id not in DB")

	claims, err := tkn.AsMap(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	r, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`select .* from users where id=`).WithArgs("user id not in DB").WillReturnError(errors.New("empty result"))

	vm := NewApp("", db)

	token, _, err := vm.JwtAuth.Encode(claims)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.WithValue(context.Background(), jwtauth.TokenCtxKey, token)

	vm.UserCtx(http.HandlerFunc(f)).ServeHTTP(w, r.WithContext(ctx))

	resp := w.Result()
	sc := resp.StatusCode

	if sc != http.StatusUnauthorized {
		t.Fatal("should fail if user id in claims does not exist in DB")
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.EqualFold(strings.TrimSpace(string(b)), "authentication error") {
		t.Errorf("wrong error body. Expected: %s, got: %s", "authentication error", string(b))
	}
}

func TestRouteProductDetailsFailMissingProduct(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`select .* from products where id=`).WithArgs("product-id-1234").WillReturnError(errors.New("no records found"))

	r, err := http.NewRequest(http.MethodGet, "/products/product-id-1234", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	NewApp("", db).Router.ServeHTTP(w, r)

	sc := w.Result().StatusCode
	if sc != http.StatusNotFound {
		t.Errorf("wrong status code for product not found. Expected: %d, got: %d", http.StatusNotFound, sc)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRouteProductDetailsSuccess(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	prod := model.Product{
		ID:              "product-id-1234",
		Name:            "name",
		AmountAvailable: 5,
		Cost:            15,
		SellerID:        "seller-id-12",
	}

	cols := []string{"id", "name", "amount_available", "cost", "seller_id"}
	colsUser := []string{"id", "username", "password", "deposit", "role"}

	mock.ExpectQuery(`select .* from products where id=`).WithArgs("product-id-1234").
		WillReturnRows(sqlmock.NewRows(cols).AddRow(prod.ID, prod.Name, prod.AmountAvailable, prod.Cost, prod.SellerID))
	mock.ExpectQuery(`select .* from users where id=\?`).WithArgs(prod.SellerID).WillReturnRows(sqlmock.NewRows(colsUser).
		AddRow(prod.SellerID, "username", "asdjfdalfj", 100, "SELLER"))

	r, err := http.NewRequest(http.MethodGet, "/products/product-id-1234", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	NewApp("", db).Router.ServeHTTP(w, r)

	resp := w.Result()

	sc := resp.StatusCode
	if sc != http.StatusOK {
		t.Errorf("wrong status code for product found. Expected: %d, got: %d", http.StatusOK, sc)
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

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestBuyerCtxFailNoUser(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, "/buy", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	NewApp("", nil).BuyerCtx(next).ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("wrong status code")
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	msg := strings.TrimSpace(string(b))
	if msg != "authentication error" {
		t.Fatalf("wrong error message. expected: %s, got: %s", "authentication error", msg)
	}
}

func TestBuyerCtxSuccessNoCoin(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, "/buy", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		coin, ok := r.Context().Value(coinValueContextKey).(*int)
		if ok || coin != nil {
			t.Errorf("there should be no coin value set")
		}
	})

	ctx := context.WithValue(r.Context(), userContextKey, &model.User{Role: model.ROLE_BUYER})
	NewApp("", nil).BuyerCtx(next).ServeHTTP(w, r.WithContext(ctx))

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("the handler should pass the request")
	}
}

func TestBuyerCtxSuccessWithCoin(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, "/buy", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		coin, ok := r.Context().Value(coinValueContextKey).(*int)
		if !ok || coin == nil {
			t.Errorf("there should be a coin value set")
		}

		if *coin != 5 {
			t.Errorf("wrong coin value. expected: %d, got: %d", 5, *coin)
		}
	})

	ctx := context.WithValue(r.Context(), userContextKey, &model.User{Role: model.ROLE_BUYER})

	// Make Chi believe there is a URL parameter set
	routerCtx := chi.NewRouteContext()
	routerCtx.URLParams.Add("coinValue", "5")
	ctxWithChi := context.WithValue(ctx, chi.RouteCtxKey, routerCtx)

	NewApp("", nil).BuyerCtx(next).ServeHTTP(w, r.WithContext(ctxWithChi))

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("the handler should pass the request")
	}
}

func TestProductCtxFailProductDoesnotExist(t *testing.T) {

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, ok := r.Context().Value(productContextKey).(*model.Product)
		if ok || p != nil {
			t.Error("there should be no product in context")
		}
	})

	r, err := http.NewRequest(http.MethodGet, "/product", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`select .* from products where id=\?`).WithArgs("product-id").WillReturnError(errors.New("no rows"))

	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("productID", "product-id")

	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, chiCtx)

	NewApp("", db).ProductCtx(next).ServeHTTP(w, r.WithContext(ctx))

	resp := w.Result()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("wrong status code. expected: %d, got: %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestProductCtxSuccess(t *testing.T) {

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, ok := r.Context().Value(productContextKey).(*model.Product)
		if !ok || p == nil {
			t.Error("there should be a product in context")
		} else {
			if p.ID != "product-id" {
				t.Errorf("wrong product in context. expected ID: %s, got: %s", "product-id", p.ID)
			}
		}

		u, ok := r.Context().Value(sellerContextKey).(*model.User)
		if !ok || u == nil {
			t.Error("there should be a seller for the product")
		} else {
			if u.ID != "seller-id-1" {
				t.Errorf("wrong seller. expected ID: %s, got: %s", "seller-id-1", u.ID)
			}
		}
	})

	r, err := http.NewRequest(http.MethodGet, "/product", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	colsProd := []string{"id", "name", "available_amount", "cost", "seller_id"}
	colsUser := []string{"id", "username", "password", "deposit", "role"}

	mock.ExpectQuery(`select .* from products where id=\?`).WithArgs("product-id").WillReturnRows(sqlmock.NewRows(colsProd).
		AddRow("product-id", "name", 10, 5, "seller-id-1"))
	mock.ExpectQuery(`select .* from users where id=\?`).WithArgs("seller-id-1").WillReturnRows(sqlmock.NewRows(colsUser).
		AddRow("seller-id-1", "username", "asdjfdalfj", 100, "SELLER"))

	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("productID", "product-id")

	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, chiCtx)

	NewApp("", db).ProductCtx(next).ServeHTTP(w, r.WithContext(ctx))

	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("wrong status code. expected: %d, got: %d", http.StatusOK, resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestProductCtxFailNoSeller(t *testing.T) {

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, ok := r.Context().Value(productContextKey).(*model.Product)
		if !ok || p == nil {
			t.Error("there should be a product in context")
		} else {
			if p.ID != "product-id" {
				t.Errorf("wrong product in context. expected ID: %s, got: %s", "product-id", p.ID)
			}
		}
	})

	r, err := http.NewRequest(http.MethodGet, "/product", nil)
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

	mock.ExpectQuery(`select .* from products where id=\?`).WithArgs("product-id").WillReturnRows(sqlmock.NewRows(cols).
		AddRow("product-id", "name", 10, 5, "seller-id-1"))
	mock.ExpectQuery(`select .* from users where id=\?`).WithArgs("seller-id-1").WillReturnError(errors.New("no rows"))

	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("productID", "product-id")

	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, chiCtx)

	NewApp("", db).ProductCtx(next).ServeHTTP(w, r.WithContext(ctx))

	resp := w.Result()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("wrong status code. expected: %d, got: %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestProductCtxSuccessWithBadAmount(t *testing.T) {

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, ok := r.Context().Value(productContextKey).(*model.Product)
		if !ok || p == nil {
			t.Error("there should be a product in context")
		} else {
			if p.ID != "product-id" {
				t.Errorf("wrong product in context. expected ID: %s, got: %s", "product-id", p.ID)
			}
		}

		a, ok := r.Context().Value(amountValueContextKey).(*int)
		if ok || a != nil {
			t.Error("amount not valid, should not pass")
		}
	})

	r, err := http.NewRequest(http.MethodGet, "/product", nil)
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
	colsUser := []string{"id", "username", "password", "deposit", "role"}

	mock.ExpectQuery(`select .* from products where id=\?`).WithArgs("product-id").WillReturnRows(sqlmock.NewRows(cols).
		AddRow("product-id", "name", 10, 5, "seller-id-1"))
	mock.ExpectQuery(`select .* from users where id=\?`).WithArgs("seller-id-1").WillReturnRows(sqlmock.NewRows(colsUser).
		AddRow("seller-id-1", "user", "dksajfjadf", 100, "SELLER"))

	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("productID", "product-id")
	chiCtx.URLParams.Add("amount", "iuwer")

	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, chiCtx)

	NewApp("", db).ProductCtx(next).ServeHTTP(w, r.WithContext(ctx))

	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("wrong status code. expected: %d, got: %d", http.StatusOK, resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestProductCtxSuccessWithGoodAmount(t *testing.T) {

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, ok := r.Context().Value(productContextKey).(*model.Product)
		if !ok || p == nil {
			t.Error("there should be a product in context")
		} else {
			if p.ID != "product-id" {
				t.Errorf("wrong product in context. expected ID: %s, got: %s", "product-id", p.ID)
			}
		}

		a, ok := r.Context().Value(amountValueContextKey).(*int)
		if !ok || a == nil {
			t.Error("amount should be available")
		}
	})

	r, err := http.NewRequest(http.MethodGet, "/product", nil)
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
	colsUser := []string{"id", "username", "password", "deposit", "role"}

	mock.ExpectQuery(`select .* from products where id=\?`).WithArgs("product-id").WillReturnRows(sqlmock.NewRows(cols).
		AddRow("product-id", "name", 10, 5, "seller-id-1"))
	mock.ExpectQuery(`select .* from users where id=\?`).WithArgs("seller-id-1").WillReturnRows(sqlmock.NewRows(colsUser).
		AddRow("seller-id-1", "user", "dksajfjadf", 100, "SELLER"))

	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("productID", "product-id")
	chiCtx.URLParams.Add("amount", "3")

	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, chiCtx)

	NewApp("", db).ProductCtx(next).ServeHTTP(w, r.WithContext(ctx))

	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("wrong status code. expected: %d, got: %d", http.StatusOK, resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
