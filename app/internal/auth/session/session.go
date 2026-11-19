package session

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrSessionNotFound = errors.New("session not found")
)

type Session struct {
	ID        uuid.UUID
	TokenHash string
	UserID    uuid.UUID
	CreatedAt *time.Time
	ExpiresAt *time.Time
}
