package user

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	CreateUser(ctx context.Context, user User) error
	FindUserById(ctx context.Context, id uuid.UUID) (User, error)
	FindUserByGoogleSub(ctx context.Context, id string) (User, error)
}

type Service struct {
	repository Repository
}

func (s *Service) CreateUser(ctx context.Context) (User, error) {
	return User{}, nil
}
