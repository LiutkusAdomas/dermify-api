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

// outcomeTestDeps holds all mocked dependencies used by OutcomeService tests.
type outcomeTestDeps struct {
	svc         *service.OutcomeService
	outcomeRepo *testutil.MockOutcomeRepository
	sessionRepo *testutil.MockSessionRepository
	consentRepo *testutil.MockConsentRepository
	moduleRepo  *testutil.MockModuleRepository
}

func newOutcomeTestDeps() outcomeTestDeps {
	outcomeRepo := &testutil.MockOutcomeRepository{}
	sessionRepo := &testutil.MockSessionRepository{}
	consentRepo := &testutil.MockConsentRepository{}
	moduleRepo := &testutil.MockModuleRepository{}

	sessionSvc := service.NewSessionService(sessionRepo, consentRepo, moduleRepo)
	svc := service.NewOutcomeService(outcomeRepo, sessionSvc)

	return outcomeTestDeps{
		svc:         svc,
		outcomeRepo: outcomeRepo,
		sessionRepo: sessionRepo,
		consentRepo: consentRepo,
		moduleRepo:  moduleRepo,
	}
}

// setupInProgressSession configures the session mock to return an in_progress session.
func (d *outcomeTestDeps) setupInProgressSession(sessionID int64) {
	d.sessionRepo.GetByIDFn = func(_ context.Context, id int64) (*domain.Session, error) {
		return &domain.Session{ID: id, Status: domain.SessionStatusInProgress}, nil
	}
}

// setupAwaitingSignoffSession configures the session mock to return an awaiting_signoff session.
func (d *outcomeTestDeps) setupAwaitingSignoffSession(sessionID int64) {
	d.sessionRepo.GetByIDFn = func(_ context.Context, id int64) (*domain.Session, error) {
		return &domain.Session{ID: id, Status: domain.SessionStatusAwaitingSignoff}, nil
	}
}

// ---------------------------------------------------------------------------
// RecordOutcome tests
// ---------------------------------------------------------------------------

func TestRecordOutcome_Success(t *testing.T) {
	deps := newOutcomeTestDeps()
	deps.setupInProgressSession(1)

	deps.outcomeRepo.ExistsForSessionFn = func(_ context.Context, _ int64) (bool, error) {
		return false, nil
	}

	createCalled := false
	deps.outcomeRepo.CreateFn = func(_ context.Context, outcome *domain.SessionOutcome) error {
		createCalled = true
		outcome.ID = 10
		return nil
	}

	endpointsCalled := false
	deps.outcomeRepo.SetEndpointsFn = func(_ context.Context, outcomeID int64, endpointIDs []int64) error {
		endpointsCalled = true
		assert.Equal(t, int64(10), outcomeID)
		assert.Equal(t, []int64{1, 2, 3}, endpointIDs)
		return nil
	}

	aftercare := "Rest for 24h"
	redFlags := "Contact us if swelling persists"
	outcome := &domain.SessionOutcome{
		SessionID:      1,
		OutcomeStatus:  domain.OutcomeStatusCompleted,
		EndpointIDs:    []int64{1, 2, 3},
		AftercareNotes: &aftercare,
		RedFlagsText:   &redFlags,
		CreatedBy:      5,
		UpdatedBy:      5,
	}

	err := deps.svc.RecordOutcome(context.Background(), outcome)

	require.NoError(t, err)
	assert.True(t, createCalled, "outcome repo Create should be called")
	assert.True(t, endpointsCalled, "outcome repo SetEndpoints should be called")
	assert.Equal(t, 1, outcome.Version)
	assert.Equal(t, int64(10), outcome.ID)
}

