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
		Role     string
	}

	return func(w http.ResponseWriter, r *http.Request) {
		_, claims, err := jwtauth.FromContext(r.Context())
		if err != nil {
			http.Error(w, "authentication error", http.StatusUnauthorized)
			return
		}

		resp := response{
			Username: claims["user"].(string),
			Role:     claims["role"].(string),
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
		Role     model.TypeRole
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var data request
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			fmt.Println("addUser error", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if err := a.CreateUser(r.Context(), data.Username, data.Password, data.Deposit, data.Role); err != nil {
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

		// TODO fetch use from database and do a proper login
		tokenString, err := a.getEncTokenString(body.Username, model.ROLE_USER)
		if err != nil {
			fmt.Printf("Error signing token: %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write([]byte(tokenString))

	}
}

func (a *app) getEncTokenString(username string, role model.TypeRole) (string, error) {
	t := jwt.New()
	t.Set(jwt.ExpirationKey, time.Now().Add(10*time.Minute))
	t.Set(jwt.NotBeforeKey, time.Now())
	t.Set("user", username)
	t.Set("role", role)

	claims, _ := t.AsMap(context.Background())

	_, tokenString, err := a.JwtAuth.Encode(claims)

	return tokenString, err

}
