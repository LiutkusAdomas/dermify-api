package testutil

import (
	"context"

	"dermify-api/internal/domain"
)

// MockAddendumRepository is a test double for service.AddendumRepository.
type MockAddendumRepository struct {
	CreateFn        func(ctx context.Context, addendum *domain.Addendum) error
	GetByIDFn       func(ctx context.Context, id int64) (*domain.Addendum, error)
	ListBySessionFn func(ctx context.Context, sessionID int64) ([]domain.Addendum, error)
}

// Create delegates to CreateFn if set, otherwise returns nil.
func (m *MockAddendumRepository) Create(ctx context.Context, addendum *domain.Addendum) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, addendum)
	}
	return nil
}

// GetByID delegates to GetByIDFn if set, otherwise returns nil and nil.
func (m *MockAddendumRepository) GetByID(ctx context.Context, id int64) (*domain.Addendum, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, nil
}

// ListBySession delegates to ListBySessionFn if set, otherwise returns nil and nil.
func (m *MockAddendumRepository) ListBySession(ctx context.Context, sessionID int64) ([]domain.Addendum, error) {
	if m.ListBySessionFn != nil {
		return m.ListBySessionFn(ctx, sessionID)
	}
	return nil, nil
}
