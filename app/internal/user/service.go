package user

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type Repository interface {
	CreateUser(ctx context.Context, user *User) (*User, error)
	FindUserByID(ctx context.Context, id uuid.UUID) (User, error)
	FindUserByGoogleSub(ctx context.Context, id string) (User, error)
}

type Service struct {
	repository Repository
}

func NewService(r Repository) *Service {
	return &Service{repository: r}
}

type FindOrCreateInput struct {
	GoogleSub  string
	Email      string
	PictureURL string
}

func (s *Service) FindOrCreate(ctx context.Context, input FindOrCreateInput) (User, error) {
	user, err := s.repository.FindUserByGoogleSub(ctx, input.GoogleSub)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, ErrUserNotFound) {
		return User{}, err
	}

	candidate := User{
		GoogleSub: input.GoogleSub,
		Email:     input.Email,
	}

	if input.PictureURL != "" {
		candidate.Picture = &input.PictureURL
	}

	created, err := s.repository.CreateUser(ctx, &candidate)
	if err != nil {
		return User{}, err
	}

	return *created, nil
}

func (s *Service) FindUserByID(ctx context.Context, id uuid.UUID) (User, error) {
	return s.repository.FindUserByID(ctx, id)
}

func (s *Service) FindUserByGoogleSub(ctx context.Context, id string) (User, error) {
	return s.repository.FindUserByGoogleSub(ctx, id)
}
