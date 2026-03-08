package service

import (
	"context"
	"errors"
	"time"

	"dermify-api/internal/domain"
)

// Sentinel errors for outcome operations.
var (
	ErrOutcomeNotFound      = errors.New("outcome not found")                      //nolint:gochecknoglobals // sentinel error
	ErrOutcomeAlreadyExists = errors.New("outcome already exists")                 //nolint:gochecknoglobals // sentinel error
	ErrInvalidOutcomeData   = errors.New("invalid outcome data")                   //nolint:gochecknoglobals // sentinel error
	ErrSessionNotReady      = errors.New("session not ready for outcome recording") //nolint:gochecknoglobals // sentinel error
)

// OutcomeRepository defines the data access contract for session outcome records.
type OutcomeRepository interface {
	// Create inserts a new session outcome record.
	Create(ctx context.Context, outcome *domain.SessionOutcome) error
	// GetBySessionID retrieves the outcome record for a session.
	GetBySessionID(ctx context.Context, sessionID int64) (*domain.SessionOutcome, error)
	// Update modifies a session outcome record using optimistic locking.
	Update(ctx context.Context, outcome *domain.SessionOutcome) error
	// ExistsForSession checks whether an outcome record exists for a session.
	ExistsForSession(ctx context.Context, sessionID int64) (bool, error)
	// SetEndpoints replaces the endpoint associations for an outcome.
	SetEndpoints(ctx context.Context, outcomeID int64, endpointIDs []int64) error
	// GetEndpoints returns the endpoint IDs associated with an outcome.
	GetEndpoints(ctx context.Context, outcomeID int64) ([]int64, error)
}

// OutcomeService handles outcome business logic following the singleton
// pattern (at most one outcome per session, like Consent).
type OutcomeService struct {
	repo       OutcomeRepository
	sessionSvc *SessionService
}

// NewOutcomeService creates a new OutcomeService with the given dependencies.
func NewOutcomeService(repo OutcomeRepository, sessionSvc *SessionService) *OutcomeService {
	return &OutcomeService{
		repo:       repo,
		sessionSvc: sessionSvc,
	}
}

// validateOutcome checks required fields and business rules on an outcome record.
func validateOutcome(outcome *domain.SessionOutcome) error {
	if outcome.SessionID <= 0 {
		return ErrInvalidOutcomeData
	}

	if outcome.OutcomeStatus != domain.OutcomeStatusCompleted &&
		outcome.OutcomeStatus != domain.OutcomeStatusPartial &&
		outcome.OutcomeStatus != domain.OutcomeStatusAborted {
		return ErrInvalidOutcomeData
	}

	// OUT-04: If aftercare notes are provided, red flags text is mandatory.
	if outcome.AftercareNotes != nil && *outcome.AftercareNotes != "" {
		if outcome.RedFlagsText == nil || *outcome.RedFlagsText == "" {
			return ErrInvalidOutcomeData
		}
	}

	return nil
}

// RecordOutcome validates and creates an outcome record for a session.
// The session must be in_progress or awaiting_signoff. Only one outcome per
// session is allowed.
func (s *OutcomeService) RecordOutcome(ctx context.Context, outcome *domain.SessionOutcome) error {
	if err := validateOutcome(outcome); err != nil {
		return err
	}

	session, err := s.sessionSvc.GetByID(ctx, outcome.SessionID)
	if err != nil {
		return err
	}

	if session.Status != domain.SessionStatusInProgress &&
		session.Status != domain.SessionStatusAwaitingSignoff {
		return ErrSessionNotReady
	}

	exists, err := s.repo.ExistsForSession(ctx, outcome.SessionID)
	if err != nil {
		return err
	}

	if exists {
		return ErrOutcomeAlreadyExists
	}

	now := time.Now()
	outcome.Version = 1
	outcome.CreatedAt = now
	outcome.UpdatedAt = now

	if err := s.repo.Create(ctx, outcome); err != nil {
		return err
	}

	if len(outcome.EndpointIDs) > 0 {
		if err := s.repo.SetEndpoints(ctx, outcome.ID, outcome.EndpointIDs); err != nil {
			return err
		}
	}

	return nil
}

// GetBySessionID retrieves the outcome record for a session, including its
// associated endpoint IDs.
func (s *OutcomeService) GetBySessionID(ctx context.Context, sessionID int64) (*domain.SessionOutcome, error) {
	outcome, err := s.repo.GetBySessionID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	endpointIDs, err := s.repo.GetEndpoints(ctx, outcome.ID)
	if err != nil {
		return nil, err
	}

	outcome.EndpointIDs = endpointIDs

	return outcome, nil
}

// UpdateOutcome validates and updates an outcome record. Endpoint associations
// are replaced (delete-all + re-insert pattern).
func (s *OutcomeService) UpdateOutcome(ctx context.Context, outcome *domain.SessionOutcome) error {
	if err := validateOutcome(outcome); err != nil {
		return err
	}

	outcome.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, outcome); err != nil {
		return err
	}

	return s.repo.SetEndpoints(ctx, outcome.ID, outcome.EndpointIDs)
}
