package auth

import (
	"errors"
	"time"
)

var (
	ErrInvalidFlow    = errors.New("invalid login flow")
	ErrGoogleProvider = errors.New("Google authentication failed")
)

type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type BeginGoogleLoginResult struct {
	AuthorizationURL string
	FlowToken        string
	ExpiresAt        time.Time
}

type CompleteGoogleLoginResult struct {
	SessionToken string
	// FlowConsumed remains true if an operation fails after consuming the flow.
	FlowConsumed bool
}
