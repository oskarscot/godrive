package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oskarscot/godrive/internal/auth"
	httpx "github.com/oskarscot/godrive/internal/platform"
	"github.com/oskarscot/godrive/internal/user"
)

func (a *application) mount(ctx context.Context) (http.Handler, error) {
	mux := chi.NewMux()

	mux.Use(middleware.Logger)

	authService, err := auth.NewAuthService(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to setup AuthService: %v\n", err)
	}

	authHandler := auth.NewAuthHandler(authService)
	authHandler.Setup(mux)

	userHandler := user.NewHandler()
	userHandler.Setup(mux)

	mux.NotFound(notFoundHandler)
	mux.MethodNotAllowed(methodNotAllowedHandler)

	return mux, nil
}

func (a *application) start(h http.Handler) error {
	server := http.Server{
		Addr:    a.config.address,
		Handler: h,
	}

	return server.ListenAndServe()
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	httpx.Error(w, r, http.StatusNotFound, "not found", fmt.Sprintf("route %s does not exist", r.URL.String()))
}

func methodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	httpx.Error(w, r, http.StatusNotFound, "method not allowed", fmt.Sprintf("method %s is not allowed for %s", r.Method, r.URL.String()))
}

type application struct {
	config *config
	logger *log.Logger
	conn   *pgxpool.Pool
}

type config struct {
	address string
}
