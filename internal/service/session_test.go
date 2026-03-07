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

// newSessionService creates a SessionService with the given mock repositories.
// If consentRepo or moduleRepo are nil, empty mocks are used.
func newSessionService(
	sessionRepo *testutil.MockSessionRepository,
	consentRepo *testutil.MockConsentRepository,
	moduleRepo *testutil.MockModuleRepository,
) *service.SessionService {
	if consentRepo == nil {
		consentRepo = &testutil.MockConsentRepository{}
	}
	if moduleRepo == nil {
		moduleRepo = &testutil.MockModuleRepository{}
	}
	return service.NewSessionService(sessionRepo, consentRepo, moduleRepo)
}

// validSession returns a session with all required fields populated.
func validSession() *domain.Session {
	return &domain.Session{
		PatientID:   1,
		ClinicianID: 2,
	}
}

// --- Create tests ---

func TestCreateSession_Valid(t *testing.T) {
	repoCalled := false
	indicationCalled := false

	mock := &testutil.MockSessionRepository{
		CreateFn: func(_ context.Context, session *domain.Session) error {
			repoCalled = true
			session.ID = 10
			return nil
		},
		SetIndicationCodesFn: func(_ context.Context, sessionID int64, codeIDs []int64) error {
			indicationCalled = true
			assert.Equal(t, int64(10), sessionID)
			assert.Equal(t, []int64{100, 200}, codeIDs)
			return nil
		},
	}

	svc := newSessionService(mock, nil, nil)
	session := validSession()
	session.IndicationCodes = []int64{100, 200}

	err := svc.Create(context.Background(), session)

	require.NoError(t, err)
	assert.True(t, repoCalled, "repository Create should be called")
	assert.True(t, indicationCalled, "SetIndicationCodes should be called when codes are provided")
	assert.Equal(t, domain.SessionStatusDraft, session.Status)
	assert.Equal(t, 1, session.Version)
	assert.False(t, session.CreatedAt.IsZero(), "created_at should be set")
	assert.False(t, session.UpdatedAt.IsZero(), "updated_at should be set")
	assert.Equal(t, session.ClinicianID, session.CreatedBy)
	assert.Equal(t, session.ClinicianID, session.UpdatedBy)
}

func TestCreateSession_ValidNoIndicationCodes(t *testing.T) {
	repoCalled := false

	mock := &testutil.MockSessionRepository{
		CreateFn: func(_ context.Context, session *domain.Session) error {
			repoCalled = true
			session.ID = 10
			return nil
		},
	}

	svc := newSessionService(mock, nil, nil)
	session := validSession()

	err := svc.Create(context.Background(), session)

	require.NoError(t, err)
	assert.True(t, repoCalled)
	assert.Equal(t, domain.SessionStatusDraft, session.Status)
}

func TestCreateSession_MissingPatientID(t *testing.T) {
	mock := &testutil.MockSessionRepository{}
	svc := newSessionService(mock, nil, nil)

	session := &domain.Session{
		PatientID:   0,
		ClinicianID: 2,
	}

	err := svc.Create(context.Background(), session)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrInvalidSessionData))
}

func TestCreateSession_MissingClinicianID(t *testing.T) {
	mock := &testutil.MockSessionRepository{}
	svc := newSessionService(mock, nil, nil)

	session := &domain.Session{
		PatientID:   1,
		ClinicianID: 0,
	}

	err := svc.Create(context.Background(), session)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrInvalidSessionData))
}

func TestCreateSession_InvalidFitzpatrick(t *testing.T) {
	mock := &testutil.MockSessionRepository{}
	svc := newSessionService(mock, nil, nil)

	tests := []struct {
		name  string
		value int
	}{
		{"too low", 0},
		{"too high", 7},
		{"negative", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := validSession()
			v := tt.value
			session.FitzpatrickType = &v

			err := svc.Create(context.Background(), session)

			require.Error(t, err)
			assert.True(t, errors.Is(err, service.ErrInvalidSessionData))
		})
	}
}

func TestCreateSession_ValidFitzpatrick(t *testing.T) {
	mock := &testutil.MockSessionRepository{
		CreateFn: func(_ context.Context, session *domain.Session) error {
			session.ID = 10
			return nil
		},
	}
	svc := newSessionService(mock, nil, nil)

	for _, v := range []int{1, 3, 6} {
		t.Run("fitzpatrick_"+string(rune('0'+v)), func(t *testing.T) {
			session := validSession()
			fitz := v
			session.FitzpatrickType = &fitz

			err := svc.Create(context.Background(), session)
			require.NoError(t, err)
		})
	}
}

