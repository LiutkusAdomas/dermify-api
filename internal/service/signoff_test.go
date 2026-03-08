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

// signoffTestDeps holds all mocked dependencies used by SignoffService tests.
type signoffTestDeps struct {
	svc         *service.SignoffService
	sessionRepo *testutil.MockSessionRepository
	consentRepo *testutil.MockConsentRepository
	moduleRepo  *testutil.MockModuleRepository
	outcomeRepo *testutil.MockOutcomeRepository
	signoffRepo *testutil.MockSignoffRepository
}

func newSignoffTestDeps() signoffTestDeps {
	sessionRepo := &testutil.MockSessionRepository{}
	consentRepo := &testutil.MockConsentRepository{}
	moduleRepo := &testutil.MockModuleRepository{}
	outcomeRepo := &testutil.MockOutcomeRepository{}
	signoffRepo := &testutil.MockSignoffRepository{}

	svc := service.NewSignoffService(sessionRepo, consentRepo, moduleRepo, outcomeRepo, signoffRepo)

	return signoffTestDeps{
		svc:         svc,
		sessionRepo: sessionRepo,
		consentRepo: consentRepo,
		moduleRepo:  moduleRepo,
		outcomeRepo: outcomeRepo,
		signoffRepo: signoffRepo,
	}
}

// setupAwaitingSignoffSession configures the session mock to return an awaiting_signoff session.
func (d *signoffTestDeps) setupAwaitingSignoffSession(sessionID int64) {
	d.sessionRepo.GetByIDFn = func(_ context.Context, id int64) (*domain.Session, error) {
		return &domain.Session{
			ID:      id,
			Status:  domain.SessionStatusAwaitingSignoff,
			Version: 1,
		}, nil
	}
}

// setupAllComplete configures all mocks to indicate a fully complete session.
func (d *signoffTestDeps) setupAllComplete(sessionID int64) {
	d.setupAwaitingSignoffSession(sessionID)

	d.consentRepo.ExistsForSessionFn = func(_ context.Context, _ int64) (bool, error) {
		return true, nil
	}

	d.moduleRepo.ListBySessionFn = func(_ context.Context, _ int64) ([]domain.SessionModule, error) {
		return []domain.SessionModule{{ID: 1}}, nil
	}

	d.outcomeRepo.ExistsForSessionFn = func(_ context.Context, _ int64) (bool, error) {
		return true, nil
	}
}

// ---------------------------------------------------------------------------
// ValidateForSignoff tests
// ---------------------------------------------------------------------------

func TestValidateForSignoff_AllPresent(t *testing.T) {
	deps := newSignoffTestDeps()
	deps.setupAllComplete(1)

	result, err := deps.svc.ValidateForSignoff(context.Background(), 1)

	require.NoError(t, err)
	assert.True(t, result.Ready)
	assert.Empty(t, result.Missing)
}

func TestValidateForSignoff_MissingConsent(t *testing.T) {
	deps := newSignoffTestDeps()
	deps.setupAwaitingSignoffSession(1)

	deps.consentRepo.ExistsForSessionFn = func(_ context.Context, _ int64) (bool, error) {
		return false, nil
	}
	deps.moduleRepo.ListBySessionFn = func(_ context.Context, _ int64) ([]domain.SessionModule, error) {
		return []domain.SessionModule{{ID: 1}}, nil
	}
	deps.outcomeRepo.ExistsForSessionFn = func(_ context.Context, _ int64) (bool, error) {
		return true, nil
	}

	result, err := deps.svc.ValidateForSignoff(context.Background(), 1)

	require.NoError(t, err)
	assert.False(t, result.Ready)
	assert.Contains(t, result.Missing, "consent record")
}

func TestValidateForSignoff_MissingModules(t *testing.T) {
	deps := newSignoffTestDeps()
	deps.setupAwaitingSignoffSession(1)

	deps.consentRepo.ExistsForSessionFn = func(_ context.Context, _ int64) (bool, error) {
		return true, nil
	}
	deps.moduleRepo.ListBySessionFn = func(_ context.Context, _ int64) ([]domain.SessionModule, error) {
		return []domain.SessionModule{}, nil
	}
	deps.outcomeRepo.ExistsForSessionFn = func(_ context.Context, _ int64) (bool, error) {
		return true, nil
	}

	result, err := deps.svc.ValidateForSignoff(context.Background(), 1)

	require.NoError(t, err)
	assert.False(t, result.Ready)
	assert.Contains(t, result.Missing, "at least one procedure module")
}

func TestValidateForSignoff_MissingOutcome(t *testing.T) {
	deps := newSignoffTestDeps()
	deps.setupAwaitingSignoffSession(1)

	deps.consentRepo.ExistsForSessionFn = func(_ context.Context, _ int64) (bool, error) {
		return true, nil
	}
	deps.moduleRepo.ListBySessionFn = func(_ context.Context, _ int64) ([]domain.SessionModule, error) {
		return []domain.SessionModule{{ID: 1}}, nil
	}
	deps.outcomeRepo.ExistsForSessionFn = func(_ context.Context, _ int64) (bool, error) {
		return false, nil
	}

	result, err := deps.svc.ValidateForSignoff(context.Background(), 1)

	require.NoError(t, err)
	assert.False(t, result.Ready)
	assert.Contains(t, result.Missing, "outcome record")
}

