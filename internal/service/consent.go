package service

import (
	"context"
	"errors"

	"dermify-api/internal/domain"
)

// Sentinel errors for consent operations.
var (
	ErrConsentNotFound      = errors.New("consent not found")       //nolint:gochecknoglobals // sentinel error
	ErrConsentAlreadyExists = errors.New("consent already exists")  //nolint:gochecknoglobals // sentinel error
	ErrInvalidConsentData   = errors.New("invalid consent data")    //nolint:gochecknoglobals // sentinel error
)

// ConsentRepository defines the data access contract for consent records.
type ConsentRepository interface {
	// Create inserts a new consent record and sets the ID on the provided struct.
	Create(ctx context.Context, consent *domain.Consent) error
	// GetBySessionID retrieves the consent record for a session.
	GetBySessionID(ctx context.Context, sessionID int64) (*domain.Consent, error)
	// Update modifies a consent record using optimistic locking.
	Update(ctx context.Context, consent *domain.Consent) error
	// ExistsForSession checks whether a consent record exists for a session.
	ExistsForSession(ctx context.Context, sessionID int64) (bool, error)
}

// ConsentService handles consent business logic.
type ConsentService struct {
	repo ConsentRepository
}

// NewConsentService creates a new ConsentService with the given repository.
func NewConsentService(repo ConsentRepository) *ConsentService {
	return &ConsentService{repo: repo}
}

// RecordConsent validates and creates a consent record for a session.
// TODO: implement in Plan 02.
func (s *ConsentService) RecordConsent(_ context.Context, _ *domain.Consent) error {
	return nil
}

// GetBySessionID retrieves the consent record for a session.
// TODO: implement in Plan 02.
func (s *ConsentService) GetBySessionID(_ context.Context, _ int64) (*domain.Consent, error) {
	return nil, nil
}

// UpdateConsent validates and updates a consent record.
// TODO: implement in Plan 02.
func (s *ConsentService) UpdateConsent(_ context.Context, _ *domain.Consent) error {
	return nil
}
