package main

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type app struct {
	Addr   string
	Router *chi.Mux
	Db     *sql.DB
}

func NewApp(addr string, db *sql.DB) *app {
	a := &app{
		Addr: addr,
		Db:   nil,
	}

	a.SetupRoutes()

	return a
}

func (a *app) HttpServer() http.Server {
	return http.Server{
		Addr:    a.Addr,
		Handler: a.Router,
	}
}
