package testutil

import (
	"context"

	"dermify-api/internal/domain"
	"dermify-api/internal/service"
)

// MockSessionRepository is a test double for service.SessionRepository.
type MockSessionRepository struct {
	CreateFn             func(ctx context.Context, session *domain.Session) error
	GetByIDFn            func(ctx context.Context, id int64) (*domain.Session, error)
	UpdateFn             func(ctx context.Context, session *domain.Session) error
	UpdateStatusFn       func(ctx context.Context, id int64, status string, expectedVersion int, userID int64) error
	ListFn               func(ctx context.Context, filter service.SessionFilter) (*service.SessionListResult, error)
	ListByPatientFn      func(ctx context.Context, patientID int64) ([]domain.SessionSummary, error)
	SetIndicationCodesFn func(ctx context.Context, sessionID int64, codeIDs []int64) error
}

// Create delegates to CreateFn if set, otherwise returns nil.
func (m *MockSessionRepository) Create(ctx context.Context, session *domain.Session) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, session)
	}
	return nil
}

// GetByID delegates to GetByIDFn if set, otherwise returns nil and nil.
func (m *MockSessionRepository) GetByID(ctx context.Context, id int64) (*domain.Session, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, nil
}

// Update delegates to UpdateFn if set, otherwise returns nil.
func (m *MockSessionRepository) Update(ctx context.Context, session *domain.Session) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, session)
	}
	return nil
}

// UpdateStatus delegates to UpdateStatusFn if set, otherwise returns nil.
func (m *MockSessionRepository) UpdateStatus(ctx context.Context, id int64, status string, expectedVersion int, userID int64) error {
	if m.UpdateStatusFn != nil {
		return m.UpdateStatusFn(ctx, id, status, expectedVersion, userID)
	}
	return nil
}

// List delegates to ListFn if set, otherwise returns empty result.
func (m *MockSessionRepository) List(ctx context.Context, filter service.SessionFilter) (*service.SessionListResult, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, filter)
	}
	return &service.SessionListResult{Sessions: []domain.Session{}, Total: 0}, nil
}

// ListByPatient delegates to ListByPatientFn if set, otherwise returns empty slice.
func (m *MockSessionRepository) ListByPatient(ctx context.Context, patientID int64) ([]domain.SessionSummary, error) {
	if m.ListByPatientFn != nil {
		return m.ListByPatientFn(ctx, patientID)
	}
	return []domain.SessionSummary{}, nil
}

// SetIndicationCodes delegates to SetIndicationCodesFn if set, otherwise returns nil.
func (m *MockSessionRepository) SetIndicationCodes(ctx context.Context, sessionID int64, codeIDs []int64) error {
	if m.SetIndicationCodesFn != nil {
		return m.SetIndicationCodesFn(ctx, sessionID, codeIDs)
	}
	return nil
}

// MockModuleRepository is a test double for service.ModuleRepository.
type MockModuleRepository struct {
	CreateFn        func(ctx context.Context, module *domain.SessionModule, userID int64) error
	ListBySessionFn func(ctx context.Context, sessionID int64) ([]domain.SessionModule, error)
	DeleteFn        func(ctx context.Context, id int64, sessionID int64) error
	NextSortOrderFn func(ctx context.Context, sessionID int64) (int, error)
}

// Create delegates to CreateFn if set, otherwise returns nil.
func (m *MockModuleRepository) Create(ctx context.Context, module *domain.SessionModule, userID int64) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, module, userID)
	}
	return nil
}

// ListBySession delegates to ListBySessionFn if set, otherwise returns empty slice.
func (m *MockModuleRepository) ListBySession(ctx context.Context, sessionID int64) ([]domain.SessionModule, error) {
	if m.ListBySessionFn != nil {
		return m.ListBySessionFn(ctx, sessionID)
	}
	return []domain.SessionModule{}, nil
}

// Delete delegates to DeleteFn if set, otherwise returns nil.
func (m *MockModuleRepository) Delete(ctx context.Context, id int64, sessionID int64) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id, sessionID)
	}
	return nil
}

// NextSortOrder delegates to NextSortOrderFn if set, otherwise returns 0.
func (m *MockModuleRepository) NextSortOrder(ctx context.Context, sessionID int64) (int, error) {
	if m.NextSortOrderFn != nil {
		return m.NextSortOrderFn(ctx, sessionID)
	}
	return 0, nil
}
