package service

import (
	"context"
	"errors"
	"time"

	"dermify-api/internal/domain"
)

// Maximum allowed per-page value for patient listing.
const maxPerPage = 100

// Default pagination values for patient listing.
const (
	defaultPage    = 1
	defaultPerPage = 20
)

// Sentinel errors for patient operations.
var (
	ErrPatientNotFound        = errors.New("patient not found")         //nolint:gochecknoglobals // sentinel error
	ErrPatientVersionConflict = errors.New("patient version conflict")  //nolint:gochecknoglobals // sentinel error
	ErrInvalidPatientData     = errors.New("invalid patient data")      //nolint:gochecknoglobals // sentinel error
)

// PatientFilter defines filtering and pagination options for patient listing.
type PatientFilter struct {
	Search  string
	Page    int
	PerPage int
}

// PatientListItem extends domain.Patient with session metadata placeholders.
type PatientListItem struct {
	domain.Patient
	SessionCount    int        `json:"session_count"`
	LastSessionDate *time.Time `json:"last_session_date"`
}

// PatientListResult holds paginated patient results.
type PatientListResult struct {
	Patients []PatientListItem
	Total    int
}

// PatientRepository defines the data access contract for patients.
type PatientRepository interface {
	// Create inserts a new patient and sets the ID on the provided struct.
	Create(ctx context.Context, patient *domain.Patient) error
	// GetByID retrieves a patient by ID.
	GetByID(ctx context.Context, id int64) (*domain.Patient, error)
	// Update modifies a patient using optimistic locking on the version field.
	Update(ctx context.Context, patient *domain.Patient) error
	// List returns paginated patients matching the given filter.
	List(ctx context.Context, filter PatientFilter) (*PatientListResult, error)
	// GetSessionHistory returns session summaries for a patient.
	GetSessionHistory(ctx context.Context, patientID int64) ([]domain.SessionSummary, error)
}

// PatientService handles patient business logic.
type PatientService struct {
	repo PatientRepository
}

// NewPatientService creates a new PatientService with the given repository.
func NewPatientService(repo PatientRepository) *PatientService {
	return &PatientService{repo: repo}
}

// Create validates and creates a new patient record.
func (s *PatientService) Create(ctx context.Context, patient *domain.Patient) error {
	if err := validatePatient(patient); err != nil {
		return err
	}

	now := time.Now()
	patient.Version = 1
	patient.CreatedAt = now
	patient.UpdatedAt = now

	return s.repo.Create(ctx, patient)
}

// GetByID retrieves a patient by ID.
func (s *PatientService) GetByID(ctx context.Context, id int64) (*domain.Patient, error) {
	return s.repo.GetByID(ctx, id)
}

// Update validates and updates a patient record with optimistic locking.
func (s *PatientService) Update(ctx context.Context, patient *domain.Patient) error {
	if err := validatePatient(patient); err != nil {
		return err
	}

	patient.UpdatedAt = time.Now()

	return s.repo.Update(ctx, patient)
}

// List returns paginated patients matching the given filter with defaults applied.
func (s *PatientService) List(ctx context.Context, filter PatientFilter) (*PatientListResult, error) {
	if filter.Page <= 0 {
		filter.Page = defaultPage
	}

	if filter.PerPage <= 0 {
		filter.PerPage = defaultPerPage
	}

	if filter.PerPage > maxPerPage {
		filter.PerPage = maxPerPage
	}

	return s.repo.List(ctx, filter)
}

// GetSessionHistory returns session summaries for a patient.
func (s *PatientService) GetSessionHistory(ctx context.Context, patientID int64) ([]domain.SessionSummary, error) {
	return s.repo.GetSessionHistory(ctx, patientID)
}

// validatePatient checks required fields on a patient record.
func validatePatient(patient *domain.Patient) error {
	if patient.FirstName == "" {
		return ErrInvalidPatientData
	}

	if patient.LastName == "" {
		return ErrInvalidPatientData
	}

	if patient.DateOfBirth.IsZero() {
		return ErrInvalidPatientData
	}

	if patient.Sex == "" {
		return ErrInvalidPatientData
	}

	return nil
}
