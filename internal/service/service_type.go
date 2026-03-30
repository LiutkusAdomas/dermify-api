package service

import (
	"context"
	"errors"
	"time"

	"dermify-api/internal/domain"
)

var (
	ErrServiceTypeNotFound    = errors.New("service type not found")       //nolint:gochecknoglobals // sentinel error
	ErrServiceTypeNameExists  = errors.New("service type name already exists") //nolint:gochecknoglobals // sentinel error
	ErrInvalidServiceTypeData = errors.New("invalid service type data")    //nolint:gochecknoglobals // sentinel error
)

// ServiceTypeRepository defines the data access contract for service types.
type ServiceTypeRepository interface {
	Create(ctx context.Context, st *domain.ServiceType) error
	GetByID(ctx context.Context, id, orgID int64) (*domain.ServiceType, error)
	Update(ctx context.Context, st *domain.ServiceType) error
	Delete(ctx context.Context, id, orgID int64) error
	ListByOrg(ctx context.Context, orgID int64, activeOnly bool) ([]*domain.ServiceType, error)
}

// ServiceTypeService handles service type business logic.
type ServiceTypeService struct {
	repo ServiceTypeRepository
}

// NewServiceTypeService creates a new ServiceTypeService.
func NewServiceTypeService(repo ServiceTypeRepository) *ServiceTypeService {
	return &ServiceTypeService{repo: repo}
}

// Create validates and inserts a new service type.
func (s *ServiceTypeService) Create(ctx context.Context, st *domain.ServiceType) error {
	if st.Name == "" || st.OrgID <= 0 {
		return ErrInvalidServiceTypeData
	}
	if st.DefaultDuration <= 0 {
		st.DefaultDuration = 30
	}
	st.Active = true
	now := time.Now()
	st.CreatedAt = now
	st.UpdatedAt = now
	return s.repo.Create(ctx, st)
}

// GetByID retrieves a single service type.
func (s *ServiceTypeService) GetByID(ctx context.Context, id, orgID int64) (*domain.ServiceType, error) {
	return s.repo.GetByID(ctx, id, orgID)
}

// Update modifies a service type.
func (s *ServiceTypeService) Update(ctx context.Context, st *domain.ServiceType) error {
	if st.Name == "" {
		return ErrInvalidServiceTypeData
	}
	if st.DefaultDuration <= 0 {
		st.DefaultDuration = 30
	}
	st.UpdatedAt = time.Now()
	return s.repo.Update(ctx, st)
}

// Delete removes a service type.
func (s *ServiceTypeService) Delete(ctx context.Context, id, orgID int64) error {
	return s.repo.Delete(ctx, id, orgID)
}

// ListByOrg returns all service types for an organization.
func (s *ServiceTypeService) ListByOrg(ctx context.Context, orgID int64, activeOnly bool) ([]*domain.ServiceType, error) {
	return s.repo.ListByOrg(ctx, orgID, activeOnly)
}
