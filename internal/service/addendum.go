package service

import (
	"context"
	"errors"
	"time"

	"dermify-api/internal/domain"
)

// Sentinel errors for addendum operations.
var (
	ErrAddendumNotFound    = errors.New("addendum not found")                     //nolint:gochecknoglobals // sentinel error
	ErrInvalidAddendumData = errors.New("invalid addendum data")                  //nolint:gochecknoglobals // sentinel error
	ErrSessionNotLocked    = errors.New("addendums only allowed on locked sessions") //nolint:gochecknoglobals // sentinel error
)

// AddendumRepository defines the data access contract for addendum records.
type AddendumRepository interface {
	// Create inserts a new addendum record and sets the ID on the provided struct.
	Create(ctx context.Context, addendum *domain.Addendum) error
	// GetByID retrieves an addendum by its ID.
	GetByID(ctx context.Context, id int64) (*domain.Addendum, error)
	// ListBySession returns all addendums for a session ordered by created_at.
	ListBySession(ctx context.Context, sessionID int64) ([]domain.Addendum, error)
}

// AddendumService handles addendum business logic.
type AddendumService struct {
	repo        AddendumRepository
	sessionRepo SessionRepository
}

// NewAddendumService creates a new AddendumService with the given dependencies.
func NewAddendumService(repo AddendumRepository, sessionRepo SessionRepository) *AddendumService {
	return &AddendumService{
		repo:        repo,
		sessionRepo: sessionRepo,
	}
}

// CreateAddendum validates and creates an addendum on a locked session.
func (s *AddendumService) CreateAddendum(ctx context.Context, addendum *domain.Addendum) error {
	if addendum.SessionID <= 0 {
		return ErrInvalidAddendumData
	}

	if addendum.AuthorID <= 0 {
		return ErrInvalidAddendumData
	}

	if addendum.Reason == "" {
		return ErrInvalidAddendumData
	}

	if addendum.Content == "" {
		return ErrInvalidAddendumData
	}

	session, err := s.sessionRepo.GetByID(ctx, addendum.SessionID)
	if err != nil {
		return err
	}

	if session.Status != domain.SessionStatusLocked {
		return ErrSessionNotLocked
	}

	addendum.CreatedAt = time.Now()

	return s.repo.Create(ctx, addendum)
}

// GetByID retrieves an addendum by its ID.
func (s *AddendumService) GetByID(ctx context.Context, id int64) (*domain.Addendum, error) {
	return s.repo.GetByID(ctx, id)
}

// ListBySession returns all addendums for a session.
func (s *AddendumService) ListBySession(ctx context.Context, sessionID int64) ([]domain.Addendum, error) {
	return s.repo.ListBySession(ctx, sessionID)
}
