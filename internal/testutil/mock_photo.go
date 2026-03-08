package testutil

import (
	"context"

	"dermify-api/internal/domain"
)

// MockPhotoRepository is a test double for service.PhotoRepository.
type MockPhotoRepository struct {
	CreateFn        func(ctx context.Context, photo *domain.Photo) error
	GetByIDFn       func(ctx context.Context, id int64) (*domain.Photo, error)
	ListBySessionFn func(ctx context.Context, sessionID int64) ([]*domain.Photo, error)
	ListByModuleFn  func(ctx context.Context, moduleID int64) ([]*domain.Photo, error)
	DeleteFn        func(ctx context.Context, id int64) error
}

// Create delegates to CreateFn if set, otherwise returns nil.
func (m *MockPhotoRepository) Create(ctx context.Context, photo *domain.Photo) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, photo)
	}
	return nil
}

// GetByID delegates to GetByIDFn if set, otherwise returns nil and nil.
func (m *MockPhotoRepository) GetByID(ctx context.Context, id int64) (*domain.Photo, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, nil
}

// ListBySession delegates to ListBySessionFn if set, otherwise returns empty slice.
func (m *MockPhotoRepository) ListBySession(ctx context.Context, sessionID int64) ([]*domain.Photo, error) {
	if m.ListBySessionFn != nil {
		return m.ListBySessionFn(ctx, sessionID)
	}
	return []*domain.Photo{}, nil
}

// ListByModule delegates to ListByModuleFn if set, otherwise returns empty slice.
func (m *MockPhotoRepository) ListByModule(ctx context.Context, moduleID int64) ([]*domain.Photo, error) {
	if m.ListByModuleFn != nil {
		return m.ListByModuleFn(ctx, moduleID)
	}
	return []*domain.Photo{}, nil
}

// Delete delegates to DeleteFn if set, otherwise returns nil.
func (m *MockPhotoRepository) Delete(ctx context.Context, id int64) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id)
	}
	return nil
}
