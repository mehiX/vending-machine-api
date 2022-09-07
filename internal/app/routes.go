package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/jwt"

	_ "github.com/mehiX/vending-machine-api/docs"
	swg "github.com/swaggo/http-swagger"
)

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
		r.Use(jwtauth.Verifier(a.JwtAuth))
		r.Use(jwtauth.Authenticator)
		r.Get("/validate", a.handleValidate)
	})

	// public routes
	a.Router.Group(func(r chi.Router) {
		r.Get("/health", a.handleHealth)
		r.Post("/login", a.handleLogin())
	})

	a.Router.Mount("/swagger", swg.WrapHandler)
}

// @Summary 	Health endpoing
// @Description Validate the application is running
// @Produces	text/plain
// @Success		200 {string} string "OK"
// @Router 		/health [get]
func (a *app) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func (a *app) handleValidate(w http.ResponseWriter, r *http.Request) {

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
		t.Set("role", ROLE_USER)

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
