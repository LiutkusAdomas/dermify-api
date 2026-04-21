package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"dermify-api/internal/domain"
	"dermify-api/internal/service"
)

// PostgresAppointmentRepository implements service.AppointmentRepository.
type PostgresAppointmentRepository struct {
	db *sql.DB
}

// NewPostgresAppointmentRepository creates a new PostgresAppointmentRepository.
func NewPostgresAppointmentRepository(db *sql.DB) *PostgresAppointmentRepository {
	return &PostgresAppointmentRepository{db: db}
}

// Create inserts a new appointment.
func (r *PostgresAppointmentRepository) Create(ctx context.Context, a *domain.Appointment) error {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO appointments (org_id, patient_id, doctor_id, service_type_id, start_time, end_time, status, notes, created_by, version)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 RETURNING id, created_at, updated_at`,
		a.OrgID, a.PatientID, a.DoctorID, a.ServiceTypeID, a.StartTime, a.EndTime, a.Status, a.Notes, a.CreatedBy, a.Version,
	).Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return fmt.Errorf("inserting appointment: %w", err)
	}
	return nil
}

const appointmentSelectColumns = `a.id, a.org_id, a.patient_id, a.doctor_id, a.service_type_id,
	a.start_time, a.end_time, a.status, COALESCE(a.notes, ''), COALESCE(a.cancellation_reason, ''),
	a.session_id, a.created_by, a.created_at, a.updated_at, a.version,
	COALESCE(p.first_name, ''), COALESCE(p.last_name, ''),
	COALESCE(u.username, ''), COALESCE(st.name, '')`

const appointmentJoins = ` FROM appointments a
	LEFT JOIN patients p ON p.id = a.patient_id
	LEFT JOIN users u ON u.id = a.doctor_id
	LEFT JOIN service_types st ON st.id = a.service_type_id`

func scanAppointment(scanner interface{ Scan(...interface{}) error }) (*domain.Appointment, error) {
	var a domain.Appointment
	err := scanner.Scan(
		&a.ID, &a.OrgID, &a.PatientID, &a.DoctorID, &a.ServiceTypeID,
		&a.StartTime, &a.EndTime, &a.Status, &a.Notes, &a.CancellationReason,
		&a.SessionID, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt, &a.Version,
		&a.PatientFirstName, &a.PatientLastName,
		&a.DoctorName, &a.ServiceTypeName,
	)
	return &a, err
}

// GetByID retrieves an appointment by ID within an organization.
func (r *PostgresAppointmentRepository) GetByID(ctx context.Context, id, orgID int64) (*domain.Appointment, error) {
	query := `SELECT ` + appointmentSelectColumns + appointmentJoins + ` WHERE a.id = $1 AND a.org_id = $2`
	a, err := scanAppointment(r.db.QueryRowContext(ctx, query, id, orgID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrAppointmentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying appointment: %w", err)
	}
	return a, nil
}

// Update modifies an appointment with optimistic locking.
func (r *PostgresAppointmentRepository) Update(ctx context.Context, a *domain.Appointment) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE appointments SET
			patient_id = $1, doctor_id = $2, service_type_id = $3,
			start_time = $4, end_time = $5, notes = $6, updated_at = $7,
			version = version + 1
		 WHERE id = $8 AND org_id = $9 AND version = $10`,
		a.PatientID, a.DoctorID, a.ServiceTypeID,
		a.StartTime, a.EndTime, a.Notes, a.UpdatedAt,
		a.ID, a.OrgID, a.Version,
	)
	if err != nil {
		return fmt.Errorf("updating appointment: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return service.ErrAppointmentVersionConflict
	}
	a.Version++
	return nil
}

// UpdateStatus changes the appointment status with optimistic locking.
func (r *PostgresAppointmentRepository) UpdateStatus(ctx context.Context, id, orgID int64, status string, version int, cancellationReason string) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE appointments SET status = $1, cancellation_reason = $2, updated_at = $3, version = version + 1
		 WHERE id = $4 AND org_id = $5 AND version = $6`,
		status, cancellationReason, time.Now().UTC(), id, orgID, version,
	)
	if err != nil {
		return fmt.Errorf("updating appointment status: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return service.ErrAppointmentVersionConflict
	}
	return nil
}

// LinkSession sets the session_id on an appointment and updates status to in_progress.
func (r *PostgresAppointmentRepository) LinkSession(ctx context.Context, id, orgID int64, sessionID int64, version int) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE appointments SET session_id = $1, status = 'in_progress', updated_at = $2, version = version + 1
		 WHERE id = $3 AND org_id = $4 AND version = $5`,
		sessionID, time.Now().UTC(), id, orgID, version,
	)
	if err != nil {
		return fmt.Errorf("linking session to appointment: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return service.ErrAppointmentVersionConflict
	}
	return nil
}

