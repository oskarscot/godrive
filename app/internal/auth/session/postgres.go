package session

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ Repository = (*SessionRepository)(nil)

type SessionRepository struct {
	pool *pgxpool.Pool
}

func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{pool: pool}
}

func (r *SessionRepository) CreateSession(ctx context.Context, session *Session) (*Session, error) {
	query := `
	INSERT INTO sessions (token_hash, user_id, expires_at)
	VALUES ($1, $2, $3)
	RETURNING id, created_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		session.TokenHash,
		session.UserID,
		session.ExpiresAt,
	).Scan(
		&session.ID,
		&session.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	return session, nil
}

func (r *SessionRepository) GetSessionByHash(ctx context.Context, hash string) (Session, error) {
	query := `
	SELECT id, token_hash, user_id, created_at, expires_at
	FROM sessions
	WHERE token_hash = $1
	AND expires_at > now();
	`

	var session Session
	err := r.pool.QueryRow(
		ctx,
		query,
		hash,
	).Scan(
		&session.ID,
		&session.TokenHash,
		&session.UserID,
		&session.CreatedAt,
		&session.ExpiresAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Session{}, fmt.Errorf("%w: %v", ErrSessionNotFound, hash)
		}
		return Session{}, err
	}
	return session, nil
}
