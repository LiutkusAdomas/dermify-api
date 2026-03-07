package service

import (
	"context"
	"errors"

	"dermify-api/internal/domain"
)

// Sentinel errors for contraindication screening operations.
var (
	ErrScreeningNotFound      = errors.New("screening not found")       //nolint:gochecknoglobals // sentinel error
	ErrScreeningAlreadyExists = errors.New("screening already exists")  //nolint:gochecknoglobals // sentinel error
	ErrInvalidScreeningData   = errors.New("invalid screening data")    //nolint:gochecknoglobals // sentinel error
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
// TODO: implement in Plan 02.
func (s *ContraindicationService) RecordScreening(_ context.Context, _ *domain.ContraindicationScreening) error {
	return nil
}

// GetBySessionID retrieves the screening record for a session.
// TODO: implement in Plan 02.
func (s *ContraindicationService) GetBySessionID(_ context.Context, _ int64) (*domain.ContraindicationScreening, error) {
	return nil, nil
}

// UpdateScreening validates and updates a screening record.
// TODO: implement in Plan 02.
func (s *ContraindicationService) UpdateScreening(_ context.Context, _ *domain.ContraindicationScreening) error {
	return nil
}
