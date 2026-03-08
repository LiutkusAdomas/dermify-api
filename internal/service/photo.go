package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path"

	"dermify-api/internal/domain"
)

// Sentinel errors for photo operations.
var (
	ErrPhotoNotFound          = errors.New("photo not found")                        //nolint:gochecknoglobals // sentinel error
	ErrPhotoConsentRequired   = errors.New("photo consent required")                 //nolint:gochecknoglobals // sentinel error
	ErrPhotoInvalidType       = errors.New("invalid photo type")                     //nolint:gochecknoglobals // sentinel error
	ErrPhotoSessionNotEditable = errors.New("session not editable for photo operations") //nolint:gochecknoglobals // sentinel error
	ErrPhotoModuleRequired    = errors.New("module ID required for label photos")    //nolint:gochecknoglobals // sentinel error
	ErrPhotoInvalidContentType = errors.New("invalid content type")                  //nolint:gochecknoglobals // sentinel error
)

// FileStore abstracts filesystem operations for photo storage.
type FileStore interface {
	// Save writes data from reader to the given relative path and returns bytes written.
	Save(ctx context.Context, relPath string, reader io.Reader) (int64, error)
	// Delete removes a file at the given relative path.
	Delete(ctx context.Context, relPath string) error
	// Exists checks whether a file exists at the given relative path.
	Exists(ctx context.Context, relPath string) (bool, error)
}

// PhotoRepository defines the data access contract for photo metadata.
type PhotoRepository interface {
	// Create inserts a new photo record and sets the ID on the provided struct.
	Create(ctx context.Context, photo *domain.Photo) error
	// GetByID retrieves a photo by ID.
	GetByID(ctx context.Context, id int64) (*domain.Photo, error)
	// ListBySession returns all photos for a session.
	ListBySession(ctx context.Context, sessionID int64) ([]*domain.Photo, error)
	// ListByModule returns all photos for a module.
	ListByModule(ctx context.Context, moduleID int64) ([]*domain.Photo, error)
	// Delete removes a photo record by ID.
	Delete(ctx context.Context, id int64) error
}

// PhotoService handles photo business logic including consent enforcement,
// organized file naming, and file lifecycle management.
type PhotoService struct {
	repo        PhotoRepository
	sessionRepo SessionRepository
	moduleRepo  ModuleRepository
	fileStore   FileStore
}

// NewPhotoService creates a new PhotoService with the given dependencies.
func NewPhotoService(repo PhotoRepository, sessionRepo SessionRepository, moduleRepo ModuleRepository, fileStore FileStore) *PhotoService {
	return &PhotoService{
		repo:        repo,
		sessionRepo: sessionRepo,
		moduleRepo:  moduleRepo,
		fileStore:   fileStore,
	}
}

// UploadPhoto validates consent and session state, saves the file, and creates
// the photo metadata record. On metadata failure, it attempts to clean up the
// orphaned file.
func (s *PhotoService) UploadPhoto(ctx context.Context, photo *domain.Photo, reader io.Reader) error {
	session, err := s.sessionRepo.GetByID(ctx, photo.SessionID)
	if err != nil {
		return fmt.Errorf("fetching session: %w", err)
	}

	if session.PhotoConsent == nil || *session.PhotoConsent == domain.PhotoConsentNo {
		return ErrPhotoConsentRequired
	}

	if !isPhotoSessionEditable(session.Status) {
		return ErrPhotoSessionNotEditable
	}

	if photo.PhotoType != domain.PhotoTypeBefore && photo.PhotoType != domain.PhotoTypeLabel {
		return ErrPhotoInvalidType
	}

	if photo.PhotoType == domain.PhotoTypeLabel && photo.ModuleID == nil {
		return ErrPhotoModuleRequired
	}

	if !isAllowedContentType(photo.ContentType) {
		return ErrPhotoInvalidContentType
	}

	relPath, err := generatePhotoPath(photo.SessionID, photo.PhotoType, photo.ContentType)
	if err != nil {
		return fmt.Errorf("generating file path: %w", err)
	}

	photo.FilePath = relPath

	written, err := s.fileStore.Save(ctx, relPath, reader)
	if err != nil {
		return fmt.Errorf("saving file: %w", err)
	}

	photo.SizeBytes = written

	if err := s.repo.Create(ctx, photo); err != nil {
		if delErr := s.fileStore.Delete(ctx, relPath); delErr != nil {
			slog.Error("failed to clean up orphaned file",
				"path", relPath,
				"error", delErr,
			)
		}

		return fmt.Errorf("creating photo record: %w", err)
	}

	return nil
}

// GetByID retrieves a photo by ID.
func (s *PhotoService) GetByID(ctx context.Context, id int64) (*domain.Photo, error) {
	return s.repo.GetByID(ctx, id)
}

// ListBySession returns all photos for a session.
func (s *PhotoService) ListBySession(ctx context.Context, sessionID int64) ([]*domain.Photo, error) {
	return s.repo.ListBySession(ctx, sessionID)
}

// ListByModule returns all photos for a module.
func (s *PhotoService) ListByModule(ctx context.Context, moduleID int64) ([]*domain.Photo, error) {
	return s.repo.ListByModule(ctx, moduleID)
}

// DeletePhoto removes a photo's file and metadata after verifying session editability.
func (s *PhotoService) DeletePhoto(ctx context.Context, id int64, _ int64) error {
	photo, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	session, err := s.sessionRepo.GetByID(ctx, photo.SessionID)
	if err != nil {
		return fmt.Errorf("fetching session: %w", err)
	}

	if !isPhotoSessionEditable(session.Status) {
		return ErrPhotoSessionNotEditable
	}

	if err := s.fileStore.Delete(ctx, photo.FilePath); err != nil {
		return fmt.Errorf("deleting file: %w", err)
	}

	return s.repo.Delete(ctx, id)
}

// isPhotoSessionEditable returns true if the session status allows photo operations.
func isPhotoSessionEditable(status string) bool {
	return status == domain.SessionStatusDraft || status == domain.SessionStatusInProgress
}

// isAllowedContentType checks whether the given MIME type is in the allowed list.
func isAllowedContentType(ct string) bool {
	for _, allowed := range domain.AllowedPhotoContentTypes {
		if ct == allowed {
			return true
		}
	}

	return false
}

// generatePhotoPath creates an organized relative file path for a photo.
// Format: sessions/{session_id}/{photo_type}/{hex_id}.{ext}
func generatePhotoPath(sessionID int64, photoType string, contentType string) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating random bytes: %w", err)
	}

	hexName := hex.EncodeToString(b)

	ext := ".jpg"
	if contentType == "image/png" {
		ext = ".png"
	}

	return path.Join(
		"sessions",
		fmt.Sprintf("%d", sessionID),
		photoType,
		hexName+ext,
	), nil
}
