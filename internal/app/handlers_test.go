package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/mehiX/vending-machine-api/internal/app/model"
)

func TestGetUserData(t *testing.T) {

	os.Setenv("JWT_SIGNKEY", "some key")
	os.Setenv("JWT_ALG", "HS256")

	vm := NewApp("", nil)

	type response struct {
		Username string
		Role     string
	}

	r, err := http.NewRequest(http.MethodGet, "/user", nil)
	if err != nil {
		t.Fatal(err)
	}

	tknStr, err := vm.getEncTokenString("mihai", model.ROLE_USER)
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

	if respData.Username != "mihai" {
		t.Errorf("GET /user wrong username in response. Expected: mihai, got: %s", respData.Username)
	}

	if respData.Role != model.ROLE_USER {
		t.Errorf("GET /user wrong role in response. Expected: %s, got: %s", model.ROLE_USER, respData.Role)
	}
}
