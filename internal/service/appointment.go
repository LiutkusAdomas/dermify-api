package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"dermify-api/internal/domain"
)

var (
	ErrAppointmentNotFound          = errors.New("appointment not found")          //nolint:gochecknoglobals // sentinel error
	ErrAppointmentVersionConflict   = errors.New("appointment version conflict")   //nolint:gochecknoglobals // sentinel error
	ErrInvalidAppointmentData       = errors.New("invalid appointment data")       //nolint:gochecknoglobals // sentinel error
	ErrAppointmentOverlap           = errors.New("appointment overlaps existing")  //nolint:gochecknoglobals // sentinel error
	ErrAppointmentInvalidTransition = errors.New("invalid status transition")      //nolint:gochecknoglobals // sentinel error
	ErrAppointmentOutsideHours      = errors.New("appointment outside work hours") //nolint:gochecknoglobals // sentinel error
)

// AppointmentFilter defines filtering and pagination options for appointment listing.
type AppointmentFilter struct {
	OrgID     int64
	DoctorID  int64
	PatientID int64
	Status    string
	Start     time.Time
	End       time.Time
	Page      int
	PerPage   int
}

// AppointmentListResult holds paginated appointment results.
type AppointmentListResult struct {
	Appointments []domain.Appointment `json:"appointments"`
	Total        int                  `json:"total"`
}

// AppointmentRepository defines the data access contract for appointments.
type AppointmentRepository interface {
	Create(ctx context.Context, a *domain.Appointment) error
	GetByID(ctx context.Context, id, orgID int64) (*domain.Appointment, error)
	Update(ctx context.Context, a *domain.Appointment) error
	UpdateStatus(ctx context.Context, id, orgID int64, status string, version int, cancellationReason string) error
	LinkSession(ctx context.Context, id, orgID int64, sessionID int64, version int) error
	List(ctx context.Context, filter AppointmentFilter) (*AppointmentListResult, error)
	HasOverlap(ctx context.Context, orgID, doctorID int64, start, end time.Time, excludeID int64) (bool, error)
	GetTimeSlotsForDate(ctx context.Context, orgID, doctorID int64, date time.Time) ([]domain.TimeSlot, error)
	GetBySessionID(ctx context.Context, orgID, sessionID int64) (*domain.Appointment, error)
}

// Valid appointment state transitions.
var validAppointmentTransitions = map[string][]string{ //nolint:gochecknoglobals // transition map
	domain.AppointmentStatusScheduled:  {domain.AppointmentStatusConfirmed, domain.AppointmentStatusCancelled, domain.AppointmentStatusNoShow},
	domain.AppointmentStatusConfirmed:  {domain.AppointmentStatusCheckedIn, domain.AppointmentStatusCancelled, domain.AppointmentStatusNoShow},
	domain.AppointmentStatusCheckedIn:  {domain.AppointmentStatusInProgress, domain.AppointmentStatusCancelled, domain.AppointmentStatusNoShow},
	domain.AppointmentStatusInProgress: {domain.AppointmentStatusCompleted},
}

// AppointmentService handles appointment business logic.
type AppointmentService struct {
	repo            AppointmentRepository
	scheduleRepo    ScheduleRepository
	sessionRepo     SessionRepository
	patientRepo     PatientRepository
	orgRepo         OrganizationRepository
	notificationSvc *NotificationService
}

// NewAppointmentService creates a new AppointmentService.
func NewAppointmentService(
	repo AppointmentRepository,
	scheduleRepo ScheduleRepository,
	sessionRepo SessionRepository,
	patientRepo PatientRepository,
	orgRepo OrganizationRepository,
	notificationSvc *NotificationService,
) *AppointmentService {
	return &AppointmentService{
		repo:            repo,
		scheduleRepo:    scheduleRepo,
		sessionRepo:     sessionRepo,
		patientRepo:     patientRepo,
		orgRepo:         orgRepo,
		notificationSvc: notificationSvc,
	}
}