func TestCreateSession_InvalidPhotoConsent(t *testing.T) {
	mock := &testutil.MockSessionRepository{}
	svc := newSessionService(mock, nil, nil)

	session := validSession()
	invalid := "maybe"
	session.PhotoConsent = &invalid

	err := svc.Create(context.Background(), session)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrInvalidSessionData))
}

func TestCreateSession_ValidPhotoConsent(t *testing.T) {
	mock := &testutil.MockSessionRepository{
		CreateFn: func(_ context.Context, session *domain.Session) error {
			session.ID = 10
			return nil
		},
	}
	svc := newSessionService(mock, nil, nil)

	for _, consent := range []string{"yes", "no", "limited"} {
		t.Run("consent_"+consent, func(t *testing.T) {
			session := validSession()
			c := consent
			session.PhotoConsent = &c

			err := svc.Create(context.Background(), session)
			require.NoError(t, err)
		})
	}
}

// --- TransitionState tests ---

func TestTransitionState_ValidDraftToInProgress(t *testing.T) {
	mock := &testutil.MockSessionRepository{
		GetByIDFn: func(_ context.Context, id int64) (*domain.Session, error) {
			return &domain.Session{
				ID:      id,
				Status:  domain.SessionStatusDraft,
				Version: 1,
			}, nil
		},
		UpdateStatusFn: func(_ context.Context, _ int64, status string, _ int, _ int64) error {
			assert.Equal(t, domain.SessionStatusInProgress, status)
			return nil
		},
	}

	svc := newSessionService(mock, nil, nil)
	err := svc.TransitionState(context.Background(), 1, domain.SessionStatusInProgress, 99)

	require.NoError(t, err)
}

func TestTransitionState_ValidInProgressToAwaitingSignoff(t *testing.T) {
	mock := &testutil.MockSessionRepository{
		GetByIDFn: func(_ context.Context, id int64) (*domain.Session, error) {
			return &domain.Session{
				ID:      id,
				Status:  domain.SessionStatusInProgress,
				Version: 2,
			}, nil
		},
		UpdateStatusFn: func(_ context.Context, _ int64, status string, _ int, _ int64) error {
			assert.Equal(t, domain.SessionStatusAwaitingSignoff, status)
			return nil
		},
	}

	svc := newSessionService(mock, nil, nil)
	err := svc.TransitionState(context.Background(), 1, domain.SessionStatusAwaitingSignoff, 99)

	require.NoError(t, err)
}

func TestTransitionState_InvalidDraftToSigned(t *testing.T) {
	mock := &testutil.MockSessionRepository{
		GetByIDFn: func(_ context.Context, id int64) (*domain.Session, error) {
			return &domain.Session{
				ID:      id,
				Status:  domain.SessionStatusDraft,
				Version: 1,
			}, nil
		},
	}

	svc := newSessionService(mock, nil, nil)
	err := svc.TransitionState(context.Background(), 1, domain.SessionStatusSigned, 99)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrInvalidStateTransition))
}

func TestTransitionState_InvalidLockedToAny(t *testing.T) {
	mock := &testutil.MockSessionRepository{
		GetByIDFn: func(_ context.Context, id int64) (*domain.Session, error) {
			return &domain.Session{
				ID:      id,
				Status:  domain.SessionStatusLocked,
				Version: 5,
			}, nil
		},
	}

	svc := newSessionService(mock, nil, nil)

	for _, target := range []string{
		domain.SessionStatusDraft,
		domain.SessionStatusInProgress,
		domain.SessionStatusAwaitingSignoff,
		domain.SessionStatusSigned,
	} {
		t.Run("locked_to_"+target, func(t *testing.T) {
			err := svc.TransitionState(context.Background(), 1, target, 99)
			require.Error(t, err)
			assert.True(t, errors.Is(err, service.ErrInvalidStateTransition))
		})
	}
}

func TestTransitionState_AwaitingSignoffBackToInProgress(t *testing.T) {
	mock := &testutil.MockSessionRepository{
		GetByIDFn: func(_ context.Context, id int64) (*domain.Session, error) {
			return &domain.Session{
				ID:      id,
				Status:  domain.SessionStatusAwaitingSignoff,
				Version: 3,
			}, nil
		},
		UpdateStatusFn: func(_ context.Context, _ int64, status string, _ int, _ int64) error {
			assert.Equal(t, domain.SessionStatusInProgress, status)
			return nil
		},
	}

	svc := newSessionService(mock, nil, nil)
	err := svc.TransitionState(context.Background(), 1, domain.SessionStatusInProgress, 99)

	require.NoError(t, err)
}

