package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"dermify-api/internal/domain"
	"dermify-api/internal/service"
	"dermify-api/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validConsent() *domain.Consent {
	return &domain.Consent{
		SessionID:      1,
		ConsentType:    "treatment",
		ConsentMethod:  "written",
		ObtainedAt:     time.Date(2026, 3, 7, 10, 0, 0, 0, time.UTC),
		RisksDiscussed: true,
		CreatedBy:      1,
		UpdatedBy:      1,
	}
}

// TestRecordConsent_Valid tests that recording consent with valid data succeeds.
func TestRecordConsent_Valid(t *testing.T) {
	createCalled := false

	mock := &testutil.MockConsentRepository{
		ExistsForSessionFn: func(_ context.Context, _ int64) (bool, error) {
			return false, nil
		},
		CreateFn: func(_ context.Context, consent *domain.Consent) error {
			createCalled = true
			consent.ID = 10
			return nil
		},
	}

	svc := service.NewConsentService(mock)
	consent := validConsent()

	err := svc.RecordConsent(context.Background(), consent)

	require.NoError(t, err)
	assert.True(t, createCalled, "repository Create should be called")
	assert.Equal(t, int64(10), consent.ID)
	assert.Equal(t, 1, consent.Version)
	assert.False(t, consent.CreatedAt.IsZero(), "created_at should be set")
	assert.False(t, consent.UpdatedAt.IsZero(), "updated_at should be set")
}

// TestRecordConsent_MissingType tests that empty consent_type returns ErrInvalidConsentData.
func TestRecordConsent_MissingType(t *testing.T) {
	mock := &testutil.MockConsentRepository{}
	svc := service.NewConsentService(mock)

	consent := validConsent()
	consent.ConsentType = ""

	err := svc.RecordConsent(context.Background(), consent)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrInvalidConsentData))
}

// TestRecordConsent_MissingMethod tests that empty consent_method returns ErrInvalidConsentData.
func TestRecordConsent_MissingMethod(t *testing.T) {
	mock := &testutil.MockConsentRepository{}
	svc := service.NewConsentService(mock)

	consent := validConsent()
	consent.ConsentMethod = ""

	err := svc.RecordConsent(context.Background(), consent)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrInvalidConsentData))
}

// TestRecordConsent_AlreadyExists tests that recording consent for a session that
// already has consent returns ErrConsentAlreadyExists.
func TestRecordConsent_AlreadyExists(t *testing.T) {
	mock := &testutil.MockConsentRepository{
		ExistsForSessionFn: func(_ context.Context, _ int64) (bool, error) {
			return true, nil
		},
	}

	svc := service.NewConsentService(mock)
	consent := validConsent()

	err := svc.RecordConsent(context.Background(), consent)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrConsentAlreadyExists))
}

// TestUpdateConsent_Valid tests that updating consent with valid data succeeds.
func TestUpdateConsent_Valid(t *testing.T) {
	updateCalled := false

	mock := &testutil.MockConsentRepository{
		UpdateFn: func(_ context.Context, _ *domain.Consent) error {
			updateCalled = true
			return nil
		},
	}

	svc := service.NewConsentService(mock)
	consent := validConsent()
	consent.ID = 5
	consent.Version = 1

	before := time.Now()
	err := svc.UpdateConsent(context.Background(), consent)
	after := time.Now()

	require.NoError(t, err)
	assert.True(t, updateCalled, "repository Update should be called")
	assert.True(t, !consent.UpdatedAt.Before(before) && !consent.UpdatedAt.After(after),
		"updated_at should be set to current time")
}

// TestUpdateConsent_MissingType tests that updating with empty consent_type returns error.
func TestUpdateConsent_MissingType(t *testing.T) {
	mock := &testutil.MockConsentRepository{}
	svc := service.NewConsentService(mock)

	consent := validConsent()
	consent.ConsentType = ""

	err := svc.UpdateConsent(context.Background(), consent)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrInvalidConsentData))
}

// TestGetConsentBySessionID tests delegating retrieval to the repository.
func TestGetConsentBySessionID(t *testing.T) {
	expected := validConsent()
	expected.ID = 7

	mock := &testutil.MockConsentRepository{
		GetBySessionIDFn: func(_ context.Context, sessionID int64) (*domain.Consent, error) {
			assert.Equal(t, int64(1), sessionID)
			return expected, nil
		},
	}

	svc := service.NewConsentService(mock)
	result, err := svc.GetBySessionID(context.Background(), 1)

	require.NoError(t, err)
	assert.Equal(t, expected, result)
}