// Create validates and inserts a new appointment.
func (s *AppointmentService) Create(ctx context.Context, a *domain.Appointment) error {
	if err := validateAppointment(a); err != nil {
		return err
	}
	if err := s.validateOrgBoundEntities(ctx, a.OrgID, a.PatientID, a.DoctorID); err != nil {
		return err
	}

	overlap, err := s.repo.HasOverlap(ctx, a.OrgID, a.DoctorID, a.StartTime, a.EndTime, 0)
	if err != nil {
		return fmt.Errorf("checking overlap: %w", err)
	}
	if overlap {
		return ErrAppointmentOverlap
	}

	now := time.Now().UTC()
	a.Status = domain.AppointmentStatusScheduled
	a.Version = 1
	a.CreatedAt = now
	a.UpdatedAt = now

	if err := s.repo.Create(ctx, a); err != nil {
		return err
	}

	go s.sendNotification(ctx, a, domain.NotificationTypeAppointmentConfirmation)

	return nil
}

// GetByID retrieves a single appointment.
func (s *AppointmentService) GetByID(ctx context.Context, id, orgID int64) (*domain.Appointment, error) {
	return s.repo.GetByID(ctx, id, orgID)
}

// Update reschedules an appointment.
func (s *AppointmentService) Update(ctx context.Context, a *domain.Appointment) error {
	existing, err := s.repo.GetByID(ctx, a.ID, a.OrgID)
	if err != nil {
		return err
	}
	if err := s.validateOrgBoundEntities(ctx, a.OrgID, a.PatientID, a.DoctorID); err != nil {
		return err
	}

	if existing.Status == domain.AppointmentStatusCancelled ||
		existing.Status == domain.AppointmentStatusCompleted ||
		existing.Status == domain.AppointmentStatusNoShow {
		return ErrAppointmentInvalidTransition
	}

	if !a.StartTime.Equal(existing.StartTime) || !a.EndTime.Equal(existing.EndTime) {
		overlap, err := s.repo.HasOverlap(ctx, a.OrgID, a.DoctorID, a.StartTime, a.EndTime, a.ID)
		if err != nil {
			return fmt.Errorf("checking overlap: %w", err)
		}
		if overlap {
			return ErrAppointmentOverlap
		}
	}

	a.Version = existing.Version
	a.Status = existing.Status
	a.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, a); err != nil {
		return err
	}

	if !a.StartTime.Equal(existing.StartTime) || !a.EndTime.Equal(existing.EndTime) {
		go s.sendNotification(ctx, a, domain.NotificationTypeAppointmentReschedule)
	}

	return nil
}

// UpdateStatus transitions an appointment's status.
func (s *AppointmentService) UpdateStatus(ctx context.Context, id, orgID int64, newStatus, cancellationReason string) error {
	existing, err := s.repo.GetByID(ctx, id, orgID)
	if err != nil {
		return err
	}

	if !isValidAppointmentTransition(existing.Status, newStatus) {
		return ErrAppointmentInvalidTransition
	}

	if err := s.repo.UpdateStatus(ctx, id, orgID, newStatus, existing.Version, cancellationReason); err != nil {
		return err
	}

	if newStatus == domain.AppointmentStatusCancelled {
		existing.Status = newStatus
		go s.sendNotification(ctx, existing, domain.NotificationTypeAppointmentCancellation)
	}

	return nil
}

