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

func validScreening() *domain.ContraindicationScreening {
	return &domain.ContraindicationScreening{
		SessionID: 1,
		CreatedBy: 1,
		UpdatedBy: 1,
	}
}

// TestRecordScreening_Valid tests that recording a screening with valid data succeeds.
func TestRecordScreening_Valid(t *testing.T) {
	createCalled := false

	mock := &testutil.MockContraindicationRepository{
		GetBySessionIDFn: func(_ context.Context, _ int64) (*domain.ContraindicationScreening, error) {
			return nil, service.ErrScreeningNotFound
		},
		CreateFn: func(_ context.Context, screening *domain.ContraindicationScreening) error {
			createCalled = true
			screening.ID = 10
			return nil
		},
	}

	svc := service.NewContraindicationService(mock)
	screening := validScreening()

	err := svc.RecordScreening(context.Background(), screening)

	require.NoError(t, err)
	assert.True(t, createCalled, "repository Create should be called")
	assert.Equal(t, int64(10), screening.ID)
	assert.Equal(t, 1, screening.Version)
	assert.False(t, screening.CreatedAt.IsZero(), "created_at should be set")
	assert.False(t, screening.UpdatedAt.IsZero(), "updated_at should be set")
}

// TestRecordScreening_AlreadyExists tests that recording a screening for a session that
// already has one returns ErrScreeningAlreadyExists.
func TestRecordScreening_AlreadyExists(t *testing.T) {
	existing := validScreening()
	existing.ID = 5

	mock := &testutil.MockContraindicationRepository{
		GetBySessionIDFn: func(_ context.Context, _ int64) (*domain.ContraindicationScreening, error) {
			return existing, nil
		},
	}

	svc := service.NewContraindicationService(mock)
	screening := validScreening()

	err := svc.RecordScreening(context.Background(), screening)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrScreeningAlreadyExists))
}

// TestRecordScreening_HasFlagsComputed tests that has_flags is automatically set to true
// when any boolean flag is true.
func TestRecordScreening_HasFlagsComputed(t *testing.T) {
	tests := []struct {
		name     string
		setFlag  func(s *domain.ContraindicationScreening)
		expected bool
	}{
		{"pregnant", func(s *domain.ContraindicationScreening) { s.Pregnant = true }, true},
		{"breastfeeding", func(s *domain.ContraindicationScreening) { s.Breastfeeding = true }, true},
		{"active infection", func(s *domain.ContraindicationScreening) { s.ActiveInfection = true }, true},
		{"active cold sores", func(s *domain.ContraindicationScreening) { s.ActiveColdSores = true }, true},
		{"isotretinoin", func(s *domain.ContraindicationScreening) { s.Isotretinoin = true }, true},
		{"photosensitivity", func(s *domain.ContraindicationScreening) { s.Photosensitivity = true }, true},
		{"autoimmune disorder", func(s *domain.ContraindicationScreening) { s.AutoimmuneDisorder = true }, true},
		{"keloid history", func(s *domain.ContraindicationScreening) { s.KeloidHistory = true }, true},
		{"anticoagulants", func(s *domain.ContraindicationScreening) { s.Anticoagulants = true }, true},
		{"recent tan", func(s *domain.ContraindicationScreening) { s.RecentTan = true }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &testutil.MockContraindicationRepository{
				GetBySessionIDFn: func(_ context.Context, _ int64) (*domain.ContraindicationScreening, error) {
					return nil, service.ErrScreeningNotFound
				},
				CreateFn: func(_ context.Context, _ *domain.ContraindicationScreening) error {
					return nil
				},
			}

			svc := service.NewContraindicationService(mock)
			screening := validScreening()
			tt.setFlag(screening)

			err := svc.RecordScreening(context.Background(), screening)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, screening.HasFlags, "has_flags should be true when %s is set", tt.name)
		})
	}
}

// TestRecordScreening_NoFlags tests that has_flags is false when all flags are false.
func TestRecordScreening_NoFlags(t *testing.T) {
	mock := &testutil.MockContraindicationRepository{
		GetBySessionIDFn: func(_ context.Context, _ int64) (*domain.ContraindicationScreening, error) {
			return nil, service.ErrScreeningNotFound
		},
		CreateFn: func(_ context.Context, _ *domain.ContraindicationScreening) error {
			return nil
		},
	}

	svc := service.NewContraindicationService(mock)
	screening := validScreening()

	err := svc.RecordScreening(context.Background(), screening)

	require.NoError(t, err)
	assert.False(t, screening.HasFlags, "has_flags should be false when no flags are set")
}

// TestUpdateScreening_Valid tests that updating a screening succeeds and recomputes has_flags.
func TestUpdateScreening_Valid(t *testing.T) {
	updateCalled := false

	mock := &testutil.MockContraindicationRepository{
		UpdateFn: func(_ context.Context, _ *domain.ContraindicationScreening) error {
			updateCalled = true
			return nil
		},
	}

	svc := service.NewContraindicationService(mock)
	screening := validScreening()
	screening.ID = 5
	screening.Version = 1
	screening.Pregnant = true

	before := time.Now()
	err := svc.UpdateScreening(context.Background(), screening)
	after := time.Now()

	require.NoError(t, err)
	assert.True(t, updateCalled, "repository Update should be called")
	assert.True(t, screening.HasFlags, "has_flags should be recomputed to true")
	assert.True(t, !screening.UpdatedAt.Before(before) && !screening.UpdatedAt.After(after),
		"updated_at should be set to current time")
}

// TestGetScreeningBySessionID tests delegating retrieval to the repository.
func TestGetScreeningBySessionID(t *testing.T) {
	expected := validScreening()
	expected.ID = 7

	mock := &testutil.MockContraindicationRepository{
		GetBySessionIDFn: func(_ context.Context, sessionID int64) (*domain.ContraindicationScreening, error) {
			assert.Equal(t, int64(1), sessionID)
			return expected, nil
		},
	}

	svc := service.NewContraindicationService(mock)
	result, err := svc.GetBySessionID(context.Background(), 1)

	require.NoError(t, err)
	assert.Equal(t, expected, result)
}
