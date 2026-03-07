package testutil

import (
	"context"

	"dermify-api/internal/domain"
)

// MockConsentRepository is a test double for service.ConsentRepository.
type MockConsentRepository struct {
	CreateFn           func(ctx context.Context, consent *domain.Consent) error
	GetBySessionIDFn   func(ctx context.Context, sessionID int64) (*domain.Consent, error)
	UpdateFn           func(ctx context.Context, consent *domain.Consent) error
	ExistsForSessionFn func(ctx context.Context, sessionID int64) (bool, error)
}

// Create delegates to CreateFn if set, otherwise returns nil.
func (m *MockConsentRepository) Create(ctx context.Context, consent *domain.Consent) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, consent)
	}
	return nil
}

// GetBySessionID delegates to GetBySessionIDFn if set, otherwise returns nil and nil.
func (m *MockConsentRepository) GetBySessionID(ctx context.Context, sessionID int64) (*domain.Consent, error) {
	if m.GetBySessionIDFn != nil {
		return m.GetBySessionIDFn(ctx, sessionID)
	}
	return nil, nil
}

// Update delegates to UpdateFn if set, otherwise returns nil.
func (m *MockConsentRepository) Update(ctx context.Context, consent *domain.Consent) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, consent)
	}
	return nil
}

// ExistsForSession delegates to ExistsForSessionFn if set, otherwise returns false.
func (m *MockConsentRepository) ExistsForSession(ctx context.Context, sessionID int64) (bool, error) {
	if m.ExistsForSessionFn != nil {
		return m.ExistsForSessionFn(ctx, sessionID)
	}
	return false, nil
}
