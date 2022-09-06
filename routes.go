package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
)

var tokenAuth = jwtauth.New("HS256", []byte(uuid.New().String()), nil)

func (a *app) SetupRoutes() {

	if a.Router == nil {
		a.Router = chi.NewMux()
	}

	a.Router.Use(middleware.RequestID)
	a.Router.Use(middleware.RealIP)
	a.Router.Use(middleware.Logger)
	a.Router.Use(middleware.Timeout(60 * time.Second))

	//protected routes
	a.Router.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(jwtauth.Authenticator)
		r.Get("/validate", a.handleValidate)
	})

	// public routes
	a.Router.Group(func(r chi.Router) {
		r.Get("/health", a.handleHealth)
		r.Post("/login", a.handleLogin)
	})

}

func (a *app) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func (a *app) handleValidate(w http.ResponseWriter, r *http.Request) {

}

func (a *app) handleLogin(w http.ResponseWriter, r *http.Request) {

	type req struct {
		Username string
		Password string
	}

	var body req
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, tokenString, _ := tokenAuth.Encode(map[string]interface{}{"user": body.Username})

	w.Write([]byte(tokenString))

}
