package app

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"

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
		r.Get("/user", a.handleShowCurrentUser())
	})

	// public routes
	a.Router.Group(func(r chi.Router) {
		r.Get("/health", a.handleHealth)
		r.Post("/login", a.handleLogin())
		r.Post("/user", a.handleAddUser())
	})

	a.Router.Mount("/swagger", swg.WrapHandler)
}

// @Summary 	Health endpoing
// @Description Validate the application is running
// @Tags		public
// @Produces	text/plain
// @Success		200 {string} string "OK"
// @Router 		/health [get]
func (a *app) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
