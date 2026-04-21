package testutil

import (
	"context"

	"dermify-api/internal/domain"
)

// MockUserRepository is a test double for service.UserRepository.
type MockUserRepository struct {
	CreateFn                func(ctx context.Context, user *domain.User) error
	GetByIDFn               func(ctx context.Context, id int64) (*domain.User, error)
	GetByEmailFn            func(ctx context.Context, email string) (*domain.User, error)
	UpdateFn                func(ctx context.Context, user *domain.User) error
	DeleteFn                func(ctx context.Context, id int64) error
	ListFn                  func(ctx context.Context) ([]*domain.User, error)
	UpdatePreferencesFn     func(ctx context.Context, userID int64, language, timezone string) error
	UpdatePasswordFn        func(ctx context.Context, userID int64, passwordHash string, clearMustChange bool) error
	SetMustChangePasswordFn func(ctx context.Context, userID int64, mustChange bool) error
}

// Create delegates to CreateFn if set, otherwise returns nil.
func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, user)
	}
	return nil
}

// GetByID delegates to GetByIDFn if set, otherwise returns nil and nil.
func (m *MockUserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, nil
}

// GetByEmail delegates to GetByEmailFn if set, otherwise returns nil and nil.
func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.GetByEmailFn != nil {
		return m.GetByEmailFn(ctx, email)
	}
	return nil, nil
}

// Update delegates to UpdateFn if set, otherwise returns nil.
func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, user)
	}
	return nil
}

// Delete delegates to DeleteFn if set, otherwise returns nil.
func (m *MockUserRepository) Delete(ctx context.Context, id int64) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id)
	}
	return nil
}

// List delegates to ListFn if set, otherwise returns empty slice.
func (m *MockUserRepository) List(ctx context.Context) ([]*domain.User, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx)
	}
	return []*domain.User{}, nil
}

// UpdatePreferences delegates to UpdatePreferencesFn if set, otherwise returns nil.
func (m *MockUserRepository) UpdatePreferences(ctx context.Context, userID int64, language, timezone string) error {
	if m.UpdatePreferencesFn != nil {
		return m.UpdatePreferencesFn(ctx, userID, language, timezone)
	}
	return nil
}

// UpdatePassword delegates to UpdatePasswordFn if set, otherwise returns nil.
func (m *MockUserRepository) UpdatePassword(ctx context.Context, userID int64, passwordHash string, clearMustChange bool) error {
	if m.UpdatePasswordFn != nil {
		return m.UpdatePasswordFn(ctx, userID, passwordHash, clearMustChange)
	}
	return nil
}

// SetMustChangePassword delegates to SetMustChangePasswordFn if set, otherwise returns nil.
func (m *MockUserRepository) SetMustChangePassword(ctx context.Context, userID int64, mustChange bool) error {
	if m.SetMustChangePasswordFn != nil {
		return m.SetMustChangePasswordFn(ctx, userID, mustChange)
	}
	return nil
}