func TestValidateForSignoff_MissingAll(t *testing.T) {
	deps := newSignoffTestDeps()
	deps.setupAwaitingSignoffSession(1)

	deps.consentRepo.ExistsForSessionFn = func(_ context.Context, _ int64) (bool, error) {
		return false, nil
	}
	// moduleRepo.ListBySessionFn defaults to empty slice.
	deps.outcomeRepo.ExistsForSessionFn = func(_ context.Context, _ int64) (bool, error) {
		return false, nil
	}

	result, err := deps.svc.ValidateForSignoff(context.Background(), 1)

	require.NoError(t, err)
	assert.False(t, result.Ready)
	assert.Len(t, result.Missing, 3)
	assert.Contains(t, result.Missing, "consent record")
	assert.Contains(t, result.Missing, "at least one procedure module")
	assert.Contains(t, result.Missing, "outcome record")
}

func TestValidateForSignoff_WrongState(t *testing.T) {
	deps := newSignoffTestDeps()

	deps.sessionRepo.GetByIDFn = func(_ context.Context, id int64) (*domain.Session, error) {
		return &domain.Session{ID: id, Status: domain.SessionStatusDraft}, nil
	}

	result, err := deps.svc.ValidateForSignoff(context.Background(), 1)

	require.NoError(t, err)
	assert.False(t, result.Ready)
	assert.Contains(t, result.Missing, "session must be in awaiting_signoff state")
}

func TestValidateForSignoff_SessionNotFound(t *testing.T) {
	deps := newSignoffTestDeps()

	deps.sessionRepo.GetByIDFn = func(_ context.Context, _ int64) (*domain.Session, error) {
		return nil, service.ErrSessionNotFound
	}

	result, err := deps.svc.ValidateForSignoff(context.Background(), 999)

	assert.Nil(t, result)
	assert.True(t, errors.Is(err, service.ErrSessionNotFound))
}

// ---------------------------------------------------------------------------
// SignOff tests
// ---------------------------------------------------------------------------

func TestSignOff_Success(t *testing.T) {
	deps := newSignoffTestDeps()
	deps.setupAllComplete(1)

	signoffCalled := false
	deps.signoffRepo.SignOffFn = func(_ context.Context, id int64, clinicianID int64, expectedVersion int) error {
		signoffCalled = true
		assert.Equal(t, int64(1), id)
		assert.Equal(t, int64(5), clinicianID)
		assert.Equal(t, 1, expectedVersion)
		return nil
	}

	err := deps.svc.SignOff(context.Background(), 1, 5)

	require.NoError(t, err)
	assert.True(t, signoffCalled, "signoffRepo.SignOff should be called")
}

func TestSignOff_Incomplete(t *testing.T) {
	deps := newSignoffTestDeps()
	deps.setupAwaitingSignoffSession(1)

	// Consent and modules present, but outcome missing.
	deps.consentRepo.ExistsForSessionFn = func(_ context.Context, _ int64) (bool, error) {
		return true, nil
	}
	deps.moduleRepo.ListBySessionFn = func(_ context.Context, _ int64) ([]domain.SessionModule, error) {
		return []domain.SessionModule{{ID: 1}}, nil
	}
	deps.outcomeRepo.ExistsForSessionFn = func(_ context.Context, _ int64) (bool, error) {
		return false, nil
	}

	err := deps.svc.SignOff(context.Background(), 1, 5)

	assert.True(t, errors.Is(err, service.ErrSessionIncomplete))
}

func TestSignOff_VersionConflict(t *testing.T) {
	deps := newSignoffTestDeps()
	deps.setupAllComplete(1)

	deps.signoffRepo.SignOffFn = func(_ context.Context, _ int64, _ int64, _ int) error {
		return service.ErrSessionVersionConflict
	}

	err := deps.svc.SignOff(context.Background(), 1, 5)

	assert.True(t, errors.Is(err, service.ErrSessionVersionConflict))
}

// ---------------------------------------------------------------------------
// LockSession tests
// ---------------------------------------------------------------------------

func TestLockSession_Success(t *testing.T) {
	deps := newSignoffTestDeps()

	deps.sessionRepo.GetByIDFn = func(_ context.Context, id int64) (*domain.Session, error) {
		return &domain.Session{
			ID:      id,
			Status:  domain.SessionStatusSigned,
			Version: 2,
		}, nil
	}

	lockCalled := false
	deps.signoffRepo.LockSessionFn = func(_ context.Context, id int64, expectedVersion int, userID int64) error {
		lockCalled = true
		assert.Equal(t, int64(1), id)
		assert.Equal(t, 2, expectedVersion)
		assert.Equal(t, int64(5), userID)
		return nil
	}

	err := deps.svc.LockSession(context.Background(), 1, 5)

	require.NoError(t, err)
	assert.True(t, lockCalled, "signoffRepo.LockSession should be called")
}

func TestLockSession_WrongState(t *testing.T) {
	deps := newSignoffTestDeps()

	deps.sessionRepo.GetByIDFn = func(_ context.Context, id int64) (*domain.Session, error) {
		return &domain.Session{
			ID:      id,
			Status:  domain.SessionStatusAwaitingSignoff,
			Version: 1,
		}, nil
	}

	err := deps.svc.LockSession(context.Background(), 1, 5)

	assert.True(t, errors.Is(err, service.ErrInvalidStateTransition))
}
