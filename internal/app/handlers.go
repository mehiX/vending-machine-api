package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/lestrrat-go/jwx/jwt"
	"github.com/mehiX/vending-machine-api/internal/app/model"
)

// @Summary 	Get information about current user
// @Description Fetches data from the auth token and returns it as a json object
// @Tags		private
// @Security 	ApiKeyAuth
// @Produces	application/json
// @Success 	200 {object} currentUserResponse
// @Failure		401 {string} string "not authorized"
// @Failure		400 {string} string "bad request"
// @Router 		/user [get]
func (a *app) handleShowCurrentUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		usr := r.Context().Value(userContextKey).(*model.User)

		resp := currentUserResponse{
			Username: usr.Username,
			Role:     usr.Role,
			Deposit:  usr.Deposit,
		}

		w.Header().Set("Content-type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			fmt.Println(err)
			http.Error(w, "error encoding response", http.StatusInternalServerError)
			return
		}
	}

}

// @Summary 	Add a new user
// @Description Receive user data in body, validate it and save in the database
// @Tags		public
// @Accept		application/json
// @Produces	application/json
// @Param 		request body addUserRequest true "user data"
// @Success		201
// @Failure		500 {string} string "user not created"
// @Failure		400 {string} string "bad request"
// @Router 		/user [post]
func (a *app) handleAddUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var data addUserRequest
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

// @Summary 	User login
// @Description Receive user credentials in body and return a valid token if they match a database record
// @Tags		public
// @Accept		application/json
// @Produces	text/plain
// @Param 		request body loginRequest true "user credentials"
// @Success		200 {string} string "jwt"
// @Failure		401 {string} string "not authorized"
// @Failure		400 {string} string "bad request"
// @Router 		/login [post]
func (a *app) handleLogin() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		var body loginRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		usr, err := a.FindByCredentials(r.Context(), body.Username, body.Password)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		// TODO fetch use from database and do a proper login
		tokenString, err := a.getEncTokenString(usr.ID, usr.Username)
		if err != nil {
			fmt.Printf("Error signing token: %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write([]byte(tokenString))

	}
}

const (
	jwtUsernameKey = "username"
	jwtUserIdKey   = "userID"
)

// TODO remove role and use userID and username
func (a *app) getEncTokenString(userID, username string) (string, error) {
	t := jwt.New()
	t.Set(jwt.ExpirationKey, time.Now().Add(10*time.Minute))
	t.Set(jwt.NotBeforeKey, time.Now())
	t.Set(jwtUserIdKey, userID)
	t.Set(jwtUsernameKey, username)

	claims, _ := t.AsMap(context.Background())

	_, tokenString, err := a.JwtAuth.Encode(claims)

	return tokenString, err

}

type currentUserResponse struct {
	Username string
	Role     string
	Deposit  int64
}
type addUserRequest struct {
	Username string
	Password string
	Deposit  int64
	Role     model.TypeRole
}

type loginRequest struct {
	Username string
	Password string
}
