package testutil

import (
	"context"
	"io"
)

// MockFileStore is a test double for service.FileStore.
type MockFileStore struct {
	SaveFn   func(ctx context.Context, relPath string, reader io.Reader) (int64, error)
	DeleteFn func(ctx context.Context, relPath string) error
	ExistsFn func(ctx context.Context, relPath string) (bool, error)
}

// Save delegates to SaveFn if set, otherwise returns 0 and nil.
func (m *MockFileStore) Save(ctx context.Context, relPath string, reader io.Reader) (int64, error) {
	if m.SaveFn != nil {
		return m.SaveFn(ctx, relPath, reader)
	}
	return 0, nil
}

// Delete delegates to DeleteFn if set, otherwise returns nil.
func (m *MockFileStore) Delete(ctx context.Context, relPath string) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, relPath)
	}
	return nil
}

// Exists delegates to ExistsFn if set, otherwise returns false and nil.
func (m *MockFileStore) Exists(ctx context.Context, relPath string) (bool, error) {
	if m.ExistsFn != nil {
		return m.ExistsFn(ctx, relPath)
	}
	return false, nil
}
