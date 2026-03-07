package service

import (
	"context"
	"errors"
	"time"

	"dermify-api/internal/domain"
)

// Sentinel errors for consent operations.
var (
	ErrConsentNotFound      = errors.New("consent not found")      //nolint:gochecknoglobals // sentinel error
	ErrConsentAlreadyExists = errors.New("consent already exists") //nolint:gochecknoglobals // sentinel error
	ErrInvalidConsentData   = errors.New("invalid consent data")   //nolint:gochecknoglobals // sentinel error
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
func (s *ConsentService) RecordConsent(ctx context.Context, consent *domain.Consent) error {
	if err := validateConsent(consent); err != nil {
		return err
	}

	exists, err := s.repo.ExistsForSession(ctx, consent.SessionID)
	if err != nil {
		return err
	}

	if exists {
		return ErrConsentAlreadyExists
	}

	now := time.Now()
	consent.Version = 1
	consent.CreatedAt = now
	consent.UpdatedAt = now

	return s.repo.Create(ctx, consent)
}

// GetBySessionID retrieves the consent record for a session.
func (s *ConsentService) GetBySessionID(ctx context.Context, sessionID int64) (*domain.Consent, error) {
	return s.repo.GetBySessionID(ctx, sessionID)
}

// UpdateConsent validates and updates a consent record.
func (s *ConsentService) UpdateConsent(ctx context.Context, consent *domain.Consent) error {
	if err := validateConsent(consent); err != nil {
		return err
	}

	consent.UpdatedAt = time.Now()

	return s.repo.Update(ctx, consent)
}

// validateConsent checks required fields on a consent record.
func validateConsent(consent *domain.Consent) error {
	if consent.SessionID <= 0 {
		return ErrInvalidConsentData
	}

	if consent.ConsentType == "" {
		return ErrInvalidConsentData
	}

	if consent.ConsentMethod == "" {
		return ErrInvalidConsentData
	}

	if consent.ObtainedAt.IsZero() {
		return ErrInvalidConsentData
	}

	return nil
}
