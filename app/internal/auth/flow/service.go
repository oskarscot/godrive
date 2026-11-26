package flow

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/oskarscot/godrive/internal/util"
)

type Repository interface {
	CreateFlow(ctx context.Context, flow *Flow) (*Flow, error)
	CompleteFlow(ctx context.Context, token string, state string) (Flow, error)
}

type Service struct {
	repository Repository
}

func NewFlowService(r Repository) *Service {
	return &Service{repository: r}
}

func (s *Service) CreateFlow(ctx context.Context, state string, verifier string) (*Flow, string, error) {
	if state == "" {
		return nil, "", fmt.Errorf("invalid state")
	}
	token, hash, err := util.GenerateToken()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	candidate := Flow{
		TokenHash:     hex.EncodeToString(hash[:]),
		ExpiresAt:     time.Now().UTC().Add(5 * time.Minute),
		ExpectedState: state,
		PKCEVerifier:  verifier,
	}
	flow, err := s.repository.CreateFlow(ctx, &candidate)
	if err != nil {
		return nil, "", fmt.Errorf("unable to create auth flow: %w", err)
	}

	return flow, token, nil
}

func (s *Service) CompleteFlow(ctx context.Context, token string, state string) (Flow, error) {
	hash := sha256.Sum256([]byte(token))

	return s.repository.CompleteFlow(
		ctx,
		hex.EncodeToString(hash[:]),
		state,
	)
}
