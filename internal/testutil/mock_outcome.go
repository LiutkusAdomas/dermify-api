package testutil

import (
	"context"

	"dermify-api/internal/domain"
)

// MockOutcomeRepository is a test double for service.OutcomeRepository.
type MockOutcomeRepository struct {
	CreateFn           func(ctx context.Context, outcome *domain.SessionOutcome) error
	GetBySessionIDFn   func(ctx context.Context, sessionID int64) (*domain.SessionOutcome, error)
	UpdateFn           func(ctx context.Context, outcome *domain.SessionOutcome) error
	ExistsForSessionFn func(ctx context.Context, sessionID int64) (bool, error)
	SetEndpointsFn     func(ctx context.Context, outcomeID int64, endpointIDs []int64) error
	GetEndpointsFn     func(ctx context.Context, outcomeID int64) ([]int64, error)
}

// Create delegates to CreateFn if set, otherwise returns nil.
func (m *MockOutcomeRepository) Create(ctx context.Context, outcome *domain.SessionOutcome) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, outcome)
	}
	return nil
}

// GetBySessionID delegates to GetBySessionIDFn if set, otherwise returns nil and nil.
func (m *MockOutcomeRepository) GetBySessionID(ctx context.Context, sessionID int64) (*domain.SessionOutcome, error) {
	if m.GetBySessionIDFn != nil {
		return m.GetBySessionIDFn(ctx, sessionID)
	}
	return nil, nil
}

// Update delegates to UpdateFn if set, otherwise returns nil.
func (m *MockOutcomeRepository) Update(ctx context.Context, outcome *domain.SessionOutcome) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, outcome)
	}
	return nil
}

// ExistsForSession delegates to ExistsForSessionFn if set, otherwise returns false and nil.
func (m *MockOutcomeRepository) ExistsForSession(ctx context.Context, sessionID int64) (bool, error) {
	if m.ExistsForSessionFn != nil {
		return m.ExistsForSessionFn(ctx, sessionID)
	}
	return false, nil
}

// SetEndpoints delegates to SetEndpointsFn if set, otherwise returns nil.
func (m *MockOutcomeRepository) SetEndpoints(ctx context.Context, outcomeID int64, endpointIDs []int64) error {
	if m.SetEndpointsFn != nil {
		return m.SetEndpointsFn(ctx, outcomeID, endpointIDs)
	}
	return nil
}

// GetEndpoints delegates to GetEndpointsFn if set, otherwise returns nil and nil.
func (m *MockOutcomeRepository) GetEndpoints(ctx context.Context, outcomeID int64) ([]int64, error) {
	if m.GetEndpointsFn != nil {
		return m.GetEndpointsFn(ctx, outcomeID)
	}
	return nil, nil
}
