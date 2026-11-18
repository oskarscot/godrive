package user

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

var _ Repository = (*PostgresRepository)(nil)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) CreateUser(ctx context.Context, user User) error {
	return nil
}
