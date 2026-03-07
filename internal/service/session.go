package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"dermify-api/internal/domain"
)

// Sentinel errors for session operations.
var (
	ErrSessionNotFound        = errors.New("session not found")          //nolint:gochecknoglobals // sentinel error
	ErrSessionVersionConflict = errors.New("session version conflict")   //nolint:gochecknoglobals // sentinel error
	ErrInvalidSessionData     = errors.New("invalid session data")       //nolint:gochecknoglobals // sentinel error
	ErrInvalidStateTransition = errors.New("invalid state transition")   //nolint:gochecknoglobals // sentinel error
	ErrSessionNotEditable     = errors.New("session is not editable")    //nolint:gochecknoglobals // sentinel error
	ErrConsentRequired        = errors.New("consent required for session") //nolint:gochecknoglobals // sentinel error
)

// SessionFilter defines filtering and pagination options for session listing.
type SessionFilter struct {
	PatientID   int64
	ClinicianID int64
	Status      string
	Page        int
	PerPage     int
}

// SessionListResult holds paginated session results.
type SessionListResult struct {
	Sessions []domain.Session
	Total    int
}

// SessionRepository defines the data access contract for sessions.
type SessionRepository interface {
	// Create inserts a new session and sets the ID on the provided struct.
	Create(ctx context.Context, session *domain.Session) error
	// GetByID retrieves a session by ID.
	GetByID(ctx context.Context, id int64) (*domain.Session, error)
	// Update modifies a session using optimistic locking on the version field.
	Update(ctx context.Context, session *domain.Session) error
	// UpdateStatus transitions a session to a new status with optimistic locking.
	UpdateStatus(ctx context.Context, id int64, status string, expectedVersion int, userID int64) error
	// List returns paginated sessions matching the given filter.
	List(ctx context.Context, filter SessionFilter) (*SessionListResult, error)
	// ListByPatient returns session summaries for a patient.
	ListByPatient(ctx context.Context, patientID int64) ([]domain.SessionSummary, error)
	// SetIndicationCodes replaces the indication code associations for a session.
	SetIndicationCodes(ctx context.Context, sessionID int64, codeIDs []int64) error
}

// ModuleRepository defines the data access contract for session modules.
type ModuleRepository interface {
	// Create inserts a new session module.
	Create(ctx context.Context, module *domain.SessionModule, userID int64) error
	// ListBySession returns all modules for a session ordered by sort_order.
	ListBySession(ctx context.Context, sessionID int64) ([]domain.SessionModule, error)
	// Delete removes a module from a session.
	Delete(ctx context.Context, id int64, sessionID int64) error
	// NextSortOrder returns the next available sort order for a session.
	NextSortOrder(ctx context.Context, sessionID int64) (int, error)
}

// Valid session state transitions.
// Key = current state, Value = set of allowed next states.
var validTransitions = map[string][]string{ //nolint:gochecknoglobals // transition map
	domain.SessionStatusDraft:           {domain.SessionStatusInProgress},
	domain.SessionStatusInProgress:      {domain.SessionStatusAwaitingSignoff},
	domain.SessionStatusAwaitingSignoff: {domain.SessionStatusSigned, domain.SessionStatusInProgress},
	domain.SessionStatusSigned:          {domain.SessionStatusLocked},
	// Locked is terminal -- no transitions out.
}

// SessionService handles session business logic.
type SessionService struct {
	repo        SessionRepository
	consentRepo ConsentRepository
	moduleRepo  ModuleRepository
}

// NewSessionService creates a new SessionService with the given repositories.
func NewSessionService(repo SessionRepository, consentRepo ConsentRepository, moduleRepo ModuleRepository) *SessionService {
	return &SessionService{
		repo:        repo,
		consentRepo: consentRepo,
		moduleRepo:  moduleRepo,
	}
}

// Create validates and creates a new session record.
// It sets defaults (status=draft, version=1, timestamps) and persists via the repository.
// If IndicationCodes are provided, they are stored after the session is created.
func (s *SessionService) Create(ctx context.Context, session *domain.Session) error {
	if err := validateSession(session); err != nil {
		return err
	}

	now := time.Now()
	session.Status = domain.SessionStatusDraft
	session.Version = 1
	session.CreatedAt = now
	session.UpdatedAt = now
	session.CreatedBy = session.ClinicianID
	session.UpdatedBy = session.ClinicianID

	if err := s.repo.Create(ctx, session); err != nil {
		return fmt.Errorf("creating session: %w", err)
	}

	if len(session.IndicationCodes) > 0 {
		if err := s.repo.SetIndicationCodes(ctx, session.ID, session.IndicationCodes); err != nil {
			return fmt.Errorf("setting indication codes: %w", err)
		}
	}

	return nil
}

// GetByID retrieves a session by ID.
func (s *SessionService) GetByID(ctx context.Context, id int64) (*domain.Session, error) {
	return s.repo.GetByID(ctx, id)
}

