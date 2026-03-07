package testutil

import (
	"context"

	"dermify-api/internal/domain"
)

// MockRegistryRepository is a test double for service.RegistryRepository.
type MockRegistryRepository struct {
	ListDevicesFn           func(ctx context.Context, deviceType string) ([]domain.Device, error)
	GetDeviceByIDFn         func(ctx context.Context, id int64) (*domain.Device, error)
	ListProductsFn          func(ctx context.Context, productType string) ([]domain.Product, error)
	GetProductByIDFn        func(ctx context.Context, id int64) (*domain.Product, error)
	ListIndicationCodesFn   func(ctx context.Context, moduleType string) ([]domain.IndicationCode, error)
	ListClinicalEndpointsFn func(ctx context.Context, moduleType string) ([]domain.ClinicalEndpoint, error)
}

// ListDevices delegates to ListDevicesFn if set, otherwise returns empty slice.
func (m *MockRegistryRepository) ListDevices(ctx context.Context, deviceType string) ([]domain.Device, error) {
	if m.ListDevicesFn != nil {
		return m.ListDevicesFn(ctx, deviceType)
	}
	return []domain.Device{}, nil
}

// GetDeviceByID delegates to GetDeviceByIDFn if set, otherwise returns nil and nil.
func (m *MockRegistryRepository) GetDeviceByID(ctx context.Context, id int64) (*domain.Device, error) {
	if m.GetDeviceByIDFn != nil {
		return m.GetDeviceByIDFn(ctx, id)
	}
	return nil, nil
}

// ListProducts delegates to ListProductsFn if set, otherwise returns empty slice.
func (m *MockRegistryRepository) ListProducts(ctx context.Context, productType string) ([]domain.Product, error) {
	if m.ListProductsFn != nil {
		return m.ListProductsFn(ctx, productType)
	}
	return []domain.Product{}, nil
}

// GetProductByID delegates to GetProductByIDFn if set, otherwise returns nil and nil.
func (m *MockRegistryRepository) GetProductByID(ctx context.Context, id int64) (*domain.Product, error) {
	if m.GetProductByIDFn != nil {
		return m.GetProductByIDFn(ctx, id)
	}
	return nil, nil
}

// ListIndicationCodes delegates to ListIndicationCodesFn if set, otherwise returns empty slice.
func (m *MockRegistryRepository) ListIndicationCodes(ctx context.Context, moduleType string) ([]domain.IndicationCode, error) {
	if m.ListIndicationCodesFn != nil {
		return m.ListIndicationCodesFn(ctx, moduleType)
	}
	return []domain.IndicationCode{}, nil
}

// ListClinicalEndpoints delegates to ListClinicalEndpointsFn if set, otherwise returns empty slice.
func (m *MockRegistryRepository) ListClinicalEndpoints(ctx context.Context, moduleType string) ([]domain.ClinicalEndpoint, error) {
	if m.ListClinicalEndpointsFn != nil {
		return m.ListClinicalEndpointsFn(ctx, moduleType)
	}
	return []domain.ClinicalEndpoint{}, nil
}
