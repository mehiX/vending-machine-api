package app

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/mehiX/vending-machine-api/internal/app/model"
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

	vm := NewApp("", nil)
	router := vm.Router

	type reqBody struct {
		Username string
		Password string
	}

	body := reqBody{"mihai", "password"}

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
	if usr, ok := claims["user"]; !ok || usr != "mihai" {
		t.Error("Wrong or missing claim 'user'")
	}
	if role, ok := claims["role"]; !ok || role != model.ROLE_BUYER {
		t.Errorf("Wrong or missing claim 'role'. Expected: %s, got: %s", model.ROLE_BUYER, role)
	}
}
