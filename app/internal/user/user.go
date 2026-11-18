package user

import "github.com/google/uuid"

type User struct {
	ID        uuid.UUID
	GoogleSub string
	Email     string
	Picture   string
}
