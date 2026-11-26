package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oskarscot/godrive/internal/auth"
	"github.com/oskarscot/godrive/internal/auth/flow"
	"github.com/oskarscot/godrive/internal/auth/session"
	"github.com/oskarscot/godrive/internal/httpx"
	"github.com/oskarscot/godrive/internal/user"
)

func (a *application) mount(ctx context.Context, pool *pgxpool.Pool) (http.Handler, error) {
	mux := chi.NewMux()

	mux.Use(middleware.Logger)

	sessionRepository := session.NewSessionRepository(pool)
	sessionService := session.NewSessionService(sessionRepository)

	flowRepository := flow.NewFlowRepository(pool)
	flowService := flow.NewFlowService(flowRepository)

	userRepository := user.NewPostgresRepository(pool)
	userService := user.NewService(userRepository)
	userHandler := user.NewHandler(userService)

	authConfig := auth.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
	}
	authService, err := auth.NewService(ctx, authConfig, userService, sessionService, flowService)
	if err != nil {
		return nil, fmt.Errorf("set up authentication: %w", err)
	}

	authHandler := auth.NewHandler(authService)
	authHandler.Setup(mux)

	mux.Group(func(r chi.Router) {
		r.Use(sessionService.RequireSession)
		userHandler.Setup(r)
		authHandler.SetupProtected(r)
	})

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
