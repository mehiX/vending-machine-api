package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/mehiX/vending-machine-api/internal/app/model"
)

const (
	jwtUsernameKey = "username"
	jwtUserIdKey   = "userID"
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

		usr, err := a.FindUserByCredentials(r.Context(), body.Username, body.Password)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		tokenString, err := a.getEncTokenString(usr.ID, usr.Username)
		if err != nil {
			fmt.Printf("Error signing token: %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write([]byte(tokenString))

	}
}

// @Summary 	Reset deposit
// @Description Resets a buyer's deposit to 0
// @Tags		private, only buyers
// @Security 	ApiKeyAuth
// @Produces	application/json
// @Success		200 {object} model.User "user with reset deposit"
// @Failure		500 {string} string "reset error"
// @Router 		/reset [post]
func (a *app) handleReset() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		usr, ok := ctx.Value(userContextKey).(*model.User)
		if !ok || !usr.IsBuyer() {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		if err := a.ResetDeposit(ctx, usr); err != nil {
			fmt.Println("resetDeposit error", err)
			http.Error(w, "reset failed", http.StatusInternalServerError)
			return
		}

		a.returnUserAsJson(ctx, w, usr.ID)
	}
}

// @Summary 	Deposit coins
// @Description Deposit 1 coin at a time
// @Tags		private, only buyers
// @Security 	ApiKeyAuth
// @Produces	application/json
// @Param 		coin path integer true "Coin value" Enums(5,10,20,50,100)
// @Success		200 {object} model.User "user with updated deposit"
// @Failure		500 {string} string "deposit not updated"
// @Failure		400 {string} string "bad request"
// @Router 		/deposit/{coin} [post]
func (a *app) handleDeposit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		usr, ok := ctx.Value(userContextKey).(*model.User)
		if !ok || !usr.IsBuyer() {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		coinValue, err := strconv.Atoi(chi.URLParam(r, "coinValue"))
		if err != nil {
			http.Error(w, "coin value must be a number", http.StatusBadRequest)
			return
		}

		if err := validateDepositCoin(coinValue); err != nil {
			http.Error(w, "coin value not allowed", http.StatusBadRequest)
			return
		}

		if err := a.UserDepositCoin(ctx, usr, coinValue); err != nil {
			fmt.Println("deposit coin", err)
			http.Error(w, "deposit faile", http.StatusInternalServerError)
			return
		}

		a.returnUserAsJson(ctx, w, usr.ID)
	}
}

func (a *app) handleBuy() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {}
}

func (a *app) returnUserAsJson(ctx context.Context, w http.ResponseWriter, userID string) {
	buyer, err := a.FindUserByID(ctx, userID)
	if err != nil {
		fmt.Println("error find user by id", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(buyer); err != nil {
		fmt.Println("error encoding user", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (a *app) getEncTokenString(userID, username string) (tokenString string, err error) {
	t := jwt.New()
	t.Set(jwt.ExpirationKey, time.Now().Add(10*time.Minute))
	t.Set(jwt.NotBeforeKey, time.Now())
	t.Set(jwtUserIdKey, userID)
	t.Set(jwtUsernameKey, username)

	claims, err := t.AsMap(context.Background())
	if err != nil {
		return
	}

	_, tokenString, err = a.JwtAuth.Encode(claims)

	return

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
