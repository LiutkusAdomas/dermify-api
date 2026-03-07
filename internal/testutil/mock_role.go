package testutil

import (
	"context"
)

// MockRoleRepository is a test double for service.RoleRepository.
type MockRoleRepository struct {
	UpdateUserRoleFn func(ctx context.Context, userID int64, role string) error
	GetUserRoleFn    func(ctx context.Context, userID int64) (string, error)
	CountUsersFn     func(ctx context.Context) (int64, error)
}

// UpdateUserRole delegates to UpdateUserRoleFn if set, otherwise returns nil.
func (m *MockRoleRepository) UpdateUserRole(ctx context.Context, userID int64, role string) error {
	if m.UpdateUserRoleFn != nil {
		return m.UpdateUserRoleFn(ctx, userID, role)
	}
	return nil
}

// GetUserRole delegates to GetUserRoleFn if set, otherwise returns empty string and nil.
func (m *MockRoleRepository) GetUserRole(ctx context.Context, userID int64) (string, error) {
	if m.GetUserRoleFn != nil {
		return m.GetUserRoleFn(ctx, userID)
	}
	return "", nil
}

// CountUsers delegates to CountUsersFn if set, otherwise returns 0 and nil.
func (m *MockRoleRepository) CountUsers(ctx context.Context) (int64, error) {
	if m.CountUsersFn != nil {
		return m.CountUsersFn(ctx)
	}
	return 0, nil
}
