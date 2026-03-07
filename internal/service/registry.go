package service

import (
	"context"
	"errors"

	"dermify-api/internal/domain"
)

// ErrDeviceNotFound is returned when no device matches the given ID.
var ErrDeviceNotFound = errors.New("device not found") //nolint:gochecknoglobals // sentinel error

// ErrProductNotFound is returned when no product matches the given ID.
var ErrProductNotFound = errors.New("product not found") //nolint:gochecknoglobals // sentinel error

// RegistryRepository defines the data access contract for registry operations.
type RegistryRepository interface {
	// ListDevices returns all active devices, optionally filtered by device type.
	ListDevices(ctx context.Context, deviceType string) ([]domain.Device, error)
	// GetDeviceByID returns a single device with its handpieces.
	GetDeviceByID(ctx context.Context, id int64) (*domain.Device, error)
	// ListProducts returns all active products, optionally filtered by product type.
	ListProducts(ctx context.Context, productType string) ([]domain.Product, error)
	// GetProductByID returns a single product by ID.
	GetProductByID(ctx context.Context, id int64) (*domain.Product, error)
	// ListIndicationCodes returns indication codes, optionally filtered by module type.
	ListIndicationCodes(ctx context.Context, moduleType string) ([]domain.IndicationCode, error)
	// ListClinicalEndpoints returns clinical endpoints, optionally filtered by module type.
	ListClinicalEndpoints(ctx context.Context, moduleType string) ([]domain.ClinicalEndpoint, error)
}

// RegistryService handles registry-related business logic.
type RegistryService struct {
	repo RegistryRepository
}

// NewRegistryService creates a new RegistryService with the given repository.
func NewRegistryService(repo RegistryRepository) *RegistryService {
	return &RegistryService{repo: repo}
}

// ListDevices returns all active devices, optionally filtered by device type.
func (s *RegistryService) ListDevices(ctx context.Context, deviceType string) ([]domain.Device, error) {
	return s.repo.ListDevices(ctx, deviceType)
}

// GetDeviceByID returns a single device with its handpieces.
func (s *RegistryService) GetDeviceByID(ctx context.Context, id int64) (*domain.Device, error) {
	return s.repo.GetDeviceByID(ctx, id)
}

// ListProducts returns all active products, optionally filtered by product type.
func (s *RegistryService) ListProducts(ctx context.Context, productType string) ([]domain.Product, error) {
	return s.repo.ListProducts(ctx, productType)
}

// GetProductByID returns a single product by ID.
func (s *RegistryService) GetProductByID(ctx context.Context, id int64) (*domain.Product, error) {
	return s.repo.GetProductByID(ctx, id)
}

// ListIndicationCodes returns indication codes, optionally filtered by module type.
func (s *RegistryService) ListIndicationCodes(ctx context.Context, moduleType string) ([]domain.IndicationCode, error) {
	return s.repo.ListIndicationCodes(ctx, moduleType)
}

// ListClinicalEndpoints returns clinical endpoints, optionally filtered by module type.
func (s *RegistryService) ListClinicalEndpoints(ctx context.Context, moduleType string) ([]domain.ClinicalEndpoint, error) {
	return s.repo.ListClinicalEndpoints(ctx, moduleType)
}
