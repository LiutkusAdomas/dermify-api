package service_test

import (
	"context"
	"errors"
	"testing"

	"dermify-api/internal/domain"
	"dermify-api/internal/service"
	"dermify-api/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// addendumTestDeps holds all mocked dependencies used by AddendumService tests.
type addendumTestDeps struct {
	svc         *service.AddendumService
	addendumRepo *testutil.MockAddendumRepository
	sessionRepo *testutil.MockSessionRepository
}

func newAddendumTestDeps() addendumTestDeps {
	addendumRepo := &testutil.MockAddendumRepository{}
	sessionRepo := &testutil.MockSessionRepository{}

	svc := service.NewAddendumService(addendumRepo, sessionRepo)

	return addendumTestDeps{
		svc:         svc,
		addendumRepo: addendumRepo,
		sessionRepo: sessionRepo,
	}
}

// setupLockedSession configures the session mock to return a locked session.
func (d *addendumTestDeps) setupLockedSession(sessionID int64) {
	d.sessionRepo.GetByIDFn = func(_ context.Context, id int64) (*domain.Session, error) {
		return &domain.Session{
			ID:     id,
			Status: domain.SessionStatusLocked,
		}, nil
	}
}

// ---------------------------------------------------------------------------
// CreateAddendum tests
// ---------------------------------------------------------------------------

func TestCreateAddendum_Success(t *testing.T) {
	deps := newAddendumTestDeps()
	deps.setupLockedSession(1)

	createCalled := false
	deps.addendumRepo.CreateFn = func(_ context.Context, addendum *domain.Addendum) error {
		createCalled = true
		addendum.ID = 10
		return nil
	}

	addendum := &domain.Addendum{
		SessionID: 1,
		AuthorID:  5,
		Reason:    "Correction needed",
		Content:   "Updated dosage information",
	}

	err := deps.svc.CreateAddendum(context.Background(), addendum)

	require.NoError(t, err)
	assert.True(t, createCalled, "addendumRepo.Create should be called")
	assert.Equal(t, int64(10), addendum.ID)
	assert.False(t, addendum.CreatedAt.IsZero(), "created_at should be set")
}

func TestCreateAddendum_SessionNotLocked(t *testing.T) {
	deps := newAddendumTestDeps()

	deps.sessionRepo.GetByIDFn = func(_ context.Context, id int64) (*domain.Session, error) {
		return &domain.Session{
			ID:     id,
			Status: domain.SessionStatusSigned,
		}, nil
	}

	addendum := &domain.Addendum{
		SessionID: 1,
		AuthorID:  5,
		Reason:    "Correction needed",
		Content:   "Updated dosage information",
	}

	err := deps.svc.CreateAddendum(context.Background(), addendum)

	assert.True(t, errors.Is(err, service.ErrSessionNotLocked))
}

func TestCreateAddendum_EmptyReason(t *testing.T) {
	deps := newAddendumTestDeps()

	addendum := &domain.Addendum{
		SessionID: 1,
		AuthorID:  5,
		Reason:    "",
		Content:   "Updated dosage information",
	}

	err := deps.svc.CreateAddendum(context.Background(), addendum)

	assert.True(t, errors.Is(err, service.ErrInvalidAddendumData))
}

func TestCreateAddendum_EmptyContent(t *testing.T) {
	deps := newAddendumTestDeps()

	addendum := &domain.Addendum{
		SessionID: 1,
		AuthorID:  5,
		Reason:    "Correction needed",
		Content:   "",
	}

	err := deps.svc.CreateAddendum(context.Background(), addendum)

	assert.True(t, errors.Is(err, service.ErrInvalidAddendumData))
}

// ---------------------------------------------------------------------------
// ListBySession tests
// ---------------------------------------------------------------------------

func TestListBySession_Delegates(t *testing.T) {
	deps := newAddendumTestDeps()

	expected := []domain.Addendum{
		{ID: 1, SessionID: 10, Reason: "first"},
		{ID: 2, SessionID: 10, Reason: "second"},
	}

	deps.addendumRepo.ListBySessionFn = func(_ context.Context, sessionID int64) ([]domain.Addendum, error) {
		assert.Equal(t, int64(10), sessionID)
		return expected, nil
	}

	result, err := deps.svc.ListBySession(context.Background(), 10)

	require.NoError(t, err)
	assert.Equal(t, expected, result)
}
