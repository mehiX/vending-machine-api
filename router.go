package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func router() http.Handler {
	r := chi.NewMux()

	return r
}
