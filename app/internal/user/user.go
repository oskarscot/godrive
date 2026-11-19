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
	ID        uuid.UUID `json:"id"`
	GoogleSub string    `json:"google_sub"`
	Email     string    `json:"email"`
	Picture   *string   `json:"picture"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
