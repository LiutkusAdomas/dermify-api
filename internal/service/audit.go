package service

import (
	"context"

	"dermify-api/internal/domain"
)

// auditMaxPerPage is the maximum per-page value for audit trail queries.
const auditMaxPerPage = 100

// auditDefaultPerPage is the default per-page value for audit trail queries.
const auditDefaultPerPage = 50

// AuditFilter defines filtering and pagination options for audit trail queries.
type AuditFilter struct {
	EntityType string
	EntityID   int64
	UserID     int64
	Page       int
	PerPage    int
}

// AuditListResult holds paginated audit trail results.
type AuditListResult struct {
	Entries []domain.AuditEntry
	Total   int
}

// AuditRepository defines the read-only data access contract for audit trail entries.
// Audit entries are created by database triggers, so no Create method is needed.
type AuditRepository interface {
	// ListByEntity returns all audit entries for a specific entity.
	ListByEntity(ctx context.Context, entityType string, entityID int64) ([]domain.AuditEntry, error)
	// List returns paginated audit entries matching the given filter.
	List(ctx context.Context, filter AuditFilter) (*AuditListResult, error)
}

// AuditService provides read-only access to the audit trail.
type AuditService struct {
	repo AuditRepository
}

// NewAuditService creates a new AuditService with the given repository.
func NewAuditService(repo AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

// ListByEntity returns all audit entries for a specific entity.
func (s *AuditService) ListByEntity(ctx context.Context, entityType string, entityID int64) ([]domain.AuditEntry, error) {
	return s.repo.ListByEntity(ctx, entityType, entityID)
}

// List returns paginated audit entries with defaults applied.
func (s *AuditService) List(ctx context.Context, filter AuditFilter) (*AuditListResult, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}

	if filter.PerPage <= 0 {
		filter.PerPage = auditDefaultPerPage
	}

	if filter.PerPage > auditMaxPerPage {
		filter.PerPage = auditMaxPerPage
	}

	return s.repo.List(ctx, filter)
}
