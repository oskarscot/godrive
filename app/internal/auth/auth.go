package auth

import (
	"context"
	"fmt"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/oskarscot/godrive/internal/auth/session"
	"golang.org/x/oauth2"
)

const (
	GoogleUserInfo     string = "https://openidconnect.googleapis.com/v1/userinfo"
	GoogleProfileScope string = "https://www.googleapis.com/auth/userinfo.profile"
	GoogleEmailScope   string = "https://www.googleapis.com/auth/userinfo.email"
	GoogleTokenUrl     string = "https://oauth2.googleapis.com/token"
	GoogleProviderUrl  string = "https://accounts.google.com"
	AuthCookie         string = session.CookieName
)

type GoogleUser struct {
	ID            string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	FamilyName    string `json:"family_name"`
	GivenName     string `json:"given_name"`
	Name          string `json:"name"`
	PictureURL    string `json:"picture"`
}

type AuthService struct {
	GoogleConfig   oauth2.Config
	GoogleProvider *oidc.Provider
	GoogleVerifier *oidc.IDTokenVerifier
	SessionService *session.SessionService
}

func NewAuthService(ctx context.Context, svc *session.SessionService) (*AuthService, error) {
	provider, err := oidc.NewProvider(ctx, GoogleProviderUrl)
	if err != nil {
		return &AuthService{}, fmt.Errorf("Failed to setup AuthService: %w", err)
	}

	config := oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Endpoint:     provider.Endpoint(),
		Scopes: []string{
			oidc.ScopeOpenID,
			GoogleProfileScope,
			GoogleEmailScope,
		},
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: config.ClientID})

	return &AuthService{
		GoogleConfig:   config,
		GoogleProvider: provider,
		GoogleVerifier: verifier,
		SessionService: svc,
	}, nil
}
