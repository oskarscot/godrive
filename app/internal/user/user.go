package user

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type User struct {
	ID        uuid.UUID
	GoogleSub string
	Email     string
	Picture   *string
	CreatedAt time.Time
	UpdatedAt time.Time
}
