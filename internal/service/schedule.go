package service

import (
	"context"
	"errors"
	"time"

	"dermify-api/internal/domain"
)

var (
	ErrScheduleInvalidData    = errors.New("invalid schedule data")      //nolint:gochecknoglobals // sentinel error
	ErrScheduleOverrideNotFound = errors.New("schedule override not found") //nolint:gochecknoglobals // sentinel error
)

// ScheduleRepository defines the data access contract for doctor schedules.
type ScheduleRepository interface {
	UpsertWorkingHours(ctx context.Context, orgID, doctorID int64, hours []domain.WorkingHours) error
	GetWorkingHours(ctx context.Context, orgID, doctorID int64) ([]domain.WorkingHours, error)
	GetWorkingHoursForDay(ctx context.Context, orgID, doctorID int64, dayOfWeek int) (*domain.WorkingHours, error)
	CreateOverride(ctx context.Context, o *domain.ScheduleOverride) error
	DeleteOverride(ctx context.Context, id, orgID int64) error
	ListOverrides(ctx context.Context, orgID, doctorID int64, from, to time.Time) ([]*domain.ScheduleOverride, error)
	GetOverrideForDate(ctx context.Context, orgID, doctorID int64, date time.Time) (*domain.ScheduleOverride, error)
}

// ScheduleService handles doctor schedule business logic.
type ScheduleService struct {
	repo ScheduleRepository
}

// NewScheduleService creates a new ScheduleService.
func NewScheduleService(repo ScheduleRepository) *ScheduleService {
	return &ScheduleService{repo: repo}
}

// SetWorkingHours replaces a doctor's weekly working hours.
func (s *ScheduleService) SetWorkingHours(ctx context.Context, orgID, doctorID int64, hours []domain.WorkingHours) error {
	for i := range hours {
		hours[i].OrgID = orgID
		hours[i].DoctorID = doctorID
		if hours[i].DayOfWeek < 0 || hours[i].DayOfWeek > 6 {
			return ErrScheduleInvalidData
		}
		if hours[i].StartTime == "" || hours[i].EndTime == "" {
			return ErrScheduleInvalidData
		}
		if hours[i].StartTime >= hours[i].EndTime {
			return ErrScheduleInvalidData
		}
	}
	return s.repo.UpsertWorkingHours(ctx, orgID, doctorID, hours)
}

// GetWorkingHours returns a doctor's weekly schedule.
func (s *ScheduleService) GetWorkingHours(ctx context.Context, orgID, doctorID int64) ([]domain.WorkingHours, error) {
	return s.repo.GetWorkingHours(ctx, orgID, doctorID)
}

// CreateOverride adds a schedule override (day off or modified hours).
func (s *ScheduleService) CreateOverride(ctx context.Context, o *domain.ScheduleOverride) error {
	if o.OrgID <= 0 || o.DoctorID <= 0 {
		return ErrScheduleInvalidData
	}
	if o.Date.IsZero() {
		return ErrScheduleInvalidData
	}
	if o.StartTime != nil && o.EndTime != nil && *o.StartTime >= *o.EndTime {
		return ErrScheduleInvalidData
	}
	o.CreatedAt = time.Now().UTC()
	return s.repo.CreateOverride(ctx, o)
}

// DeleteOverride removes a schedule override.
func (s *ScheduleService) DeleteOverride(ctx context.Context, id, orgID int64) error {
	return s.repo.DeleteOverride(ctx, id, orgID)
}

// ListOverrides returns overrides for a doctor within a date range.
func (s *ScheduleService) ListOverrides(ctx context.Context, orgID, doctorID int64, from, to time.Time) ([]*domain.ScheduleOverride, error) {
	return s.repo.ListOverrides(ctx, orgID, doctorID, from, to)
}

// GetAvailableSlots computes free time slots for a doctor on a given date.
func (s *ScheduleService) GetAvailableSlots(ctx context.Context, orgID, doctorID int64, date time.Time, durationMinutes int, existingAppointments []domain.TimeSlot) ([]domain.TimeSlot, error) {
	if durationMinutes <= 0 {
		durationMinutes = 30
	}

	dayOfWeek := int(date.Weekday())

	override, err := s.repo.GetOverrideForDate(ctx, orgID, doctorID, date)
	if err != nil && !errors.Is(err, ErrScheduleOverrideNotFound) {
		return nil, err
	}

	var dayStart, dayEnd string
	if override != nil {
		if override.StartTime == nil {
			return []domain.TimeSlot{}, nil
		}
		dayStart = *override.StartTime
		dayEnd = *override.EndTime
	} else {
		wh, err := s.repo.GetWorkingHoursForDay(ctx, orgID, doctorID, dayOfWeek)
		if err != nil {
			return []domain.TimeSlot{}, nil //nolint:nilerr // no working hours = no slots
		}
		dayStart = wh.StartTime
		dayEnd = wh.EndTime
	}

	startTime, err := parseTimeOnDate(date, dayStart)
	if err != nil {
		return nil, err
	}
	endTime, err := parseTimeOnDate(date, dayEnd)
	if err != nil {
		return nil, err
	}

	duration := time.Duration(durationMinutes) * time.Minute
	interval := 15 * time.Minute

	var slots []domain.TimeSlot
	for current := startTime; current.Add(duration).Before(endTime) || current.Add(duration).Equal(endTime); current = current.Add(interval) {
		slotEnd := current.Add(duration)
		if !overlapsAny(current, slotEnd, existingAppointments) {
			slots = append(slots, domain.TimeSlot{Start: current, End: slotEnd})
		}
	}

	return slots, nil
}

func parseTimeOnDate(date time.Time, timeStr string) (time.Time, error) {
	t, err := time.Parse("15:04", timeStr)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(date.Year(), date.Month(), date.Day(), t.Hour(), t.Minute(), 0, 0, date.Location()), nil
}

func overlapsAny(start, end time.Time, appointments []domain.TimeSlot) bool {
	for _, a := range appointments {
		if start.Before(a.End) && end.After(a.Start) {
			return true
		}
	}
	return false
}