// Update validates and updates a session record with optimistic locking.
// Only sessions in draft or in_progress state are editable.
func (s *SessionService) Update(ctx context.Context, session *domain.Session) error {
	existing, err := s.repo.GetByID(ctx, session.ID)
	if err != nil {
		return err
	}

	if !isEditable(existing.Status) {
		return ErrSessionNotEditable
	}

	if err := validateSessionFields(session); err != nil {
		return err
	}

	session.UpdatedAt = time.Now()

	return s.repo.Update(ctx, session)
}

// TransitionState validates and applies a session state change.
// It fetches the current session, checks the transition is valid, and updates
// the status with optimistic locking.
func (s *SessionService) TransitionState(ctx context.Context, id int64, newStatus string, userID int64) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if !IsValidTransition(existing.Status, newStatus) {
		return ErrInvalidStateTransition
	}

	return s.repo.UpdateStatus(ctx, id, newStatus, existing.Version, userID)
}

// List returns paginated sessions matching the given filter with defaults applied.
func (s *SessionService) List(ctx context.Context, filter SessionFilter) (*SessionListResult, error) {
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

// ListByPatient returns session summaries for a patient.
func (s *SessionService) ListByPatient(ctx context.Context, patientID int64) ([]domain.SessionSummary, error) {
	return s.repo.ListByPatient(ctx, patientID)
}

// validModuleTypes contains the set of recognized procedure module types.
var validModuleTypes = map[string]bool{ //nolint:gochecknoglobals // module type set
	domain.ModuleTypeIPL:       true,
	domain.ModuleTypeNdYAG:     true,
	domain.ModuleTypeCO2:       true,
	domain.ModuleTypeRF:        true,
	domain.ModuleTypeFiller:    true,
	domain.ModuleTypeBotulinum: true,
}

// AddModule validates consent exists before inserting a procedure module slot.
// It checks session editability, validates the module type, enforces the consent
// gate, and assigns the next sort order before creating the module.
func (s *SessionService) AddModule(ctx context.Context, sessionID int64, moduleType string, userID int64) (*domain.SessionModule, error) {
	session, err := s.repo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if !isEditable(session.Status) {
		return nil, ErrSessionNotEditable
	}

	if !validModuleTypes[moduleType] {
		return nil, ErrInvalidSessionData
	}

	hasConsent, err := s.consentRepo.ExistsForSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if !hasConsent {
		return nil, ErrConsentRequired
	}

	sortOrder, err := s.moduleRepo.NextSortOrder(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	module := &domain.SessionModule{
		SessionID:  sessionID,
		ModuleType: moduleType,
		SortOrder:  sortOrder,
		Version:    1,
		CreatedBy:  userID,
		UpdatedBy:  userID,
	}

	if err := s.moduleRepo.Create(ctx, module, userID); err != nil {
		return nil, err
	}

	return module, nil
}

// ListModules returns all modules for a session ordered by sort order.
func (s *SessionService) ListModules(ctx context.Context, sessionID int64) ([]domain.SessionModule, error) {
	return s.moduleRepo.ListBySession(ctx, sessionID)
}

// RemoveModule removes a module from an editable session.
func (s *SessionService) RemoveModule(ctx context.Context, sessionID int64, moduleID int64, _ int64) error {
	session, err := s.repo.GetByID(ctx, sessionID)
	if err != nil {
		return err
	}

	if !isEditable(session.Status) {
		return ErrSessionNotEditable
	}

	return s.moduleRepo.Delete(ctx, moduleID, sessionID)
}

// isEditable returns true if the session status allows field updates.
func isEditable(status string) bool {
	return status == domain.SessionStatusDraft || status == domain.SessionStatusInProgress
}

// validateSession checks required fields for session creation.
func validateSession(session *domain.Session) error {
	if session.PatientID <= 0 {
		return ErrInvalidSessionData
	}

	if session.ClinicianID <= 0 {
		return ErrInvalidSessionData
	}

	return validateSessionFields(session)
}

// validateSessionFields checks optional fields that have constrained values.
func validateSessionFields(session *domain.Session) error {
	if session.FitzpatrickType != nil {
		fitz := *session.FitzpatrickType
		if fitz < domain.FitzpatrickMin || fitz > domain.FitzpatrickMax {
			return ErrInvalidSessionData
		}
	}

	if session.PhotoConsent != nil {
		consent := *session.PhotoConsent
		if consent != domain.PhotoConsentYes &&
			consent != domain.PhotoConsentNo &&
			consent != domain.PhotoConsentLimited {
			return ErrInvalidSessionData
		}
	}

	return nil
}

// IsValidTransition checks whether a transition from the current state to the
// new state is allowed. Exported for use in tests and documentation.
func IsValidTransition(current, next string) bool {
	allowed, ok := validTransitions[current]
	if !ok {
		return false
	}

	for _, a := range allowed {
		if a == next {
			return true
		}
	}

	return false
}
