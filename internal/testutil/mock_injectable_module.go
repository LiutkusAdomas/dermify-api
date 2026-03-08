package testutil

import (
	"context"

	"dermify-api/internal/domain"
)

// MockFillerModuleRepository is a test double for service.FillerModuleRepository.
type MockFillerModuleRepository struct {
	CreateFn        func(ctx context.Context, detail *domain.FillerModuleDetail) error
	GetByModuleIDFn func(ctx context.Context, moduleID int64) (*domain.FillerModuleDetail, error)
	UpdateFn        func(ctx context.Context, detail *domain.FillerModuleDetail) error
}

// Create delegates to CreateFn if set, otherwise returns nil.
func (m *MockFillerModuleRepository) Create(ctx context.Context, detail *domain.FillerModuleDetail) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, detail)
	}
	return nil
}

// GetByModuleID delegates to GetByModuleIDFn if set, otherwise returns nil and nil.
func (m *MockFillerModuleRepository) GetByModuleID(ctx context.Context, moduleID int64) (*domain.FillerModuleDetail, error) {
	if m.GetByModuleIDFn != nil {
		return m.GetByModuleIDFn(ctx, moduleID)
	}
	return nil, nil
}

// Update delegates to UpdateFn if set, otherwise returns nil.
func (m *MockFillerModuleRepository) Update(ctx context.Context, detail *domain.FillerModuleDetail) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, detail)
	}
	return nil
}

// MockBotulinumModuleRepository is a test double for service.BotulinumModuleRepository.
type MockBotulinumModuleRepository struct {
	CreateFn        func(ctx context.Context, detail *domain.BotulinumModuleDetail) error
	GetByModuleIDFn func(ctx context.Context, moduleID int64) (*domain.BotulinumModuleDetail, error)
	UpdateFn        func(ctx context.Context, detail *domain.BotulinumModuleDetail) error
}

// Create delegates to CreateFn if set, otherwise returns nil.
func (m *MockBotulinumModuleRepository) Create(ctx context.Context, detail *domain.BotulinumModuleDetail) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, detail)
	}
	return nil
}

// GetByModuleID delegates to GetByModuleIDFn if set, otherwise returns nil and nil.
func (m *MockBotulinumModuleRepository) GetByModuleID(ctx context.Context, moduleID int64) (*domain.BotulinumModuleDetail, error) {
	if m.GetByModuleIDFn != nil {
		return m.GetByModuleIDFn(ctx, moduleID)
	}
	return nil, nil
}

// Update delegates to UpdateFn if set, otherwise returns nil.
func (m *MockBotulinumModuleRepository) Update(ctx context.Context, detail *domain.BotulinumModuleDetail) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, detail)
	}
	return nil
}
