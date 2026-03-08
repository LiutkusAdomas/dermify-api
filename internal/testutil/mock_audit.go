package testutil

import (
	"context"

	"dermify-api/internal/domain"
	"dermify-api/internal/service"
)

// MockAuditRepository is a test double for service.AuditRepository.
type MockAuditRepository struct {
	ListByEntityFn func(ctx context.Context, entityType string, entityID int64) ([]domain.AuditEntry, error)
	ListFn         func(ctx context.Context, filter service.AuditFilter) (*service.AuditListResult, error)
}

// ListByEntity delegates to ListByEntityFn if set, otherwise returns nil and nil.
func (m *MockAuditRepository) ListByEntity(ctx context.Context, entityType string, entityID int64) ([]domain.AuditEntry, error) {
	if m.ListByEntityFn != nil {
		return m.ListByEntityFn(ctx, entityType, entityID)
	}
	return nil, nil
}

// List delegates to ListFn if set, otherwise returns nil and nil.
func (m *MockAuditRepository) List(ctx context.Context, filter service.AuditFilter) (*service.AuditListResult, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, filter)
	}
	return nil, nil
}