// StartSession creates a session from an appointment and links them.
func (s *AppointmentService) StartSession(ctx context.Context, appointmentID, orgID, userID int64) (*domain.Session, error) {
	appt, err := s.repo.GetByID(ctx, appointmentID, orgID)
	if err != nil {
		return nil, err
	}
	if err := s.validateDoctorRole(ctx, orgID, appt.DoctorID); err != nil {
		return nil, err
	}

	if appt.Status != domain.AppointmentStatusCheckedIn && appt.Status != domain.AppointmentStatusConfirmed && appt.Status != domain.AppointmentStatusScheduled {
		return nil, ErrAppointmentInvalidTransition
	}

	session := &domain.Session{
		PatientID:   appt.PatientID,
		ClinicianID: appt.DoctorID,
	}

	now := time.Now().UTC()
	session.Status = domain.SessionStatusDraft
	session.Version = 1
	session.CreatedAt = now
	session.UpdatedAt = now
	session.CreatedBy = userID
	session.UpdatedBy = userID
	session.StartedAt = &now

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("creating session: %w", err)
	}

	if err := s.repo.LinkSession(ctx, appointmentID, orgID, session.ID, appt.Version); err != nil {
		return nil, fmt.Errorf("linking session: %w", err)
	}

	return session, nil
}

// List returns appointments matching the given filter.
func (s *AppointmentService) List(ctx context.Context, filter AppointmentFilter) (*AppointmentListResult, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PerPage <= 0 {
		filter.PerPage = 100
	}
	if filter.PerPage > 500 {
		filter.PerPage = 500
	}
	return s.repo.List(ctx, filter)
}

// GetTimeSlotsForDate returns existing appointment time slots for a doctor on a date.
func (s *AppointmentService) GetTimeSlotsForDate(ctx context.Context, orgID, doctorID int64, date time.Time) ([]domain.TimeSlot, error) {
	return s.repo.GetTimeSlotsForDate(ctx, orgID, doctorID, date)
}

func (s *AppointmentService) sendNotification(ctx context.Context, a *domain.Appointment, notifType string) {
	if s.notificationSvc == nil {
		return
	}

	patient, err := s.patientRepo.GetByID(ctx, a.PatientID)
	if err != nil || patient == nil {
		return
	}

	recipient := ""
	if patient.Email != nil {
		recipient = *patient.Email
	} else if patient.Phone != nil {
		recipient = *patient.Phone
	}
	if recipient == "" {
		return
	}

	subject := "Appointment Update"
	body := fmt.Sprintf("Your appointment on %s has been updated. Status: %s",
		a.StartTime.Format("2006-01-02 15:04"), notifType)

	s.notificationSvc.SendAppointmentNotification(ctx, a.OrgID, a.PatientID, a.ID, notifType, recipient, subject, body)
}

func validateAppointment(a *domain.Appointment) error {
	if a.OrgID <= 0 || a.PatientID <= 0 || a.DoctorID <= 0 || a.ServiceTypeID <= 0 {
		return ErrInvalidAppointmentData
	}
	if a.StartTime.IsZero() || a.EndTime.IsZero() {
		return ErrInvalidAppointmentData
	}
	if !a.EndTime.After(a.StartTime) {
		return ErrInvalidAppointmentData
	}
	return nil
}

func (s *AppointmentService) validateOrgBoundEntities(ctx context.Context, orgID, patientID, doctorID int64) error {
	if err := s.validateDoctorRole(ctx, orgID, doctorID); err != nil {
		return err
	}

	patient, err := s.patientRepo.GetByID(ctx, patientID)
	if err != nil {
		return err
	}

	if _, err := s.orgRepo.GetMemberRole(ctx, orgID, patient.CreatedBy); err != nil {
		return ErrInvalidAppointmentData
	}

	return nil
}

func (s *AppointmentService) validateDoctorRole(ctx context.Context, orgID, doctorID int64) error {
	role, err := s.orgRepo.GetMemberRole(ctx, orgID, doctorID)
	if err != nil {
		return ErrInvalidAppointmentData
	}
	if role != domain.OrgRoleDoctor && role != domain.OrgRoleAdmin {
		return ErrInvalidAppointmentData
	}
	return nil
}

func isValidAppointmentTransition(current, next string) bool {
	allowed, ok := validAppointmentTransitions[current]
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