func TestRecordOutcome_AwaitingSignoffSession(t *testing.T) {
	deps := newOutcomeTestDeps()
	deps.setupAwaitingSignoffSession(1)

	deps.outcomeRepo.ExistsForSessionFn = func(_ context.Context, _ int64) (bool, error) {
		return false, nil
	}

	deps.outcomeRepo.CreateFn = func(_ context.Context, outcome *domain.SessionOutcome) error {
		outcome.ID = 11
		return nil
	}

	outcome := &domain.SessionOutcome{
		SessionID:     1,
		OutcomeStatus: domain.OutcomeStatusCompleted,
		CreatedBy:     5,
		UpdatedBy:     5,
	}

	err := deps.svc.RecordOutcome(context.Background(), outcome)

	require.NoError(t, err)
}

func TestRecordOutcome_DraftSession(t *testing.T) {
	deps := newOutcomeTestDeps()

	deps.sessionRepo.GetByIDFn = func(_ context.Context, id int64) (*domain.Session, error) {
		return &domain.Session{ID: id, Status: domain.SessionStatusDraft}, nil
	}

	outcome := &domain.SessionOutcome{
		SessionID:     1,
		OutcomeStatus: domain.OutcomeStatusCompleted,
	}

	err := deps.svc.RecordOutcome(context.Background(), outcome)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrSessionNotReady))
}

func TestRecordOutcome_SignedSession(t *testing.T) {
	deps := newOutcomeTestDeps()

	deps.sessionRepo.GetByIDFn = func(_ context.Context, id int64) (*domain.Session, error) {
		return &domain.Session{ID: id, Status: domain.SessionStatusSigned}, nil
	}

	outcome := &domain.SessionOutcome{
		SessionID:     1,
		OutcomeStatus: domain.OutcomeStatusCompleted,
	}

	err := deps.svc.RecordOutcome(context.Background(), outcome)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrSessionNotReady))
}

func TestRecordOutcome_AlreadyExists(t *testing.T) {
	deps := newOutcomeTestDeps()
	deps.setupInProgressSession(1)

	deps.outcomeRepo.ExistsForSessionFn = func(_ context.Context, _ int64) (bool, error) {
		return true, nil
	}

	outcome := &domain.SessionOutcome{
		SessionID:     1,
		OutcomeStatus: domain.OutcomeStatusCompleted,
	}

	err := deps.svc.RecordOutcome(context.Background(), outcome)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrOutcomeAlreadyExists))
}

func TestRecordOutcome_InvalidStatus(t *testing.T) {
	deps := newOutcomeTestDeps()

	outcome := &domain.SessionOutcome{
		SessionID:     1,
		OutcomeStatus: "invalid_status",
	}

	err := deps.svc.RecordOutcome(context.Background(), outcome)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrInvalidOutcomeData))
}

func TestRecordOutcome_AftercareWithoutRedFlags(t *testing.T) {
	deps := newOutcomeTestDeps()

	aftercare := "Rest for 24h"
	outcome := &domain.SessionOutcome{
		SessionID:      1,
		OutcomeStatus:  domain.OutcomeStatusCompleted,
		AftercareNotes: &aftercare,
		RedFlagsText:   nil, // missing red flags
	}

	err := deps.svc.RecordOutcome(context.Background(), outcome)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrInvalidOutcomeData))
}

