package session

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/oskarscot/godrive/internal/user"
)

type Repository interface {
	CreateSession(ctx context.Context, session *Session) (*Session, error)
	GetSessionByHash(ctx context.Context, hash string) (Session, error)
	DeleteSessionsForUser(ctx context.Context, id uuid.UUID) error
}

type SessionService struct {
	repository Repository
}

func NewSessionService(r Repository) *SessionService {
	return &SessionService{repository: r}
}

func (s *SessionService) FindSessionFromToken(ctx context.Context, token string) (Session, error) {
	hash := sha256.Sum256([]byte(token))

	return s.repository.GetSessionByHash(ctx, hex.EncodeToString(hash[:]))
}

func (s *SessionService) CreateSession(ctx context.Context, user user.User) (string, error) {
	token, hash, err := generateSessionToken()
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

func (s *SessionService) DeleteSessionsForUser(ctx context.Context, id uuid.UUID) error {
	return s.repository.DeleteSessionsForUser(ctx, id)
}

func generateSessionToken() (string, [32]byte, error) {
	var randomBytes [32]byte
	if _, err := rand.Read(randomBytes[:]); err != nil {
		return "", [32]byte{}, err
	}

	token := base64.RawURLEncoding.EncodeToString(randomBytes[:])
	hash := sha256.Sum256([]byte(token))

	return token, hash, nil
}
