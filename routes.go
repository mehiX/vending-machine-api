package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (a *app) SetupRoutes() {

	if a.Router == nil {
		a.Router = chi.NewMux()
	}

	a.Router.Get("/health", a.handleHealth)

}

func (a *app) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