// List returns paginated appointments matching the given filter.
func (r *PostgresAppointmentRepository) List(ctx context.Context, filter service.AppointmentFilter) (*service.AppointmentListResult, error) {
	conditions := []string{"a.org_id = $1"}
	args := []interface{}{filter.OrgID}
	argIdx := 2

	if filter.DoctorID > 0 {
		conditions = append(conditions, "a.doctor_id = $"+strconv.Itoa(argIdx))
		args = append(args, filter.DoctorID)
		argIdx++
	}
	if filter.PatientID > 0 {
		conditions = append(conditions, "a.patient_id = $"+strconv.Itoa(argIdx))
		args = append(args, filter.PatientID)
		argIdx++
	}
	if filter.Status != "" {
		conditions = append(conditions, "a.status = $"+strconv.Itoa(argIdx))
		args = append(args, filter.Status)
		argIdx++
	}
	if !filter.Start.IsZero() {
		conditions = append(conditions, "a.start_time >= $"+strconv.Itoa(argIdx))
		args = append(args, filter.Start)
		argIdx++
	}
	if !filter.End.IsZero() {
		conditions = append(conditions, "a.start_time < $"+strconv.Itoa(argIdx))
		args = append(args, filter.End)
		argIdx++
	}

	where := " WHERE " + strings.Join(conditions, " AND ")

	var total int
	countQuery := "SELECT COUNT(*)" + appointmentJoins + where
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("counting appointments: %w", err)
	}

	offset := (filter.Page - 1) * filter.PerPage
	query := "SELECT " + appointmentSelectColumns + appointmentJoins + where +
		" ORDER BY a.start_time ASC" +
		" LIMIT $" + strconv.Itoa(argIdx) + " OFFSET $" + strconv.Itoa(argIdx+1)
	args = append(args, filter.PerPage, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying appointments: %w", err)
	}
	defer rows.Close()

	var appointments []domain.Appointment
	for rows.Next() {
		a, err := scanAppointment(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning appointment: %w", err)
		}
		appointments = append(appointments, *a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &service.AppointmentListResult{
		Appointments: appointments,
		Total:        total,
	}, nil
}

// HasOverlap checks if a time range overlaps with existing non-cancelled appointments.
func (r *PostgresAppointmentRepository) HasOverlap(ctx context.Context, orgID, doctorID int64, start, end time.Time, excludeID int64) (bool, error) {
	query := `SELECT COUNT(*) FROM appointments
		 WHERE org_id = $1 AND doctor_id = $2
		 AND status NOT IN ('cancelled', 'no_show', 'completed')
		 AND start_time < $3 AND end_time > $4`
	args := []interface{}{orgID, doctorID, end, start}

	if excludeID > 0 {
		query += ` AND id != $5`
		args = append(args, excludeID)
	}

	var count int
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return false, fmt.Errorf("checking overlap: %w", err)
	}
	return count > 0, nil
}

// GetTimeSlotsForDate returns booked time slots for a doctor on a given date.
func (r *PostgresAppointmentRepository) GetTimeSlotsForDate(ctx context.Context, orgID, doctorID int64, date time.Time) ([]domain.TimeSlot, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.AddDate(0, 0, 1)

	rows, err := r.db.QueryContext(ctx,
		`SELECT start_time, end_time FROM appointments
		 WHERE org_id = $1 AND doctor_id = $2
		 AND start_time >= $3 AND start_time < $4
		 AND status NOT IN ('cancelled', 'no_show')
		 ORDER BY start_time`,
		orgID, doctorID, startOfDay, endOfDay,
	)
	if err != nil {
		return nil, fmt.Errorf("querying appointments for date: %w", err)
	}
	defer rows.Close()

	var slots []domain.TimeSlot
	for rows.Next() {
		var s domain.TimeSlot
		if err := rows.Scan(&s.Start, &s.End); err != nil {
			return nil, fmt.Errorf("scanning time slot: %w", err)
		}
		slots = append(slots, s)
	}
	return slots, rows.Err()
}

// GetBySessionID retrieves an appointment by its linked session ID.
func (r *PostgresAppointmentRepository) GetBySessionID(ctx context.Context, orgID, sessionID int64) (*domain.Appointment, error) {
	query := `SELECT ` + appointmentSelectColumns + appointmentJoins + ` WHERE a.session_id = $1 AND a.org_id = $2`
	a, err := scanAppointment(r.db.QueryRowContext(ctx, query, sessionID, orgID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrAppointmentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying appointment by session: %w", err)
	}
	return a, nil
}
