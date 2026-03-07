package testutil

import (
	"context"

	"dermify-api/internal/domain"
)

// MockContraindicationRepository is a test double for service.ContraindicationRepository.
type MockContraindicationRepository struct {
	CreateFn         func(ctx context.Context, screening *domain.ContraindicationScreening) error
	GetBySessionIDFn func(ctx context.Context, sessionID int64) (*domain.ContraindicationScreening, error)
	UpdateFn         func(ctx context.Context, screening *domain.ContraindicationScreening) error
}

// Create delegates to CreateFn if set, otherwise returns nil.
func (m *MockContraindicationRepository) Create(ctx context.Context, screening *domain.ContraindicationScreening) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, screening)
	}
	return nil
}

// GetBySessionID delegates to GetBySessionIDFn if set, otherwise returns nil and nil.
func (m *MockContraindicationRepository) GetBySessionID(ctx context.Context, sessionID int64) (*domain.ContraindicationScreening, error) {
	if m.GetBySessionIDFn != nil {
		return m.GetBySessionIDFn(ctx, sessionID)
	}
	return nil, nil
}

// Update delegates to UpdateFn if set, otherwise returns nil.
func (m *MockContraindicationRepository) Update(ctx context.Context, screening *domain.ContraindicationScreening) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, screening)
	}
	return nil
}
