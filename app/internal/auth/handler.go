package auth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	httpx "github.com/oskarscot/godrive/internal/platform"
	"github.com/oskarscot/godrive/internal/user"
	"golang.org/x/oauth2"
)

type Handler struct {
	svc     *AuthService
	userSvc *user.Service
}

func NewAuthHandler(svc *AuthService, userSvc *user.Service) *Handler {
	return &Handler{svc: svc, userSvc: userSvc}
}

func (h *Handler) Setup(mux *chi.Mux) {
	mux.Get("/auth/google", h.startGoogleLogin)
	mux.Get("/auth/google/callback", h.googleCallbackHandler)
}

func (h *Handler) startGoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := h.svc.GoogleConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *Handler) googleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		httpx.Error(w, r, http.StatusBadRequest, "bad request", "code is missing from the URL path")
		return
	}

	ctx := r.Context()

	token, err := h.svc.GoogleConfig.Exchange(ctx, code)
	if err != nil {
		httpx.Error(w, r, http.StatusInternalServerError, "authentication error", "unable to exchange token with provider")
		return
	}

	client := h.svc.GoogleConfig.Client(ctx, token)

	resp, err := client.Get(GoogleUserInfo)

	if err != nil {
		httpx.Error(w, r, http.StatusInternalServerError, "authentication error", "unable to retrieve user data from the provider")
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		httpx.Error(w, r, http.StatusBadGateway, "authentication error", "provider returned an unsuccessful user info response")
		return
	}

	var googleUser GoogleUser
	if err = json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		httpx.Error(w, r, http.StatusUnprocessableEntity, "authentication error", "unable to parse user data from the provider")

		return
	}

	if googleUser.ID == "" {
		httpx.Error(w, r, http.StatusBadGateway, "authentication error", "provider response is missing the user identity")
		return
	}

	user, err := h.userSvc.FindOrCreate(ctx, user.FindOrCreateInput{
		GoogleSub:  googleUser.ID,
		Email:      googleUser.Email,
		PictureURL: googleUser.PictureURL,
	})
	if err != nil {
		httpx.Error(w, r, http.StatusInternalServerError, "authentication error", "unable to find or create user")
		return
	}

	opaque, err := h.svc.SessionService.CreateSession(ctx, user)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		httpx.Error(w, r, http.StatusInternalServerError, "authentication error", "failed to create session")
		return
	}

	cookie := http.Cookie{
		Name:     "__Host-session",
		Value:    opaque,
		Path:     "/",
		MaxAge:   0,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &cookie)
}
