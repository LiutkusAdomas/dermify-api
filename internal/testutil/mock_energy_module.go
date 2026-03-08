package testutil

import (
	"context"

	"dermify-api/internal/domain"
)

// MockIPLModuleRepository is a test double for service.IPLModuleRepository.
type MockIPLModuleRepository struct {
	CreateFn        func(ctx context.Context, detail *domain.IPLModuleDetail) error
	GetByModuleIDFn func(ctx context.Context, moduleID int64) (*domain.IPLModuleDetail, error)
	UpdateFn        func(ctx context.Context, detail *domain.IPLModuleDetail) error
}

// Create delegates to CreateFn if set, otherwise returns nil.
func (m *MockIPLModuleRepository) Create(ctx context.Context, detail *domain.IPLModuleDetail) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, detail)
	}
	return nil
}

// GetByModuleID delegates to GetByModuleIDFn if set, otherwise returns nil and nil.
func (m *MockIPLModuleRepository) GetByModuleID(ctx context.Context, moduleID int64) (*domain.IPLModuleDetail, error) {
	if m.GetByModuleIDFn != nil {
		return m.GetByModuleIDFn(ctx, moduleID)
	}
	return nil, nil
}

// Update delegates to UpdateFn if set, otherwise returns nil.
func (m *MockIPLModuleRepository) Update(ctx context.Context, detail *domain.IPLModuleDetail) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, detail)
	}
	return nil
}

// MockNdYAGModuleRepository is a test double for service.NdYAGModuleRepository.
type MockNdYAGModuleRepository struct {
	CreateFn        func(ctx context.Context, detail *domain.NdYAGModuleDetail) error
	GetByModuleIDFn func(ctx context.Context, moduleID int64) (*domain.NdYAGModuleDetail, error)
	UpdateFn        func(ctx context.Context, detail *domain.NdYAGModuleDetail) error
}

// Create delegates to CreateFn if set, otherwise returns nil.
func (m *MockNdYAGModuleRepository) Create(ctx context.Context, detail *domain.NdYAGModuleDetail) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, detail)
	}
	return nil
}

// GetByModuleID delegates to GetByModuleIDFn if set, otherwise returns nil and nil.
func (m *MockNdYAGModuleRepository) GetByModuleID(ctx context.Context, moduleID int64) (*domain.NdYAGModuleDetail, error) {
	if m.GetByModuleIDFn != nil {
		return m.GetByModuleIDFn(ctx, moduleID)
	}
	return nil, nil
}

// Update delegates to UpdateFn if set, otherwise returns nil.
func (m *MockNdYAGModuleRepository) Update(ctx context.Context, detail *domain.NdYAGModuleDetail) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, detail)
	}
	return nil
}

// MockCO2ModuleRepository is a test double for service.CO2ModuleRepository.
type MockCO2ModuleRepository struct {
	CreateFn        func(ctx context.Context, detail *domain.CO2ModuleDetail) error
	GetByModuleIDFn func(ctx context.Context, moduleID int64) (*domain.CO2ModuleDetail, error)
	UpdateFn        func(ctx context.Context, detail *domain.CO2ModuleDetail) error
}

// Create delegates to CreateFn if set, otherwise returns nil.
func (m *MockCO2ModuleRepository) Create(ctx context.Context, detail *domain.CO2ModuleDetail) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, detail)
	}
	return nil
}

// GetByModuleID delegates to GetByModuleIDFn if set, otherwise returns nil and nil.
func (m *MockCO2ModuleRepository) GetByModuleID(ctx context.Context, moduleID int64) (*domain.CO2ModuleDetail, error) {
	if m.GetByModuleIDFn != nil {
		return m.GetByModuleIDFn(ctx, moduleID)
	}
	return nil, nil
}

// Update delegates to UpdateFn if set, otherwise returns nil.
func (m *MockCO2ModuleRepository) Update(ctx context.Context, detail *domain.CO2ModuleDetail) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, detail)
	}
	return nil
}

// MockRFModuleRepository is a test double for service.RFModuleRepository.
type MockRFModuleRepository struct {
	CreateFn        func(ctx context.Context, detail *domain.RFModuleDetail) error
	GetByModuleIDFn func(ctx context.Context, moduleID int64) (*domain.RFModuleDetail, error)
	UpdateFn        func(ctx context.Context, detail *domain.RFModuleDetail) error
}

// Create delegates to CreateFn if set, otherwise returns nil.
func (m *MockRFModuleRepository) Create(ctx context.Context, detail *domain.RFModuleDetail) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, detail)
	}
	return nil
}

// GetByModuleID delegates to GetByModuleIDFn if set, otherwise returns nil and nil.
func (m *MockRFModuleRepository) GetByModuleID(ctx context.Context, moduleID int64) (*domain.RFModuleDetail, error) {
	if m.GetByModuleIDFn != nil {
		return m.GetByModuleIDFn(ctx, moduleID)
	}
	return nil, nil
}

// Update delegates to UpdateFn if set, otherwise returns nil.
func (m *MockRFModuleRepository) Update(ctx context.Context, detail *domain.RFModuleDetail) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, detail)
	}
	return nil
}
