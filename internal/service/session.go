package service

import (
	"context"
	"errors"

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
// TODO: implement in Plan 02.
func (s *SessionService) Create(_ context.Context, _ *domain.Session) error {
	return nil
}

// GetByID retrieves a session by ID.
// TODO: implement in Plan 02.
func (s *SessionService) GetByID(_ context.Context, _ int64) (*domain.Session, error) {
	return nil, nil
}

// Update validates and updates a session record with optimistic locking.
// TODO: implement in Plan 02.
func (s *SessionService) Update(_ context.Context, _ *domain.Session) error {
	return nil
}

// TransitionState validates and applies a session state change.
// TODO: implement in Plan 02.
func (s *SessionService) TransitionState(_ context.Context, _ int64, _ string, _ int64) error {
	return nil
}

// List returns paginated sessions matching the given filter.
// TODO: implement in Plan 02.
func (s *SessionService) List(_ context.Context, _ SessionFilter) (*SessionListResult, error) {
	return nil, nil
}

// ListByPatient returns session summaries for a patient.
// TODO: implement in Plan 02.
func (s *SessionService) ListByPatient(_ context.Context, _ int64) ([]domain.SessionSummary, error) {
	return nil, nil
}

// AddModule validates consent exists before inserting a procedure module slot.
// TODO: implement in Plan 02.
func (s *SessionService) AddModule(_ context.Context, _ int64, _ string, _ int64) (*domain.SessionModule, error) {
	return nil, nil
}

// ListModules returns all modules for a session.
// TODO: implement in Plan 02.
func (s *SessionService) ListModules(_ context.Context, _ int64) ([]domain.SessionModule, error) {
	return nil, nil
}

// RemoveModule removes a module from a session.
// TODO: implement in Plan 02.
func (s *SessionService) RemoveModule(_ context.Context, _ int64, _ int64, _ int64) error {
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
