package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mehiX/vending-machine-api/internal/app/model"
	"golang.org/x/crypto/bcrypt"
)

func TestGetUserData(t *testing.T) {

	os.Setenv("JWT_SIGNKEY", "some key")
	os.Setenv("JWT_ALG", "HS256")

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

	type response struct {
		Username string
		Role     string
	}

	r, err := http.NewRequest(http.MethodGet, "/user", nil)
	if err != nil {
		t.Fatal(err)
	}

	tknStr, err := vm.getEncTokenString(testUser, testRole)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Authorization", "BEARER "+tknStr)

	w := httptest.NewRecorder()

	vm.Router.ServeHTTP(w, r)

	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /user wrong response code. Expected: %d, got: %d", http.StatusOK, resp.StatusCode)
	}

	defer resp.Body.Close()

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
