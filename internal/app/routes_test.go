package app

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mehiX/vending-machine-api/internal/app/model"
	"golang.org/x/crypto/bcrypt"
)

func TestHealth(t *testing.T) {

	r, err := http.NewRequest(http.MethodGet, "/health", nil)
	if err != nil {
		t.Error(err)
	}
	w := httptest.NewRecorder()

	NewApp("", nil).Router.ServeHTTP(w, r)

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
