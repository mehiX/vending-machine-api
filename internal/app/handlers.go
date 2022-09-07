package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/mehiX/vending-machine-api/internal/app/model"
)

func (a *app) handleShowCurrentUser() http.HandlerFunc {
	type response struct {
		Username string
		Role     int
	}

	return func(w http.ResponseWriter, r *http.Request) {
		_, claims, err := jwtauth.FromContext(r.Context())
		if err != nil {
			http.Error(w, "authentication error", http.StatusUnauthorized)
			return
		}

		resp := response{
			Username: claims["user"].(string),
			Role:     int(claims["role"].(float64)),
		}

		w.Header().Set("Content-type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			fmt.Println(err)
			http.Error(w, "error encoding response", http.StatusInternalServerError)
			return
		}
	}

}

func (a *app) handleAddUser() http.HandlerFunc {

	type request struct {
		Username string
		Password string
		Deposit  int64
		Role     int
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var data request
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			fmt.Println("addUser error", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if err := a.CreateUser(data.Username, data.Password, data.Deposit, model.TypeRole(data.Role)); err != nil {
			fmt.Println("createUser error", err)
			http.Error(w, "user not created", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func (a *app) handleLogin() http.HandlerFunc {

	type req struct {
		Username string
		Password string
	}

	return func(w http.ResponseWriter, r *http.Request) {

		var body req
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		t := jwt.New()
		t.Set(jwt.ExpirationKey, time.Now().Add(10*time.Minute))
		t.Set(jwt.NotBeforeKey, time.Now())
		t.Set("user", body.Username)
		t.Set("role", model.ROLE_USER)

		claims, _ := t.AsMap(context.Background())

		_, tokenString, err := a.JwtAuth.Encode(claims)
		if err != nil {
			fmt.Printf("Error signing token: %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write([]byte(tokenString))

	}
}
