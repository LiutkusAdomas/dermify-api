package service_test

import (
	"context"
	"errors"
	"io"
	"regexp"
	"strings"
	"testing"

	"dermify-api/internal/domain"
	"dermify-api/internal/service"
	"dermify-api/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// photoTestDeps holds all mocked dependencies used by PhotoService tests.
type photoTestDeps struct {
	svc         *service.PhotoService
	photoRepo   *testutil.MockPhotoRepository
	sessionRepo *testutil.MockSessionRepository
	moduleRepo  *testutil.MockModuleRepository
	fileStore   *testutil.MockFileStore
}

func newPhotoTestDeps() photoTestDeps {
	photoRepo := &testutil.MockPhotoRepository{}
	sessionRepo := &testutil.MockSessionRepository{}
	moduleRepo := &testutil.MockModuleRepository{}
	fileStore := &testutil.MockFileStore{}

	svc := service.NewPhotoService(photoRepo, sessionRepo, moduleRepo, fileStore)

	return photoTestDeps{
		svc:         svc,
		photoRepo:   photoRepo,
		sessionRepo: sessionRepo,
		moduleRepo:  moduleRepo,
		fileStore:   fileStore,
	}
}

// setupEditableSessionWithConsent configures the mock session repo to return a
// session with status="in_progress" and the given PhotoConsent value.
func (d *photoTestDeps) setupEditableSessionWithConsent(sessionID int64, consent *string) {
	d.sessionRepo.GetByIDFn = func(_ context.Context, id int64) (*domain.Session, error) {
		return &domain.Session{
			ID:           id,
			Status:       domain.SessionStatusInProgress,
			PhotoConsent: consent,
			Version:      1,
		}, nil
	}
}

// setupFileStore configures the mock file store to accept saves.
func (d *photoTestDeps) setupFileStore() {
	d.fileStore.SaveFn = func(_ context.Context, _ string, _ io.Reader) (int64, error) {
		return 15, nil
	}
}

// ---------------------------------------------------------------------------
// photoPathRegex validates organized file naming convention.
// ---------------------------------------------------------------------------
var photoPathRegex = regexp.MustCompile(`^sessions/\d+/(before|label)/[0-9a-f]{32}\.(jpg|png)$`)

// ---------------------------------------------------------------------------
// UploadPhoto tests - Consent gate
// ---------------------------------------------------------------------------

func TestUploadBeforePhoto_Success(t *testing.T) {
	deps := newPhotoTestDeps()
	consent := domain.PhotoConsentYes
	deps.setupEditableSessionWithConsent(1, &consent)

	deps.fileStore.SaveFn = func(_ context.Context, _ string, _ io.Reader) (int64, error) {
		return 15, nil
	}

	var capturedPhoto *domain.Photo
	deps.photoRepo.CreateFn = func(_ context.Context, photo *domain.Photo) error {
		capturedPhoto = photo
		photo.ID = 100
		return nil
	}

	photo := &domain.Photo{
		SessionID:    1,
		PhotoType:    domain.PhotoTypeBefore,
		OriginalName: "test.jpg",
		ContentType:  "image/jpeg",
		CreatedBy:    5,
		UpdatedBy:    5,
	}

	err := deps.svc.UploadPhoto(context.Background(), photo, strings.NewReader("fake image data"))

	require.NoError(t, err)
	require.NotNil(t, capturedPhoto)
	assert.Equal(t, int64(100), photo.ID)
	assert.Equal(t, int64(15), capturedPhoto.SizeBytes)
	assert.True(t, photoPathRegex.MatchString(capturedPhoto.FilePath),
		"FilePath %q should match pattern sessions/{id}/before/{hex}.jpg", capturedPhoto.FilePath)
	assert.Contains(t, capturedPhoto.FilePath, "/before/")
}

func TestUploadBeforePhoto_ConsentVariants(t *testing.T) {
	tests := []struct {
		name    string
		consent *string
		wantErr error
	}{
		{
			name:    "ConsentYes",
			consent: photoStrPtr(domain.PhotoConsentYes),
			wantErr: nil,
		},
		{
			name:    "ConsentLimited",
			consent: photoStrPtr(domain.PhotoConsentLimited),
			wantErr: nil,
		},
		{
			name:    "ConsentNo",
			consent: photoStrPtr(domain.PhotoConsentNo),
			wantErr: service.ErrPhotoConsentRequired,
		},
		{
			name:    "ConsentNil",
			consent: nil,
			wantErr: service.ErrPhotoConsentRequired,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			deps := newPhotoTestDeps()
			deps.setupEditableSessionWithConsent(1, tc.consent)

			deps.fileStore.SaveFn = func(_ context.Context, _ string, _ io.Reader) (int64, error) {
				return 15, nil
			}

			photo := &domain.Photo{
				SessionID:    1,
				PhotoType:    domain.PhotoTypeBefore,
				OriginalName: "test.jpg",
				ContentType:  "image/jpeg",
				CreatedBy:    5,
				UpdatedBy:    5,
			}

			err := deps.svc.UploadPhoto(context.Background(), photo, strings.NewReader("fake image data"))

			if tc.wantErr != nil {
				assert.True(t, errors.Is(err, tc.wantErr), "expected %v, got %v", tc.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// UploadPhoto tests - Label photos
// ---------------------------------------------------------------------------

func TestUploadLabelPhoto_Success(t *testing.T) {
	deps := newPhotoTestDeps()
	consent := domain.PhotoConsentYes
	deps.setupEditableSessionWithConsent(1, &consent)

	deps.fileStore.SaveFn = func(_ context.Context, _ string, _ io.Reader) (int64, error) {
		return 20, nil
	}

	var capturedPhoto *domain.Photo
	deps.photoRepo.CreateFn = func(_ context.Context, photo *domain.Photo) error {
		capturedPhoto = photo
		photo.ID = 200
		return nil
	}

	moduleID := int64(10)
	photo := &domain.Photo{
		SessionID:    1,
		ModuleID:     &moduleID,
		PhotoType:    domain.PhotoTypeLabel,
		OriginalName: "label.png",
		ContentType:  "image/png",
		CreatedBy:    5,
		UpdatedBy:    5,
	}

	err := deps.svc.UploadPhoto(context.Background(), photo, strings.NewReader("fake label image data"))

	require.NoError(t, err)
	require.NotNil(t, capturedPhoto)
	assert.Equal(t, int64(200), photo.ID)
	assert.Contains(t, capturedPhoto.FilePath, "/label/")
	assert.True(t, photoPathRegex.MatchString(capturedPhoto.FilePath),
		"FilePath %q should match pattern", capturedPhoto.FilePath)
}

func TestUploadLabelPhoto_NoModuleID(t *testing.T) {
	deps := newPhotoTestDeps()
	consent := domain.PhotoConsentYes
	deps.setupEditableSessionWithConsent(1, &consent)

	photo := &domain.Photo{
		SessionID:    1,
		ModuleID:     nil,
		PhotoType:    domain.PhotoTypeLabel,
		OriginalName: "label.jpg",
		ContentType:  "image/jpeg",
		CreatedBy:    5,
		UpdatedBy:    5,
	}

	err := deps.svc.UploadPhoto(context.Background(), photo, strings.NewReader("fake data"))

	assert.True(t, errors.Is(err, service.ErrPhotoModuleRequired))
}

// ---------------------------------------------------------------------------
// UploadPhoto tests - Validation errors
// ---------------------------------------------------------------------------

func TestUploadPhoto_InvalidType(t *testing.T) {
	deps := newPhotoTestDeps()
	consent := domain.PhotoConsentYes
	deps.setupEditableSessionWithConsent(1, &consent)

	photo := &domain.Photo{
		SessionID:    1,
		PhotoType:    "unknown",
		OriginalName: "test.jpg",
		ContentType:  "image/jpeg",
		CreatedBy:    5,
		UpdatedBy:    5,
	}

	err := deps.svc.UploadPhoto(context.Background(), photo, strings.NewReader("fake data"))

	assert.True(t, errors.Is(err, service.ErrPhotoInvalidType))
}

func TestUploadPhoto_InvalidContentType(t *testing.T) {
	deps := newPhotoTestDeps()
	consent := domain.PhotoConsentYes
	deps.setupEditableSessionWithConsent(1, &consent)

	photo := &domain.Photo{
		SessionID:    1,
		PhotoType:    domain.PhotoTypeBefore,
		OriginalName: "test.txt",
		ContentType:  "text/plain",
		CreatedBy:    5,
		UpdatedBy:    5,
	}

	err := deps.svc.UploadPhoto(context.Background(), photo, strings.NewReader("fake data"))

	assert.True(t, errors.Is(err, service.ErrPhotoInvalidContentType))
}

// ---------------------------------------------------------------------------
// UploadPhoto tests - Session state
// ---------------------------------------------------------------------------

func TestUploadPhoto_SessionNotEditable(t *testing.T) {
	deps := newPhotoTestDeps()
	consent := domain.PhotoConsentYes

	deps.sessionRepo.GetByIDFn = func(_ context.Context, id int64) (*domain.Session, error) {
		return &domain.Session{
			ID:           id,
			Status:       domain.SessionStatusSigned,
			PhotoConsent: &consent,
			Version:      1,
		}, nil
	}

	photo := &domain.Photo{
		SessionID:    1,
		PhotoType:    domain.PhotoTypeBefore,
		OriginalName: "test.jpg",
		ContentType:  "image/jpeg",
		CreatedBy:    5,
		UpdatedBy:    5,
	}

	err := deps.svc.UploadPhoto(context.Background(), photo, strings.NewReader("fake data"))

	assert.True(t, errors.Is(err, service.ErrPhotoSessionNotEditable))
}

func TestUploadPhoto_SessionNotFound(t *testing.T) {
	deps := newPhotoTestDeps()

	deps.sessionRepo.GetByIDFn = func(_ context.Context, _ int64) (*domain.Session, error) {
		return nil, service.ErrSessionNotFound
	}

	photo := &domain.Photo{
		SessionID:    999,
		PhotoType:    domain.PhotoTypeBefore,
		OriginalName: "test.jpg",
		ContentType:  "image/jpeg",
		CreatedBy:    5,
		UpdatedBy:    5,
	}

	err := deps.svc.UploadPhoto(context.Background(), photo, strings.NewReader("fake data"))

	assert.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrSessionNotFound))
}

// ---------------------------------------------------------------------------
// UploadPhoto tests - Cleanup on DB failure
// ---------------------------------------------------------------------------

func TestUploadPhoto_DBFailCleansUpFile(t *testing.T) {
	deps := newPhotoTestDeps()
	consent := domain.PhotoConsentYes
	deps.setupEditableSessionWithConsent(1, &consent)

	var savedPath string
	deps.fileStore.SaveFn = func(_ context.Context, relPath string, _ io.Reader) (int64, error) {
		savedPath = relPath
		return 15, nil
	}

	dbErr := errors.New("db write failed")
	deps.photoRepo.CreateFn = func(_ context.Context, _ *domain.Photo) error {
		return dbErr
	}

	var deletedPath string
	deps.fileStore.DeleteFn = func(_ context.Context, relPath string) error {
		deletedPath = relPath
		return nil
	}

	photo := &domain.Photo{
		SessionID:    1,
		PhotoType:    domain.PhotoTypeBefore,
		OriginalName: "test.jpg",
		ContentType:  "image/jpeg",
		CreatedBy:    5,
		UpdatedBy:    5,
	}

	err := deps.svc.UploadPhoto(context.Background(), photo, strings.NewReader("fake data"))

	assert.Error(t, err)
	assert.NotEmpty(t, savedPath, "file should have been saved")
	assert.Equal(t, savedPath, deletedPath, "orphaned file should be cleaned up")
}

// ---------------------------------------------------------------------------
// UploadPhoto tests - Organized naming
// ---------------------------------------------------------------------------

func TestPhotoFilePath_OrganizedNaming(t *testing.T) {
	deps := newPhotoTestDeps()
	consent := domain.PhotoConsentYes
	deps.setupEditableSessionWithConsent(42, &consent)

	deps.fileStore.SaveFn = func(_ context.Context, _ string, _ io.Reader) (int64, error) {
		return 15, nil
	}

	var capturedPhoto *domain.Photo
	deps.photoRepo.CreateFn = func(_ context.Context, photo *domain.Photo) error {
		capturedPhoto = photo
		return nil
	}

	photo := &domain.Photo{
		SessionID:    42,
		PhotoType:    domain.PhotoTypeBefore,
		OriginalName: "test.jpg",
		ContentType:  "image/jpeg",
		CreatedBy:    5,
		UpdatedBy:    5,
	}

	err := deps.svc.UploadPhoto(context.Background(), photo, strings.NewReader("fake data"))

	require.NoError(t, err)
	require.NotNil(t, capturedPhoto)

	// Validate organized naming: sessions/{sessionID}/{type}/{32hex}.{ext}
	assert.Regexp(t, `^sessions/42/(before|label)/[0-9a-f]{32}\.jpg$`, capturedPhoto.FilePath)
}

// ---------------------------------------------------------------------------
// DeletePhoto tests
// ---------------------------------------------------------------------------

func TestDeletePhoto_Success(t *testing.T) {
	deps := newPhotoTestDeps()

	deps.photoRepo.GetByIDFn = func(_ context.Context, id int64) (*domain.Photo, error) {
		return &domain.Photo{
			ID:        id,
			SessionID: 1,
			FilePath:  "sessions/1/before/abc123.jpg",
		}, nil
	}

	consent := domain.PhotoConsentYes
	deps.sessionRepo.GetByIDFn = func(_ context.Context, id int64) (*domain.Session, error) {
		return &domain.Session{
			ID:           id,
			Status:       domain.SessionStatusInProgress,
			PhotoConsent: &consent,
			Version:      1,
		}, nil
	}

	fileDeleted := false
	deps.fileStore.DeleteFn = func(_ context.Context, relPath string) error {
		fileDeleted = true
		assert.Equal(t, "sessions/1/before/abc123.jpg", relPath)
		return nil
	}

	repoDeleted := false
	deps.photoRepo.DeleteFn = func(_ context.Context, id int64) error {
		repoDeleted = true
		assert.Equal(t, int64(10), id)
		return nil
	}

	err := deps.svc.DeletePhoto(context.Background(), 10, 5)

	require.NoError(t, err)
	assert.True(t, fileDeleted, "file should be deleted")
	assert.True(t, repoDeleted, "repo record should be deleted")
}

func TestDeletePhoto_SessionNotEditable(t *testing.T) {
	deps := newPhotoTestDeps()

	deps.photoRepo.GetByIDFn = func(_ context.Context, id int64) (*domain.Photo, error) {
		return &domain.Photo{
			ID:        id,
			SessionID: 1,
			FilePath:  "sessions/1/before/abc123.jpg",
		}, nil
	}

	deps.sessionRepo.GetByIDFn = func(_ context.Context, id int64) (*domain.Session, error) {
		return &domain.Session{
			ID:      id,
			Status:  domain.SessionStatusSigned,
			Version: 2,
		}, nil
	}

	err := deps.svc.DeletePhoto(context.Background(), 10, 5)

	assert.True(t, errors.Is(err, service.ErrPhotoSessionNotEditable))
}

// ---------------------------------------------------------------------------
// ListBySession tests
// ---------------------------------------------------------------------------

func TestListBySession_Success(t *testing.T) {
	deps := newPhotoTestDeps()

	expected := []*domain.Photo{
		{ID: 1, SessionID: 10, PhotoType: domain.PhotoTypeBefore},
		{ID: 2, SessionID: 10, PhotoType: domain.PhotoTypeLabel},
	}

	deps.photoRepo.ListBySessionFn = func(_ context.Context, sessionID int64) ([]*domain.Photo, error) {
		assert.Equal(t, int64(10), sessionID)
		return expected, nil
	}

	result, err := deps.svc.ListBySession(context.Background(), 10)

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, expected, result)
}

// ---------------------------------------------------------------------------
// ListByModule tests
// ---------------------------------------------------------------------------

func TestListByModule_Success(t *testing.T) {
	deps := newPhotoTestDeps()

	expected := []*domain.Photo{
		{ID: 3, SessionID: 10, PhotoType: domain.PhotoTypeLabel},
	}

	deps.photoRepo.ListByModuleFn = func(_ context.Context, moduleID int64) ([]*domain.Photo, error) {
		assert.Equal(t, int64(20), moduleID)
		return expected, nil
	}

	result, err := deps.svc.ListByModule(context.Background(), 20)

	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, expected, result)
}

// ---------------------------------------------------------------------------
// GetByID tests
// ---------------------------------------------------------------------------

func TestPhotoGetByID_Success(t *testing.T) {
	deps := newPhotoTestDeps()

	expected := &domain.Photo{
		ID:        1,
		SessionID: 10,
		PhotoType: domain.PhotoTypeBefore,
		FilePath:  "sessions/10/before/abc.jpg",
	}

	deps.photoRepo.GetByIDFn = func(_ context.Context, id int64) (*domain.Photo, error) {
		assert.Equal(t, int64(1), id)
		return expected, nil
	}

	result, err := deps.svc.GetByID(context.Background(), 1)

	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestPhotoGetByID_NotFound(t *testing.T) {
	deps := newPhotoTestDeps()

	deps.photoRepo.GetByIDFn = func(_ context.Context, _ int64) (*domain.Photo, error) {
		return nil, service.ErrPhotoNotFound
	}

	result, err := deps.svc.GetByID(context.Background(), 999)

	assert.Nil(t, result)
	assert.True(t, errors.Is(err, service.ErrPhotoNotFound))
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func photoStrPtr(s string) *string {
	return &s
}
