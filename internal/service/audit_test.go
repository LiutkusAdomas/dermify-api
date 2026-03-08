package service_test

import (
	"context"
	"testing"

	"dermify-api/internal/domain"
	"dermify-api/internal/service"
	"dermify-api/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// auditTestDeps holds all mocked dependencies used by AuditService tests.
type auditTestDeps struct {
	svc      *service.AuditService
	auditRepo *testutil.MockAuditRepository
}

func newAuditTestDeps() auditTestDeps {
	auditRepo := &testutil.MockAuditRepository{}

	svc := service.NewAuditService(auditRepo)

	return auditTestDeps{
		svc:      svc,
		auditRepo: auditRepo,
	}
}

// ---------------------------------------------------------------------------
// ListByEntity tests
// ---------------------------------------------------------------------------

func TestAuditListByEntity_Delegates(t *testing.T) {
	deps := newAuditTestDeps()

	expected := []domain.AuditEntry{
		{ID: 1, EntityType: "session", EntityID: 10, Action: "UPDATE"},
		{ID: 2, EntityType: "session", EntityID: 10, Action: "INSERT"},
	}

	deps.auditRepo.ListByEntityFn = func(_ context.Context, entityType string, entityID int64) ([]domain.AuditEntry, error) {
		assert.Equal(t, "session", entityType)
		assert.Equal(t, int64(10), entityID)
		return expected, nil
	}

	result, err := deps.svc.ListByEntity(context.Background(), "session", 10)

	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

// ---------------------------------------------------------------------------
// List (pagination) tests
// ---------------------------------------------------------------------------

func TestAuditList_DefaultPagination(t *testing.T) {
	deps := newAuditTestDeps()

	deps.auditRepo.ListFn = func(_ context.Context, filter service.AuditFilter) (*service.AuditListResult, error) {
		// Verify defaults were applied.
		assert.Equal(t, 1, filter.Page, "page=0 should become page=1")
		assert.Equal(t, 50, filter.PerPage, "perPage=0 should become 50")
		return &service.AuditListResult{Entries: []domain.AuditEntry{}, Total: 0}, nil
	}

	filter := service.AuditFilter{
		Page:    0,
		PerPage: 0,
	}

	result, err := deps.svc.List(context.Background(), filter)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.Total)
}

func TestAuditList_MaxPerPage(t *testing.T) {
	deps := newAuditTestDeps()

	deps.auditRepo.ListFn = func(_ context.Context, filter service.AuditFilter) (*service.AuditListResult, error) {
		// Verify cap was applied.
		assert.Equal(t, 100, filter.PerPage, "perPage=200 should be capped to 100")
		return &service.AuditListResult{Entries: []domain.AuditEntry{}, Total: 0}, nil
	}

	filter := service.AuditFilter{
		Page:    1,
		PerPage: 200,
	}

	result, err := deps.svc.List(context.Background(), filter)

	require.NoError(t, err)
	assert.NotNil(t, result)
}