func TestTransitionState_SessionNotFound(t *testing.T) {
	mock := &testutil.MockSessionRepository{
		GetByIDFn: func(_ context.Context, _ int64) (*domain.Session, error) {
			return nil, service.ErrSessionNotFound
		},
	}

	svc := newSessionService(mock, nil, nil)
	err := svc.TransitionState(context.Background(), 999, domain.SessionStatusInProgress, 99)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrSessionNotFound))
}

// --- Update tests ---

func TestUpdateSession_EditableState(t *testing.T) {
	editableStates := []string{domain.SessionStatusDraft, domain.SessionStatusInProgress}

	for _, status := range editableStates {
		t.Run("update_in_"+status, func(t *testing.T) {
			updateCalled := false

			mock := &testutil.MockSessionRepository{
				GetByIDFn: func(_ context.Context, _ int64) (*domain.Session, error) {
					return &domain.Session{
						ID:      1,
						Status:  status,
						Version: 1,
					}, nil
				},
				UpdateFn: func(_ context.Context, _ *domain.Session) error {
					updateCalled = true
					return nil
				},
			}

			svc := newSessionService(mock, nil, nil)
			session := &domain.Session{
				ID:          1,
				PatientID:   1,
				ClinicianID: 2,
				Version:     1,
			}

			err := svc.Update(context.Background(), session)
			require.NoError(t, err)
			assert.True(t, updateCalled, "repo.Update should be called")
			assert.False(t, session.UpdatedAt.IsZero(), "updated_at should be set")
		})
	}
}

func TestUpdateSession_NonEditableState(t *testing.T) {
	nonEditableStates := []string{
		domain.SessionStatusAwaitingSignoff,
		domain.SessionStatusSigned,
		domain.SessionStatusLocked,
	}

	for _, status := range nonEditableStates {
		t.Run("update_in_"+status, func(t *testing.T) {
			mock := &testutil.MockSessionRepository{
				GetByIDFn: func(_ context.Context, _ int64) (*domain.Session, error) {
					return &domain.Session{
						ID:      1,
						Status:  status,
						Version: 1,
					}, nil
				},
			}

			svc := newSessionService(mock, nil, nil)
			session := &domain.Session{
				ID:      1,
				Version: 1,
			}

			err := svc.Update(context.Background(), session)
			require.Error(t, err)
			assert.True(t, errors.Is(err, service.ErrSessionNotEditable))
		})
	}
}

func TestUpdateSession_VersionConflict(t *testing.T) {
	mock := &testutil.MockSessionRepository{
		GetByIDFn: func(_ context.Context, _ int64) (*domain.Session, error) {
			return &domain.Session{
				ID:      1,
				Status:  domain.SessionStatusDraft,
				Version: 1,
			}, nil
		},
		UpdateFn: func(_ context.Context, _ *domain.Session) error {
			return service.ErrSessionVersionConflict
		},
	}

	svc := newSessionService(mock, nil, nil)
	session := &domain.Session{
		ID:      1,
		Version: 1,
	}

	err := svc.Update(context.Background(), session)
	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrSessionVersionConflict))
}

func TestUpdateSession_InvalidFitzpatrick(t *testing.T) {
	mock := &testutil.MockSessionRepository{
		GetByIDFn: func(_ context.Context, _ int64) (*domain.Session, error) {
			return &domain.Session{
				ID:      1,
				Status:  domain.SessionStatusDraft,
				Version: 1,
			}, nil
		},
	}

	svc := newSessionService(mock, nil, nil)
	badFitz := 10
	session := &domain.Session{
		ID:              1,
		Version:         1,
		FitzpatrickType: &badFitz,
	}

	err := svc.Update(context.Background(), session)
	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrInvalidSessionData))
}

func TestUpdateSession_InvalidPhotoConsent(t *testing.T) {
	mock := &testutil.MockSessionRepository{
		GetByIDFn: func(_ context.Context, _ int64) (*domain.Session, error) {
			return &domain.Session{
				ID:      1,
				Status:  domain.SessionStatusDraft,
				Version: 1,
			}, nil
		},
	}

	svc := newSessionService(mock, nil, nil)
	bad := "whatever"
	session := &domain.Session{
		ID:           1,
		Version:      1,
		PhotoConsent: &bad,
	}

	err := svc.Update(context.Background(), session)
	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrInvalidSessionData))
}

