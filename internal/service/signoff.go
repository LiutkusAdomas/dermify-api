package service

import (
	"context"
	"errors"

	"dermify-api/internal/domain"
)

// Sentinel errors for sign-off operations.
var (
	ErrSessionIncomplete       = errors.New("session is incomplete for sign-off")       //nolint:gochecknoglobals // sentinel error
	ErrSessionNotAwaitingSignoff = errors.New("session not in awaiting_signoff state")   //nolint:gochecknoglobals // sentinel error
)

// SignoffRepository defines the data access contract for sign-off operations.
type SignoffRepository interface {
	// SignOff atomically sets status='signed', signed_at=now, signed_by=clinicianID,
	// increments version, with optimistic locking on expectedVersion.
	SignOff(ctx context.Context, id int64, clinicianID int64, expectedVersion int) error
	// LockSession transitions a session from signed to locked status.
	LockSession(ctx context.Context, id int64, expectedVersion int, userID int64) error
}

// ValidationResult holds the sign-off completeness check results.
type ValidationResult struct {
	Ready   bool     `json:"ready"`
	Missing []string `json:"missing,omitempty"`
}

// SignoffService validates session completeness and orchestrates sign-off.
type SignoffService struct {
	sessionRepo SessionRepository
	consentRepo ConsentRepository
	moduleRepo  ModuleRepository
	outcomeRepo OutcomeRepository
	signoffRepo SignoffRepository
}

// NewSignoffService creates a new SignoffService with the given repositories.
func NewSignoffService(
	sessionRepo SessionRepository,
	consentRepo ConsentRepository,
	moduleRepo ModuleRepository,
	outcomeRepo OutcomeRepository,
	signoffRepo SignoffRepository,
) *SignoffService {
	return &SignoffService{
		sessionRepo: sessionRepo,
		consentRepo: consentRepo,
		moduleRepo:  moduleRepo,
		outcomeRepo: outcomeRepo,
		signoffRepo: signoffRepo,
	}
}

// ValidateForSignoff checks that all required session components exist.
func (s *SignoffService) ValidateForSignoff(ctx context.Context, sessionID int64) (*ValidationResult, error) {
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if session.Status != domain.SessionStatusAwaitingSignoff {
		return &ValidationResult{
			Ready:   false,
			Missing: []string{"session must be in awaiting_signoff state"},
		}, nil
	}

	var missing []string

	// Check consent exists.
	hasConsent, err := s.consentRepo.ExistsForSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if !hasConsent {
		missing = append(missing, "consent record")
	}

	// Check at least one module exists.
	modules, err := s.moduleRepo.ListBySession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if len(modules) == 0 {
		missing = append(missing, "at least one procedure module")
	}

	// Check outcome exists.
	hasOutcome, err := s.outcomeRepo.ExistsForSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if !hasOutcome {
		missing = append(missing, "outcome record")
	}

	return &ValidationResult{Ready: len(missing) == 0, Missing: missing}, nil
}

// SignOff validates completeness and transitions the session to signed status.
// It sets signed_at and signed_by atomically via the SignoffRepository.
func (s *SignoffService) SignOff(ctx context.Context, sessionID int64, clinicianID int64) error {
	result, err := s.ValidateForSignoff(ctx, sessionID)
	if err != nil {
		return err
	}

	if !result.Ready {
		return ErrSessionIncomplete
	}

	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return err
	}

	return s.signoffRepo.SignOff(ctx, sessionID, clinicianID, session.Version)
}

// LockSession transitions a signed session to the locked state.
func (s *SignoffService) LockSession(ctx context.Context, sessionID int64, userID int64) error {
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return err
	}

	if session.Status != domain.SessionStatusSigned {
		return ErrInvalidStateTransition
	}

	return s.signoffRepo.LockSession(ctx, sessionID, session.Version, userID)
}
