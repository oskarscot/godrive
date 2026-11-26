package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

const (
	googleIssuerURL    = "https://accounts.google.com"
	googleUserInfoURL  = "https://openidconnect.googleapis.com/v1/userinfo"
	googleProfileScope = "https://www.googleapis.com/auth/userinfo.profile"
	googleEmailScope   = "https://www.googleapis.com/auth/userinfo.email"
)

type googleUser struct {
	ID         string `json:"sub"`
	Email      string `json:"email"`
	PictureURL string `json:"picture"`
}

func newGoogleConfig(ctx context.Context, config Config) (oauth2.Config, error) {
	if config.ClientID == "" || config.ClientSecret == "" || config.RedirectURL == "" {
		return oauth2.Config{}, fmt.Errorf("Google client ID, client secret, and redirect URL are required")
	}

	provider, err := oidc.NewProvider(ctx, googleIssuerURL)
	if err != nil {
		return oauth2.Config{}, fmt.Errorf("discover Google provider: %w", err)
	}

	return oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes: []string{
			oidc.ScopeOpenID,
			googleProfileScope,
			googleEmailScope,
		},
	}, nil
}

func (s *Service) fetchGoogleUser(ctx context.Context, token *oauth2.Token) (googleUser, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, googleUserInfoURL, nil)
	if err != nil {
		return googleUser{}, fmt.Errorf("create Google userinfo request: %w", err)
	}

	client := s.googleConfig.Client(ctx, token)
	response, err := client.Do(request)
	if err != nil {
		return googleUser{}, fmt.Errorf("fetch Google userinfo: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return googleUser{}, fmt.Errorf("Google userinfo returned HTTP %d", response.StatusCode)
	}

	var profile googleUser
	if err := json.NewDecoder(response.Body).Decode(&profile); err != nil {
		return googleUser{}, fmt.Errorf("decode Google userinfo: %w", err)
	}
	if profile.ID == "" {
		return googleUser{}, fmt.Errorf("Google userinfo is missing the subject")
	}

	return profile, nil
}
