package session

import (
	"errors"
	"fmt"
	"net/http"

	httpx "github.com/oskarscot/godrive/internal/platform"
	"github.com/oskarscot/godrive/internal/requestctx"
)

func (s *Service) RequireSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(CookieName)
		if err != nil || cookie.Value == "" {
			httpx.Error(w, r, http.StatusUnauthorized, "unauthorized", "invalid credentials")
			return
		}

		ctx := r.Context()
		currentSession, err := s.FindSessionFromToken(ctx, cookie.Value)
		if err != nil {
			if errors.Is(err, ErrSessionNotFound) {
				httpx.Error(w, r, http.StatusUnauthorized, "unauthorized", "invalid session")
				return
			}
			fmt.Printf("Error: %v\n", err)
			httpx.Error(w, r, http.StatusInternalServerError, "internal server error", "unable to validate session")
			return
		}

		ctx = requestctx.WithUserID(ctx, currentSession.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
