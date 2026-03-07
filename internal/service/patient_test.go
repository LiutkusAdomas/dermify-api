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

func validPatient() *domain.Patient {
	return &domain.Patient{
		FirstName:   "Jane",
		LastName:    "Doe",
		DateOfBirth: time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC),
		Sex:         "female",
		CreatedBy:   1,
		UpdatedBy:   1,
	}
}

// TestCreatePatient_ValidData tests creating a patient with all required fields.
func TestCreatePatient_ValidData(t *testing.T) {
	repoCalled := false

	mock := &testutil.MockPatientRepository{
		CreateFn: func(_ context.Context, patient *domain.Patient) error {
			repoCalled = true
			patient.ID = 42
			return nil
		},
	}

	svc := service.NewPatientService(mock)
	patient := validPatient()
	err := svc.Create(context.Background(), patient)

	require.NoError(t, err)
	assert.True(t, repoCalled, "repository Create should be called")
	assert.Equal(t, int64(42), patient.ID)
	assert.Equal(t, 1, patient.Version)
	assert.False(t, patient.CreatedAt.IsZero(), "created_at should be set")
	assert.False(t, patient.UpdatedAt.IsZero(), "updated_at should be set")
}

// TestCreatePatient_MissingFields tests that creating a patient without required fields returns an error.
func TestCreatePatient_MissingFields(t *testing.T) {
	mock := &testutil.MockPatientRepository{}
	svc := service.NewPatientService(mock)

	tests := []struct {
		name    string
		patient *domain.Patient
	}{
		{
			name: "missing first name",
			patient: &domain.Patient{
				LastName:    "Doe",
				DateOfBirth: time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC),
				Sex:         "female",
			},
		},
		{
			name: "missing last name",
			patient: &domain.Patient{
				FirstName:   "Jane",
				DateOfBirth: time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC),
				Sex:         "female",
			},
		},
		{
			name: "missing date of birth",
			patient: &domain.Patient{
				FirstName: "Jane",
				LastName:  "Doe",
				Sex:       "female",
			},
		},
		{
			name: "missing sex",
			patient: &domain.Patient{
				FirstName:   "Jane",
				LastName:    "Doe",
				DateOfBirth: time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.Create(context.Background(), tt.patient)

			require.Error(t, err)
			assert.True(t, errors.Is(err, service.ErrInvalidPatientData))
		})
	}
}

// TestListPatients_DefaultPagination tests listing patients with default pagination values.
func TestListPatients_DefaultPagination(t *testing.T) {
	var capturedFilter service.PatientFilter

	mock := &testutil.MockPatientRepository{
		ListFn: func(_ context.Context, filter service.PatientFilter) (*service.PatientListResult, error) {
			capturedFilter = filter
			return &service.PatientListResult{Patients: []service.PatientListItem{}, Total: 0}, nil
		},
	}

	svc := service.NewPatientService(mock)
	_, err := svc.List(context.Background(), service.PatientFilter{})

	require.NoError(t, err)
	assert.Equal(t, 1, capturedFilter.Page, "default page should be 1")
	assert.Equal(t, 20, capturedFilter.PerPage, "default per_page should be 20")
}

// TestListPatients_CustomPagination tests listing patients with custom page and per_page values.
func TestListPatients_CustomPagination(t *testing.T) {
	var capturedFilter service.PatientFilter

	mock := &testutil.MockPatientRepository{
		ListFn: func(_ context.Context, filter service.PatientFilter) (*service.PatientListResult, error) {
			capturedFilter = filter
			return &service.PatientListResult{Patients: []service.PatientListItem{}, Total: 0}, nil
		},
	}

	svc := service.NewPatientService(mock)

	// PerPage > 100 should be capped at 100.
	_, err := svc.List(context.Background(), service.PatientFilter{Page: 3, PerPage: 200})

	require.NoError(t, err)
	assert.Equal(t, 3, capturedFilter.Page)
	assert.Equal(t, 100, capturedFilter.PerPage, "per_page should be capped at 100")
}

// TestUpdatePatient_VersionConflict tests that concurrent updates trigger a version conflict error.
func TestUpdatePatient_VersionConflict(t *testing.T) {
	mock := &testutil.MockPatientRepository{
		UpdateFn: func(_ context.Context, _ *domain.Patient) error {
			return service.ErrPatientVersionConflict
		},
	}

	svc := service.NewPatientService(mock)
	patient := validPatient()
	patient.ID = 1
	patient.Version = 1

	err := svc.Update(context.Background(), patient)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrPatientVersionConflict))
}

// TestGetPatientSessions_EmptyList tests that a patient with no sessions returns an empty list.
func TestGetPatientSessions_EmptyList(t *testing.T) {
	mock := &testutil.MockPatientRepository{
		GetSessionHistoryFn: func(_ context.Context, _ int64) ([]domain.SessionSummary, error) {
			return []domain.SessionSummary{}, nil
		},
	}

	svc := service.NewPatientService(mock)
	sessions, err := svc.GetSessionHistory(context.Background(), 1)

	require.NoError(t, err)
	assert.Empty(t, sessions)
}

// TestMetadataTracking tests that created_at, created_by, updated_at, updated_by are populated.
func TestMetadataTracking(t *testing.T) {
	mock := &testutil.MockPatientRepository{
		CreateFn: func(_ context.Context, _ *domain.Patient) error {
			return nil
		},
	}

	svc := service.NewPatientService(mock)
	patient := validPatient()
	patient.CreatedBy = 7
	patient.UpdatedBy = 7

	before := time.Now()
	err := svc.Create(context.Background(), patient)
	after := time.Now()

	require.NoError(t, err)
	assert.Equal(t, int64(7), patient.CreatedBy, "created_by should be preserved")
	assert.Equal(t, int64(7), patient.UpdatedBy, "updated_by should be preserved")
	assert.True(t, !patient.CreatedAt.Before(before) && !patient.CreatedAt.After(after),
		"created_at should be set to current time")
	assert.True(t, !patient.UpdatedAt.Before(before) && !patient.UpdatedAt.After(after),
		"updated_at should be set to current time")
}

// TestVersionIncrement tests that the version number is set on create.
func TestVersionIncrement(t *testing.T) {
	mock := &testutil.MockPatientRepository{
		CreateFn: func(_ context.Context, _ *domain.Patient) error {
			return nil
		},
	}

	svc := service.NewPatientService(mock)
	patient := validPatient()
	patient.Version = 0

	err := svc.Create(context.Background(), patient)

	require.NoError(t, err)
	assert.Equal(t, 1, patient.Version, "version should be set to 1 on creation")
}
