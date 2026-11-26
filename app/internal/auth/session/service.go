package session

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/oskarscot/godrive/internal/user"
	"github.com/oskarscot/godrive/internal/util"
)

const CookieName = "__Host-session"

type Repository interface {
	CreateSession(ctx context.Context, session *Session) (*Session, error)
	GetSessionByHash(ctx context.Context, hash string) (Session, error)
	DeleteSessionsForUser(ctx context.Context, id uuid.UUID) error
}

type Service struct {
	repository Repository
}

func NewSessionService(r Repository) *Service {
	return &Service{repository: r}
}

func (s *Service) FindSessionFromToken(ctx context.Context, token string) (Session, error) {
	hash := sha256.Sum256([]byte(token))

	return s.repository.GetSessionByHash(ctx, hex.EncodeToString(hash[:]))
}

func (s *Service) CreateSession(ctx context.Context, user user.User) (string, error) {
	token, hash, err := util.GenerateToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	expiry := time.Now().UTC().Add(15 * time.Minute)
	session := Session{
		TokenHash: hex.EncodeToString(hash[:]),
		UserID:    user.ID,
		ExpiresAt: &expiry,
	}
	_, err = s.repository.CreateSession(ctx, &session)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	return token, nil
}

func (s *Service) DeleteSessionsForUser(ctx context.Context, id uuid.UUID) error {
	return s.repository.DeleteSessionsForUser(ctx, id)
}
