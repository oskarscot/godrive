package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func (a *application) mount() http.Handler {
	mux := chi.NewMux()

	mux.Use(middleware.Logger)

	return mux
}

func (a *application) start(h http.Handler) error {
	server := http.Server{
		Addr:    a.config.address,
		Handler: h,
	}

	return server.ListenAndServe()
}

type application struct {
	config *config
	conn   *pgxpool.Pool
}

type config struct {
	address string
}
