package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/oskarscot/godrive/internal/auth/flow"
	httpx "github.com/oskarscot/godrive/internal/platform"
	"github.com/oskarscot/godrive/internal/requestctx"
	"github.com/oskarscot/godrive/internal/user"
	"github.com/oskarscot/godrive/internal/util"
	"golang.org/x/oauth2"
)

type Handler struct {
	svc     *AuthService
	userSvc *user.Service
}

func NewAuthHandler(svc *AuthService, userSvc *user.Service) *Handler {
	return &Handler{svc: svc, userSvc: userSvc}
}

func (h *Handler) Setup(mux chi.Router) {
	mux.Get("/auth/google", h.startGoogleLogin)
	mux.Get("/auth/google/callback", h.googleCallbackHandler)
	mux.With(h.svc.SessionService.RequireSession).Get("/auth/logout", h.logoutHandler)
}

func (h *Handler) startGoogleLogin(w http.ResponseWriter, r *http.Request) {
	state, err := util.RandomToken()
	if err != nil {
		httpx.Error(w, r, http.StatusInternalServerError, "server error", "could not generate token, try again")
		return
	}

	ctx := r.Context()
	verifier := oauth2.GenerateVerifier()
	created, token, err := h.svc.AuthFlowService.CreateFlow(ctx, state, verifier)
	if err != nil {
		httpx.Error(w, r, http.StatusInternalServerError, "server error", "could not start auth flow")
		return
	}

	cookie := &http.Cookie{
		Name:     flow.AuthFlowCookie,
		Value:    token,
		Path:     "/",
		MaxAge:   int(time.Until(created.ExpiresAt).Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, cookie)

	url := h.svc.GoogleConfig.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier))
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *Handler) googleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		httpx.Error(w, r, http.StatusBadRequest, "bad request", "code is missing from the URL path")
		return
	}

	state := r.URL.Query().Get("state")
	if state == "" {
		httpx.Error(w, r, http.StatusBadRequest, "bad request", "state is missing from the URL path")
		return
	}

	ctx := r.Context()

	flowCookie, err := r.Cookie(flow.AuthFlowCookie)
	if err != nil {
		httpx.Error(w, r, http.StatusUnauthorized, "unauthorized", "invalid credentials")
		return
	}

	authFlow, err := h.svc.AuthFlowService.CompleteFlow(ctx, flowCookie.Value, state)
	if err != nil {
		httpx.Error(w, r, http.StatusUnauthorized, "unauthorized", "unable to find authentication flow for this session")
		return
	}

	cookie := http.Cookie{
		Name:     flow.AuthFlowCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &cookie)

	token, err := h.svc.GoogleConfig.Exchange(ctx, code, oauth2.VerifierOption(authFlow.PKCEVerifier))
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

	sessionCookie := &http.Cookie{
		Name:     AuthCookie,
		Value:    opaque,
		Path:     "/",
		MaxAge:   0,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, sessionCookie)
	http.Redirect(w, r, "/user/me", http.StatusSeeOther)
}

func (h *Handler) logoutHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, ok := requestctx.UserIDFromContext(ctx)
	if !ok {
		httpx.Error(w, r, http.StatusUnauthorized, "unauthorized", "no session found")
		return
	}

	_, err := r.Cookie(AuthCookie)
	if err != nil {
		httpx.Error(w, r, http.StatusBadRequest, "invalid cookie", "missing authorization cookie")
		return
	}

	err = h.svc.SessionService.DeleteSessionsForUser(ctx, id)
	if err != nil {
		httpx.Error(w, r, http.StatusInternalServerError, "internal error", "unable to delete sessions")
		return
	}

	cookie := http.Cookie{
		Name:     AuthCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &cookie)

	httpx.JSON(w, http.StatusOK, "successfully logged out")
}
