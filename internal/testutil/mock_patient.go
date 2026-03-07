package testutil

import (
	"context"

	"dermify-api/internal/domain"
	"dermify-api/internal/service"
)

// MockPatientRepository is a test double for service.PatientRepository.
type MockPatientRepository struct {
	CreateFn            func(ctx context.Context, patient *domain.Patient) error
	GetByIDFn           func(ctx context.Context, id int64) (*domain.Patient, error)
	UpdateFn            func(ctx context.Context, patient *domain.Patient) error
	ListFn              func(ctx context.Context, filter service.PatientFilter) (*service.PatientListResult, error)
	GetSessionHistoryFn func(ctx context.Context, patientID int64) ([]domain.SessionSummary, error)
}

// Create delegates to CreateFn if set, otherwise returns nil.
func (m *MockPatientRepository) Create(ctx context.Context, patient *domain.Patient) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, patient)
	}
	return nil
}

// GetByID delegates to GetByIDFn if set, otherwise returns nil and nil.
func (m *MockPatientRepository) GetByID(ctx context.Context, id int64) (*domain.Patient, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, nil
}

// Update delegates to UpdateFn if set, otherwise returns nil.
func (m *MockPatientRepository) Update(ctx context.Context, patient *domain.Patient) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, patient)
	}
	return nil
}

// List delegates to ListFn if set, otherwise returns empty result.
func (m *MockPatientRepository) List(ctx context.Context, filter service.PatientFilter) (*service.PatientListResult, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, filter)
	}
	return &service.PatientListResult{Patients: []service.PatientListItem{}, Total: 0}, nil
}

// GetSessionHistory delegates to GetSessionHistoryFn if set, otherwise returns empty slice.
func (m *MockPatientRepository) GetSessionHistory(ctx context.Context, patientID int64) ([]domain.SessionSummary, error) {
	if m.GetSessionHistoryFn != nil {
		return m.GetSessionHistoryFn(ctx, patientID)
	}
	return []domain.SessionSummary{}, nil
}
