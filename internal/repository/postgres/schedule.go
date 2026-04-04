package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"dermify-api/internal/domain"
	"dermify-api/internal/service"
)

// PostgresScheduleRepository implements service.ScheduleRepository.
type PostgresScheduleRepository struct {
	db *sql.DB
}

// NewPostgresScheduleRepository creates a new PostgresScheduleRepository.
func NewPostgresScheduleRepository(db *sql.DB) *PostgresScheduleRepository {
	return &PostgresScheduleRepository{db: db}
}

// UpsertWorkingHours replaces a doctor's weekly working hours in a transaction.
func (r *PostgresScheduleRepository) UpsertWorkingHours(ctx context.Context, orgID, doctorID int64, hours []domain.WorkingHours) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // best-effort rollback

	_, err = tx.ExecContext(ctx,
		`DELETE FROM working_hours WHERE org_id = $1 AND doctor_id = $2`,
		orgID, doctorID,
	)
	if err != nil {
		return fmt.Errorf("clearing working hours: %w", err)
	}

	for _, h := range hours {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO working_hours (org_id, doctor_id, day_of_week, start_time, end_time)
			 VALUES ($1, $2, $3, $4, $5)`,
			orgID, doctorID, h.DayOfWeek, h.StartTime, h.EndTime,
		)
		if err != nil {
			return fmt.Errorf("inserting working hours: %w", err)
		}
	}

	return tx.Commit()
}

// GetWorkingHours returns all working hours for a doctor.
func (r *PostgresScheduleRepository) GetWorkingHours(ctx context.Context, orgID, doctorID int64) ([]domain.WorkingHours, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, org_id, doctor_id, day_of_week, start_time::text, end_time::text
		 FROM working_hours WHERE org_id = $1 AND doctor_id = $2
		 ORDER BY day_of_week`,
		orgID, doctorID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying working hours: %w", err)
	}
	defer rows.Close()

	var hours []domain.WorkingHours
	for rows.Next() {
		var h domain.WorkingHours
		if err := rows.Scan(&h.ID, &h.OrgID, &h.DoctorID, &h.DayOfWeek, &h.StartTime, &h.EndTime); err != nil {
			return nil, fmt.Errorf("scanning working hours: %w", err)
		}
		// Trim TIME format "HH:MM:SS" to "HH:MM"
		if len(h.StartTime) > 5 {
			h.StartTime = h.StartTime[:5]
		}
		if len(h.EndTime) > 5 {
			h.EndTime = h.EndTime[:5]
		}
		hours = append(hours, h)
	}
	return hours, rows.Err()
}

// GetWorkingHoursForDay returns working hours for a specific day of week.
func (r *PostgresScheduleRepository) GetWorkingHoursForDay(ctx context.Context, orgID, doctorID int64, dayOfWeek int) (*domain.WorkingHours, error) {
	var h domain.WorkingHours
	err := r.db.QueryRowContext(ctx,
		`SELECT id, org_id, doctor_id, day_of_week, start_time::text, end_time::text
		 FROM working_hours WHERE org_id = $1 AND doctor_id = $2 AND day_of_week = $3`,
		orgID, doctorID, dayOfWeek,
	).Scan(&h.ID, &h.OrgID, &h.DoctorID, &h.DayOfWeek, &h.StartTime, &h.EndTime)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrScheduleInvalidData
	}
	if err != nil {
		return nil, fmt.Errorf("querying working hours for day: %w", err)
	}
	if len(h.StartTime) > 5 {
		h.StartTime = h.StartTime[:5]
	}
	if len(h.EndTime) > 5 {
		h.EndTime = h.EndTime[:5]
	}
	return &h, nil
}

// CreateOverride inserts a schedule override.
func (r *PostgresScheduleRepository) CreateOverride(ctx context.Context, o *domain.ScheduleOverride) error {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO schedule_overrides (org_id, doctor_id, date, start_time, end_time, reason)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, created_at`,
		o.OrgID, o.DoctorID, o.Date, o.StartTime, o.EndTime, o.Reason,
	).Scan(&o.ID, &o.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return service.ErrScheduleInvalidData
		}
		return fmt.Errorf("inserting schedule override: %w", err)
	}
	return nil
}

// DeleteOverride removes a schedule override.
func (r *PostgresScheduleRepository) DeleteOverride(ctx context.Context, id, orgID int64) error {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM schedule_overrides WHERE id = $1 AND org_id = $2`, id, orgID,
	)
	if err != nil {
		return fmt.Errorf("deleting schedule override: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return service.ErrScheduleOverrideNotFound
	}
	return nil
}

// ListOverrides returns schedule overrides within a date range.
func (r *PostgresScheduleRepository) ListOverrides(ctx context.Context, orgID, doctorID int64, from, to time.Time) ([]*domain.ScheduleOverride, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, org_id, doctor_id, date, start_time::text, end_time::text, COALESCE(reason, ''), created_at
		 FROM schedule_overrides
		 WHERE org_id = $1 AND doctor_id = $2 AND date >= $3 AND date <= $4
		 ORDER BY date`,
		orgID, doctorID, from, to,
	)
	if err != nil {
		return nil, fmt.Errorf("querying schedule overrides: %w", err)
	}
	defer rows.Close()

	var overrides []*domain.ScheduleOverride
	for rows.Next() {
		var o domain.ScheduleOverride
		if err := rows.Scan(&o.ID, &o.OrgID, &o.DoctorID, &o.Date, &o.StartTime, &o.EndTime, &o.Reason, &o.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning schedule override: %w", err)
		}
		overrides = append(overrides, &o)
	}
	return overrides, rows.Err()
}

// GetOverrideForDate returns an override for a specific date if one exists.
func (r *PostgresScheduleRepository) GetOverrideForDate(ctx context.Context, orgID, doctorID int64, date time.Time) (*domain.ScheduleOverride, error) {
	var o domain.ScheduleOverride
	err := r.db.QueryRowContext(ctx,
		`SELECT id, org_id, doctor_id, date, start_time::text, end_time::text, COALESCE(reason, ''), created_at
		 FROM schedule_overrides WHERE org_id = $1 AND doctor_id = $2 AND date = $3`,
		orgID, doctorID, date,
	).Scan(&o.ID, &o.OrgID, &o.DoctorID, &o.Date, &o.StartTime, &o.EndTime, &o.Reason, &o.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrScheduleOverrideNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying schedule override: %w", err)
	}
	return &o, nil
}
