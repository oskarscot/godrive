package flow

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ Repository = (*FlowRepository)(nil)

type FlowRepository struct {
	pool *pgxpool.Pool
}

func NewFlowRepository(pool *pgxpool.Pool) *FlowRepository {
	return &FlowRepository{pool: pool}
}

func (r *FlowRepository) CompleteFlow(ctx context.Context, tokenHash string, state string) (Flow, error) {
	query := `
	DELETE FROM auth_flow
	WHERE token_hash = $1
    AND expected_state = $2
    AND expires_at > now()
	RETURNING id, expected_state, verifier, created_at, expires_at;
	`

	var flow Flow
	err := r.pool.QueryRow(
		ctx,
		query,
		tokenHash,
		state,
	).Scan(
		&flow.ID,
		&flow.ExpectedState,
		&flow.PKCEVerifier,
		&flow.CreatedAt,
		&flow.ExpiresAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Flow{}, ErrFlowNotFound
		}
		return Flow{}, fmt.Errorf("complete auth flow: %w", err)
	}

	return flow, nil
}

func (r *FlowRepository) CreateFlow(ctx context.Context, flow *Flow) (*Flow, error) {
	query := `
	INSERT INTO auth_flow (token_hash, expected_state, verifier, expires_at)
	VALUES ($1, $2, $3, $4)
	RETURNING id, created_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		flow.TokenHash,
		flow.ExpectedState,
		flow.PKCEVerifier,
		flow.ExpiresAt,
	).Scan(
		&flow.ID,
		&flow.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create auth flow: %w", err)
	}

	return flow, nil
}
