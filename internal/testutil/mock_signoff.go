package testutil

import "context"

// MockSignoffRepository is a test double for service.SignoffRepository.
type MockSignoffRepository struct {
	SignOffFn     func(ctx context.Context, id int64, clinicianID int64, expectedVersion int) error
	LockSessionFn func(ctx context.Context, id int64, expectedVersion int, userID int64) error
}

// SignOff delegates to SignOffFn if set, otherwise returns nil.
func (m *MockSignoffRepository) SignOff(ctx context.Context, id int64, clinicianID int64, expectedVersion int) error {
	if m.SignOffFn != nil {
		return m.SignOffFn(ctx, id, clinicianID, expectedVersion)
	}
	return nil
}

// LockSession delegates to LockSessionFn if set, otherwise returns nil.
func (m *MockSignoffRepository) LockSession(ctx context.Context, id int64, expectedVersion int, userID int64) error {
	if m.LockSessionFn != nil {
		return m.LockSessionFn(ctx, id, expectedVersion, userID)
	}
	return nil
}
