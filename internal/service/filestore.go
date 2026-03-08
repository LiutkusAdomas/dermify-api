package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Compile-time assertion: LocalFileStore implements FileStore.
var _ FileStore = (*LocalFileStore)(nil)

// LocalFileStore stores files on the local filesystem under a base directory.
type LocalFileStore struct {
	basePath string
}

// NewLocalFileStore creates a new LocalFileStore rooted at the given base path.
func NewLocalFileStore(basePath string) *LocalFileStore {
	return &LocalFileStore{basePath: basePath}
}

// Save writes data from reader to the given relative path under the base directory.
// It creates intermediate directories as needed and returns the number of bytes written.
func (fs *LocalFileStore) Save(_ context.Context, relPath string, reader io.Reader) (int64, error) {
	fullPath := filepath.Join(fs.basePath, filepath.FromSlash(relPath))

	if err := os.MkdirAll(filepath.Dir(fullPath), 0o750); err != nil {
		return 0, fmt.Errorf("creating directories: %w", err)
	}

	f, err := os.Create(fullPath) //nolint:gosec // path is server-controlled via generatePhotoPath
	if err != nil {
		return 0, fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()

	written, err := io.Copy(f, reader)
	if err != nil {
		os.Remove(fullPath) //nolint:errcheck // best-effort cleanup
		return 0, fmt.Errorf("writing file: %w", err)
	}

	return written, nil
}

// Delete removes a file at the given relative path under the base directory.
func (fs *LocalFileStore) Delete(_ context.Context, relPath string) error {
	fullPath := filepath.Join(fs.basePath, filepath.FromSlash(relPath))

	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("removing file: %w", err)
	}

	return nil
}

// Exists checks whether a file exists at the given relative path under the base directory.
func (fs *LocalFileStore) Exists(_ context.Context, relPath string) (bool, error) {
	fullPath := filepath.Join(fs.basePath, filepath.FromSlash(relPath))

	_, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("checking file: %w", err)
	}

	return true, nil
}
