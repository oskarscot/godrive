package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ Repository = (*PostgresRepository)(nil)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) CreateUser(ctx context.Context, user *User) (*User, error) {
	const query = `
			INSERT INTO users (google_sub, email, picture)
			VALUES ($1, $2, $3)
			RETURNING id, created_at, updated_at
		`

	err := r.pool.QueryRow(
		ctx,
		query,
		user.GoogleSub,
		user.Email,
		user.Picture,
	).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func (r *PostgresRepository) FindUserByID(ctx context.Context, id uuid.UUID) (User, error) {
	const query = `
		SELECT id, google_sub, email, picture, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.GoogleSub,
		&user.Email,
		&user.Picture,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, fmt.Errorf("%w: %v", ErrUserNotFound, id)
		}
		return User{}, fmt.Errorf("find user by ID %s: %w", id, err)
	}

	return user, nil
}

func (r *PostgresRepository) FindUserByGoogleSub(ctx context.Context, id string) (User, error) {
	query := `
	SELECT id, google_sub, email, picture, created_at, updated_at
	FROM users
	WHERE google_sub = $1
	`

	var user User
	err := r.pool.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&user.ID,
		&user.GoogleSub,
		&user.Email,
		&user.Picture,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, fmt.Errorf("%w: %v", ErrUserNotFound, id)
		}
		return User{}, fmt.Errorf("find user by ID %s: %w", id, err)
	}

	return user, nil
}
