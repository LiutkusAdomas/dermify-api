package service

import (
	"context"
	"errors"
	"time"

	"dermify-api/internal/domain"
)

// Sentinel errors for contraindication screening operations.
var (
	ErrScreeningNotFound      = errors.New("screening not found")      //nolint:gochecknoglobals // sentinel error
	ErrScreeningAlreadyExists = errors.New("screening already exists") //nolint:gochecknoglobals // sentinel error
	ErrInvalidScreeningData   = errors.New("invalid screening data")   //nolint:gochecknoglobals // sentinel error
)

// ContraindicationRepository defines the data access contract for
// contraindication screening records.
type ContraindicationRepository interface {
	// Create inserts a new screening record and sets the ID on the provided struct.
	Create(ctx context.Context, screening *domain.ContraindicationScreening) error
	// GetBySessionID retrieves the screening record for a session.
	GetBySessionID(ctx context.Context, sessionID int64) (*domain.ContraindicationScreening, error)
	// Update modifies a screening record using optimistic locking.
	Update(ctx context.Context, screening *domain.ContraindicationScreening) error
}

// ContraindicationService handles contraindication screening business logic.
type ContraindicationService struct {
	repo ContraindicationRepository
}

// NewContraindicationService creates a new ContraindicationService with the
// given repository.
func NewContraindicationService(repo ContraindicationRepository) *ContraindicationService {
	return &ContraindicationService{repo: repo}
}

// RecordScreening validates and creates a screening record for a session.
func (s *ContraindicationService) RecordScreening(ctx context.Context, screening *domain.ContraindicationScreening) error {
	if screening.SessionID <= 0 {
		return ErrInvalidScreeningData
	}

	// Check if screening already exists for this session.
	existing, err := s.repo.GetBySessionID(ctx, screening.SessionID)
	if err != nil && !errors.Is(err, ErrScreeningNotFound) {
		return err
	}

	if existing != nil {
		return ErrScreeningAlreadyExists
	}

	computeHasFlags(screening)

	now := time.Now()
	screening.Version = 1
	screening.CreatedAt = now
	screening.UpdatedAt = now

	return s.repo.Create(ctx, screening)
}

// GetBySessionID retrieves the screening record for a session.
func (s *ContraindicationService) GetBySessionID(ctx context.Context, sessionID int64) (*domain.ContraindicationScreening, error) {
	return s.repo.GetBySessionID(ctx, sessionID)
}

// UpdateScreening validates and updates a screening record.
func (s *ContraindicationService) UpdateScreening(ctx context.Context, screening *domain.ContraindicationScreening) error {
	computeHasFlags(screening)
	screening.UpdatedAt = time.Now()

	return s.repo.Update(ctx, screening)
}

// computeHasFlags sets HasFlags to true if any contraindication boolean flag is true.
func computeHasFlags(screening *domain.ContraindicationScreening) {
	screening.HasFlags = screening.Pregnant ||
		screening.Breastfeeding ||
		screening.ActiveInfection ||
		screening.ActiveColdSores ||
		screening.Isotretinoin ||
		screening.Photosensitivity ||
		screening.AutoimmuneDisorder ||
		screening.KeloidHistory ||
		screening.Anticoagulants ||
		screening.RecentTan
}