func TestRecordOutcome_RedFlagsWithAftercareOK(t *testing.T) {
	deps := newOutcomeTestDeps()
	deps.setupInProgressSession(1)

	deps.outcomeRepo.ExistsForSessionFn = func(_ context.Context, _ int64) (bool, error) {
		return false, nil
	}
	deps.outcomeRepo.CreateFn = func(_ context.Context, outcome *domain.SessionOutcome) error {
		outcome.ID = 12
		return nil
	}

	aftercare := "Apply ice"
	redFlags := "Seek help if bleeding"
	outcome := &domain.SessionOutcome{
		SessionID:      1,
		OutcomeStatus:  domain.OutcomeStatusCompleted,
		AftercareNotes: &aftercare,
		RedFlagsText:   &redFlags,
		CreatedBy:      5,
		UpdatedBy:      5,
	}

	err := deps.svc.RecordOutcome(context.Background(), outcome)

	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// GetOutcome tests
// ---------------------------------------------------------------------------

func TestGetOutcome_Success(t *testing.T) {
	deps := newOutcomeTestDeps()

	deps.outcomeRepo.GetBySessionIDFn = func(_ context.Context, sessionID int64) (*domain.SessionOutcome, error) {
		assert.Equal(t, int64(1), sessionID)
		return &domain.SessionOutcome{
			ID:            10,
			SessionID:     1,
			OutcomeStatus: domain.OutcomeStatusCompleted,
			Version:       1,
		}, nil
	}

	deps.outcomeRepo.GetEndpointsFn = func(_ context.Context, outcomeID int64) ([]int64, error) {
		assert.Equal(t, int64(10), outcomeID)
		return []int64{1, 2, 3}, nil
	}

	result, err := deps.svc.GetBySessionID(context.Background(), 1)

	require.NoError(t, err)
	assert.Equal(t, int64(10), result.ID)
	assert.Equal(t, domain.OutcomeStatusCompleted, result.OutcomeStatus)
	assert.Equal(t, []int64{1, 2, 3}, result.EndpointIDs)
}

func TestGetOutcome_NotFound(t *testing.T) {
	deps := newOutcomeTestDeps()

	deps.outcomeRepo.GetBySessionIDFn = func(_ context.Context, _ int64) (*domain.SessionOutcome, error) {
		return nil, service.ErrOutcomeNotFound
	}

	_, err := deps.svc.GetBySessionID(context.Background(), 999)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrOutcomeNotFound))
}

// ---------------------------------------------------------------------------
// UpdateOutcome tests
// ---------------------------------------------------------------------------

func TestUpdateOutcome_Success(t *testing.T) {
	deps := newOutcomeTestDeps()

	updateCalled := false
	deps.outcomeRepo.UpdateFn = func(_ context.Context, outcome *domain.SessionOutcome) error {
		updateCalled = true
		return nil
	}

	endpointsCalled := false
	deps.outcomeRepo.SetEndpointsFn = func(_ context.Context, outcomeID int64, endpointIDs []int64) error {
		endpointsCalled = true
		assert.Equal(t, int64(10), outcomeID)
		assert.Equal(t, []int64{4, 5}, endpointIDs)
		return nil
	}

	aftercare := "Updated aftercare"
	redFlags := "Updated red flags"
	outcome := &domain.SessionOutcome{
		ID:             10,
		SessionID:      1,
		OutcomeStatus:  domain.OutcomeStatusPartial,
		EndpointIDs:    []int64{4, 5},
		AftercareNotes: &aftercare,
		RedFlagsText:   &redFlags,
		Version:        1,
		UpdatedBy:      7,
	}

	err := deps.svc.UpdateOutcome(context.Background(), outcome)

	require.NoError(t, err)
	assert.True(t, updateCalled, "outcome repo Update should be called")
	assert.True(t, endpointsCalled, "outcome repo SetEndpoints should be called")
}

func TestUpdateOutcome_AftercareRedFlagsCoupling(t *testing.T) {
	deps := newOutcomeTestDeps()

	aftercare := "Updated aftercare"
	outcome := &domain.SessionOutcome{
		ID:             10,
		SessionID:      1,
		OutcomeStatus:  domain.OutcomeStatusCompleted,
		AftercareNotes: &aftercare,
		RedFlagsText:   nil, // missing red flags
		Version:        1,
	}

	err := deps.svc.UpdateOutcome(context.Background(), outcome)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrInvalidOutcomeData))
}

func TestUpdateOutcome_VersionConflict(t *testing.T) {
	deps := newOutcomeTestDeps()

	deps.outcomeRepo.UpdateFn = func(_ context.Context, _ *domain.SessionOutcome) error {
		return service.ErrOutcomeNotFound
	}

	outcome := &domain.SessionOutcome{
		ID:            10,
		SessionID:     1,
		OutcomeStatus: domain.OutcomeStatusCompleted,
		Version:       1,
	}

	err := deps.svc.UpdateOutcome(context.Background(), outcome)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrOutcomeNotFound))
}
