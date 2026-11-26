package flow

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

const AuthFlowCookie = "auth-flow-state"

var (
	ErrFlowExpired  = errors.New("auth flow expired")
	ErrFlowNotFound = errors.New("auth flow does not exist")
)

type Flow struct {
	ID            uuid.UUID
	TokenHash     string
	ExpectedState string
	PKCEVerifier  string
	CreateAt      time.Time
	ExpiresAt     time.Time
}
