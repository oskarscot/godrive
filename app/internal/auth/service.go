package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/oskarscot/godrive/internal/auth/flow"
	"github.com/oskarscot/godrive/internal/auth/session"
	"github.com/oskarscot/godrive/internal/user"
	"github.com/oskarscot/godrive/internal/util"
	"golang.org/x/oauth2"
)

type Service struct {
	googleConfig oauth2.Config
	users        *user.Service
	sessions     *session.Service
	flows        *flow.Service
}

func NewService(ctx context.Context, config Config, users *user.Service, sessions *session.Service, flows *flow.Service) (*Service, error) {
	googleConfig, err := newGoogleConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	return &Service{
		googleConfig: googleConfig,
		users:        users,
		sessions:     sessions,
		flows:        flows,
	}, nil
}

func (s *Service) BeginGoogleLogin(ctx context.Context) (BeginGoogleLoginResult, error) {
	state, err := util.RandomToken()
	if err != nil {
		return BeginGoogleLoginResult{}, fmt.Errorf("generate login state: %w", err)
	}

	verifier := oauth2.GenerateVerifier()
	created, token, err := s.flows.CreateFlow(ctx, state, verifier)
	if err != nil {
		return BeginGoogleLoginResult{}, fmt.Errorf("start Google login: %w", err)
	}

	return BeginGoogleLoginResult{
		AuthorizationURL: s.googleConfig.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier)),
		FlowToken:        token,
		ExpiresAt:        created.ExpiresAt,
	}, nil
}

func (s *Service) CompleteGoogleLogin(ctx context.Context, flowToken, state, code string) (CompleteGoogleLoginResult, error) {
	var result CompleteGoogleLoginResult

	authFlow, err := s.flows.CompleteFlow(ctx, flowToken, state)
	if err != nil {
		if errors.Is(err, flow.ErrFlowNotFound) {
			return result, fmt.Errorf("%w: %w", ErrInvalidFlow, err)
		}
		return result, fmt.Errorf("complete login flow: %w", err)
	}
	result.FlowConsumed = true

	token, err := s.googleConfig.Exchange(ctx, code, oauth2.VerifierOption(authFlow.PKCEVerifier))
	if err != nil {
		return result, fmt.Errorf("%w: exchange code: %w", ErrGoogleProvider, err)
	}

	profile, err := s.fetchGoogleUser(ctx, token)
	if err != nil {
		return result, fmt.Errorf("%w: %w", ErrGoogleProvider, err)
	}

	localUser, err := s.users.FindOrCreate(ctx, user.FindOrCreateInput{
		GoogleSub:  profile.ID,
		Email:      profile.Email,
		PictureURL: profile.PictureURL,
	})
	if err != nil {
		return result, fmt.Errorf("find or create Google user: %w", err)
	}

	result.SessionToken, err = s.sessions.CreateSession(ctx, localUser.ID)
	if err != nil {
		return result, fmt.Errorf("create login session: %w", err)
	}

	return result, nil
}

func (s *Service) Logout(ctx context.Context, userID uuid.UUID) error {
	if err := s.sessions.DeleteSessionsForUser(ctx, userID); err != nil {
		return fmt.Errorf("log out user: %w", err)
	}
	return nil
}
