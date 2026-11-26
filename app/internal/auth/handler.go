package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/oskarscot/godrive/internal/auth/flow"
	"github.com/oskarscot/godrive/internal/auth/session"
	"github.com/oskarscot/godrive/internal/httpx"
	"github.com/oskarscot/godrive/internal/requestctx"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Setup(r chi.Router) {
	r.Get("/auth/google", h.startGoogleLogin)
	r.Get("/auth/google/callback", h.googleCallbackHandler)
}

func (h *Handler) SetupProtected(r chi.Router) {
	r.Post("/auth/logout", h.logoutHandler)
}

func (h *Handler) startGoogleLogin(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.BeginGoogleLogin(r.Context())
	if err != nil {
		httpx.Error(w, r, http.StatusInternalServerError, "server error", "could not start login")
		return
	}

	cookie := newAuthCookie(flow.AuthFlowCookie, result.FlowToken, int(time.Until(result.ExpiresAt).Seconds()))
	http.SetCookie(w, cookie)
	http.Redirect(w, r, result.AuthorizationURL, http.StatusTemporaryRedirect)
}

func (h *Handler) googleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" || state == "" {
		httpx.Error(w, r, http.StatusBadRequest, "bad request", "code and state are required")
		return
	}

	flowCookie, err := r.Cookie(flow.AuthFlowCookie)
	if err != nil || flowCookie.Value == "" {
		httpx.Error(w, r, http.StatusUnauthorized, "unauthorized", "missing login flow")
		return
	}

	result, err := h.service.CompleteGoogleLogin(r.Context(), flowCookie.Value, state, code)
	if result.FlowConsumed {
		http.SetCookie(w, newAuthCookie(flow.AuthFlowCookie, "", -1))
	}
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidFlow):
			httpx.Error(w, r, http.StatusUnauthorized, "unauthorized", "invalid or expired login flow")
		case errors.Is(err, ErrGoogleProvider):
			httpx.Error(w, r, http.StatusBadGateway, "authentication error", "unable to authenticate with Google")
		default:
			httpx.Error(w, r, http.StatusInternalServerError, "server error", "unable to complete login")
		}
		return
	}

	http.SetCookie(w, newAuthCookie(session.CookieName, result.SessionToken, 0))
	http.Redirect(w, r, "/user/me", http.StatusSeeOther)
}

func (h *Handler) logoutHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := requestctx.UserIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, r, http.StatusInternalServerError, "server error", "authenticated user is missing")
		return
	}

	if err := h.service.Logout(r.Context(), userID); err != nil {
		httpx.Error(w, r, http.StatusInternalServerError, "server error", "unable to delete sessions")
		return
	}

	http.SetCookie(w, newAuthCookie(session.CookieName, "", -1))
	httpx.JSON(w, http.StatusOK, "successfully logged out")
}

func newAuthCookie(name, value string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
}