// --- List tests ---

func TestListSessions_PaginationDefaults(t *testing.T) {
	var capturedFilter service.SessionFilter

	mock := &testutil.MockSessionRepository{
		ListFn: func(_ context.Context, filter service.SessionFilter) (*service.SessionListResult, error) {
			capturedFilter = filter
			return &service.SessionListResult{Sessions: []domain.Session{}, Total: 0}, nil
		},
	}

	svc := newSessionService(mock, nil, nil)
	_, err := svc.List(context.Background(), service.SessionFilter{})

	require.NoError(t, err)
	assert.Equal(t, 1, capturedFilter.Page, "default page should be 1")
	assert.Equal(t, 20, capturedFilter.PerPage, "default per_page should be 20")
}

func TestListSessions_WithFilters(t *testing.T) {
	var capturedFilter service.SessionFilter

	mock := &testutil.MockSessionRepository{
		ListFn: func(_ context.Context, filter service.SessionFilter) (*service.SessionListResult, error) {
			capturedFilter = filter
			return &service.SessionListResult{Sessions: []domain.Session{}, Total: 0}, nil
		},
	}

	svc := newSessionService(mock, nil, nil)
	_, err := svc.List(context.Background(), service.SessionFilter{
		PatientID:   5,
		ClinicianID: 10,
		Status:      domain.SessionStatusDraft,
		Page:        2,
		PerPage:     50,
	})

	require.NoError(t, err)
	assert.Equal(t, int64(5), capturedFilter.PatientID)
	assert.Equal(t, int64(10), capturedFilter.ClinicianID)
	assert.Equal(t, domain.SessionStatusDraft, capturedFilter.Status)
	assert.Equal(t, 2, capturedFilter.Page)
	assert.Equal(t, 50, capturedFilter.PerPage)
}

func TestListSessions_PerPageCapped(t *testing.T) {
	var capturedFilter service.SessionFilter

	mock := &testutil.MockSessionRepository{
		ListFn: func(_ context.Context, filter service.SessionFilter) (*service.SessionListResult, error) {
			capturedFilter = filter
			return &service.SessionListResult{Sessions: []domain.Session{}, Total: 0}, nil
		},
	}

	svc := newSessionService(mock, nil, nil)
	_, err := svc.List(context.Background(), service.SessionFilter{
		Page:    1,
		PerPage: 500,
	})

	require.NoError(t, err)
	assert.Equal(t, 100, capturedFilter.PerPage, "per_page should be capped at 100")
}

func TestListByPatient_Delegates(t *testing.T) {
	var capturedPatientID int64

	mock := &testutil.MockSessionRepository{
		ListByPatientFn: func(_ context.Context, patientID int64) ([]domain.SessionSummary, error) {
			capturedPatientID = patientID
			return []domain.SessionSummary{
				{ID: 1, Status: domain.SessionStatusDraft},
				{ID: 2, Status: domain.SessionStatusLocked},
			}, nil
		},
	}

	svc := newSessionService(mock, nil, nil)
	result, err := svc.ListByPatient(context.Background(), 42)

	require.NoError(t, err)
	assert.Equal(t, int64(42), capturedPatientID)
	assert.Len(t, result, 2)
}

// --- GetByID tests ---

func TestGetByID_Delegates(t *testing.T) {
	mock := &testutil.MockSessionRepository{
		GetByIDFn: func(_ context.Context, id int64) (*domain.Session, error) {
			return &domain.Session{
				ID:     id,
				Status: domain.SessionStatusDraft,
			}, nil
		},
	}

	svc := newSessionService(mock, nil, nil)
	session, err := svc.GetByID(context.Background(), 7)

	require.NoError(t, err)
	require.NotNil(t, session)
	assert.Equal(t, int64(7), session.ID)
}

func TestGetByID_NotFound(t *testing.T) {
	mock := &testutil.MockSessionRepository{
		GetByIDFn: func(_ context.Context, _ int64) (*domain.Session, error) {
			return nil, service.ErrSessionNotFound
		},
	}

	svc := newSessionService(mock, nil, nil)
	session, err := svc.GetByID(context.Background(), 999)

	require.Error(t, err)
	assert.Nil(t, session)
	assert.True(t, errors.Is(err, service.ErrSessionNotFound))
}
