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

func newSessionServiceForModule(
	sessionRepo *testutil.MockSessionRepository,
	consentRepo *testutil.MockConsentRepository,
	moduleRepo *testutil.MockModuleRepository,
) *service.SessionService {
	return service.NewSessionService(sessionRepo, consentRepo, moduleRepo)
}

// TestAddModule_WithConsent tests that adding a module to a session with consent succeeds.
func TestAddModule_WithConsent(t *testing.T) {
	createCalled := false

	sessionRepo := &testutil.MockSessionRepository{
		GetByIDFn: func(_ context.Context, _ int64) (*domain.Session, error) {
			return &domain.Session{ID: 1, Status: domain.SessionStatusDraft}, nil
		},
	}
	consentRepo := &testutil.MockConsentRepository{
		ExistsForSessionFn: func(_ context.Context, _ int64) (bool, error) {
			return true, nil
		},
	}
	moduleRepo := &testutil.MockModuleRepository{
		NextSortOrderFn: func(_ context.Context, _ int64) (int, error) {
			return 1, nil
		},
		CreateFn: func(_ context.Context, module *domain.SessionModule, _ int64) error {
			createCalled = true
			module.ID = 42
			return nil
		},
	}

	svc := newSessionServiceForModule(sessionRepo, consentRepo, moduleRepo)
	mod, err := svc.AddModule(context.Background(), 1, domain.ModuleTypeIPL, 10)

	require.NoError(t, err)
	assert.True(t, createCalled, "module Create should be called")
	assert.NotNil(t, mod)
	assert.Equal(t, int64(42), mod.ID)
	assert.Equal(t, int64(1), mod.SessionID)
	assert.Equal(t, domain.ModuleTypeIPL, mod.ModuleType)
	assert.Equal(t, 1, mod.SortOrder)
}

// TestAddModule_WithoutConsent tests that adding a module without consent returns ErrConsentRequired.
func TestAddModule_WithoutConsent(t *testing.T) {
	sessionRepo := &testutil.MockSessionRepository{
		GetByIDFn: func(_ context.Context, _ int64) (*domain.Session, error) {
			return &domain.Session{ID: 1, Status: domain.SessionStatusDraft}, nil
		},
	}
	consentRepo := &testutil.MockConsentRepository{
		ExistsForSessionFn: func(_ context.Context, _ int64) (bool, error) {
			return false, nil
		},
	}
	moduleRepo := &testutil.MockModuleRepository{}

	svc := newSessionServiceForModule(sessionRepo, consentRepo, moduleRepo)
	mod, err := svc.AddModule(context.Background(), 1, domain.ModuleTypeIPL, 10)

	require.Error(t, err)
	assert.Nil(t, mod)
	assert.True(t, errors.Is(err, service.ErrConsentRequired))
}

// TestAddModule_NonEditableSession tests that adding a module to a signed session fails.
func TestAddModule_NonEditableSession(t *testing.T) {
	nonEditableStatuses := []string{
		domain.SessionStatusAwaitingSignoff,
		domain.SessionStatusSigned,
		domain.SessionStatusLocked,
	}

	for _, status := range nonEditableStatuses {
		t.Run(status, func(t *testing.T) {
			sessionRepo := &testutil.MockSessionRepository{
				GetByIDFn: func(_ context.Context, _ int64) (*domain.Session, error) {
					return &domain.Session{ID: 1, Status: status}, nil
				},
			}
			consentRepo := &testutil.MockConsentRepository{}
			moduleRepo := &testutil.MockModuleRepository{}

			svc := newSessionServiceForModule(sessionRepo, consentRepo, moduleRepo)
			mod, err := svc.AddModule(context.Background(), 1, domain.ModuleTypeIPL, 10)

			require.Error(t, err)
			assert.Nil(t, mod)
			assert.True(t, errors.Is(err, service.ErrSessionNotEditable))
		})
	}
}

// TestAddModule_InvalidModuleType tests that adding a module with unknown type returns error.
func TestAddModule_InvalidModuleType(t *testing.T) {
	sessionRepo := &testutil.MockSessionRepository{
		GetByIDFn: func(_ context.Context, _ int64) (*domain.Session, error) {
			return &domain.Session{ID: 1, Status: domain.SessionStatusDraft}, nil
		},
	}
	consentRepo := &testutil.MockConsentRepository{}
	moduleRepo := &testutil.MockModuleRepository{}

	svc := newSessionServiceForModule(sessionRepo, consentRepo, moduleRepo)
	mod, err := svc.AddModule(context.Background(), 1, "unknown_type", 10)

	require.Error(t, err)
	assert.Nil(t, mod)
	assert.True(t, errors.Is(err, service.ErrInvalidSessionData))
}

// TestListModules tests delegating module listing to the repository.
func TestListModules(t *testing.T) {
	expected := []domain.SessionModule{
		{ID: 1, SessionID: 1, ModuleType: domain.ModuleTypeIPL, SortOrder: 1},
		{ID: 2, SessionID: 1, ModuleType: domain.ModuleTypeCO2, SortOrder: 2},
	}

	moduleRepo := &testutil.MockModuleRepository{
		ListBySessionFn: func(_ context.Context, sessionID int64) ([]domain.SessionModule, error) {
			assert.Equal(t, int64(1), sessionID)
			return expected, nil
		},
	}

	svc := newSessionServiceForModule(&testutil.MockSessionRepository{}, &testutil.MockConsentRepository{}, moduleRepo)
	result, err := svc.ListModules(context.Background(), 1)

	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

// TestRemoveModule_EditableSession tests that removing a module from an editable session succeeds.
func TestRemoveModule_EditableSession(t *testing.T) {
	deleteCalled := false

	sessionRepo := &testutil.MockSessionRepository{
		GetByIDFn: func(_ context.Context, _ int64) (*domain.Session, error) {
			return &domain.Session{ID: 1, Status: domain.SessionStatusDraft}, nil
		},
	}
	moduleRepo := &testutil.MockModuleRepository{
		DeleteFn: func(_ context.Context, id int64, sessionID int64) error {
			deleteCalled = true
			assert.Equal(t, int64(5), id)
			assert.Equal(t, int64(1), sessionID)
			return nil
		},
	}

	svc := newSessionServiceForModule(sessionRepo, &testutil.MockConsentRepository{}, moduleRepo)
	err := svc.RemoveModule(context.Background(), 1, 5, 10)

	require.NoError(t, err)
	assert.True(t, deleteCalled, "module Delete should be called")
}

// TestRemoveModule_NonEditableSession tests that removing a module from a locked session fails.
func TestRemoveModule_NonEditableSession(t *testing.T) {
	sessionRepo := &testutil.MockSessionRepository{
		GetByIDFn: func(_ context.Context, _ int64) (*domain.Session, error) {
			return &domain.Session{ID: 1, Status: domain.SessionStatusLocked}, nil
		},
	}

	svc := newSessionServiceForModule(sessionRepo, &testutil.MockConsentRepository{}, &testutil.MockModuleRepository{})
	err := svc.RemoveModule(context.Background(), 1, 5, 10)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrSessionNotEditable))
}
