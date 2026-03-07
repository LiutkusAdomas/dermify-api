package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"dermify-api/internal/domain"
	"dermify-api/internal/service"
)

// PostgresSessionRepository implements service.SessionRepository using PostgreSQL.
type PostgresSessionRepository struct {
	db *sql.DB
}

// NewPostgresSessionRepository creates a new PostgresSessionRepository.
func NewPostgresSessionRepository(db *sql.DB) *PostgresSessionRepository {
	return &PostgresSessionRepository{db: db}
}

// Create inserts a new session record and sets the ID on the provided struct.
func (r *PostgresSessionRepository) Create(ctx context.Context, session *domain.Session) error {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO sessions (patient_id, clinician_id, status, scheduled_at, started_at,
			completed_at, patient_goal, fitzpatrick_type, is_tanned, is_pregnant,
			on_anticoagulants, photo_consent, notes, version,
			created_at, created_by, updated_at, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		RETURNING id`,
		session.PatientID, session.ClinicianID, session.Status,
		session.ScheduledAt, session.StartedAt, session.CompletedAt,
		session.PatientGoal, session.FitzpatrickType,
		session.IsTanned, session.IsPregnant, session.OnAnticoagulants,
		session.PhotoConsent, session.Notes,
		session.Version, session.CreatedAt, session.CreatedBy, session.UpdatedAt, session.UpdatedBy,
	).Scan(&session.ID)
	if err != nil {
		return fmt.Errorf("inserting session: %w", err)
	}

	return nil
}

// GetByID retrieves a session by ID, including its associated indication code IDs.
func (r *PostgresSessionRepository) GetByID(ctx context.Context, id int64) (*domain.Session, error) {
	var s domain.Session

	err := r.db.QueryRowContext(ctx,
		`SELECT id, patient_id, clinician_id, status, scheduled_at, started_at,
			completed_at, patient_goal, fitzpatrick_type, is_tanned, is_pregnant,
			on_anticoagulants, photo_consent, notes, version,
			created_at, created_by, updated_at, updated_by
		FROM sessions WHERE id = $1`, id,
	).Scan(
		&s.ID, &s.PatientID, &s.ClinicianID, &s.Status,
		&s.ScheduledAt, &s.StartedAt, &s.CompletedAt,
		&s.PatientGoal, &s.FitzpatrickType,
		&s.IsTanned, &s.IsPregnant, &s.OnAnticoagulants,
		&s.PhotoConsent, &s.Notes,
		&s.Version, &s.CreatedAt, &s.CreatedBy, &s.UpdatedAt, &s.UpdatedBy,
	)

	if err == sql.ErrNoRows {
		return nil, service.ErrSessionNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("querying session: %w", err)
	}

	// Load indication codes for this session.
	codes, err := r.loadIndicationCodes(ctx, id)
	if err != nil {
		return nil, err
	}

	s.IndicationCodes = codes

	return &s, nil
}

// Update modifies a session using optimistic locking on the version field.
func (r *PostgresSessionRepository) Update(ctx context.Context, session *domain.Session) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE sessions SET patient_goal = $1, fitzpatrick_type = $2,
			is_tanned = $3, is_pregnant = $4, on_anticoagulants = $5,
			photo_consent = $6, notes = $7, scheduled_at = $8,
			version = version + 1, updated_at = $9, updated_by = $10
		WHERE id = $11 AND version = $12`,
		session.PatientGoal, session.FitzpatrickType,
		session.IsTanned, session.IsPregnant, session.OnAnticoagulants,
		session.PhotoConsent, session.Notes, session.ScheduledAt,
		session.UpdatedAt, session.UpdatedBy,
		session.ID, session.Version,
	)
	if err != nil {
		return fmt.Errorf("updating session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return service.ErrSessionVersionConflict
	}

	return nil
}

// UpdateStatus transitions a session to a new status with optimistic locking.
func (r *PostgresSessionRepository) UpdateStatus(ctx context.Context, id int64, status string, expectedVersion int, userID int64) error {
	now := time.Now()

	result, err := r.db.ExecContext(ctx,
		`UPDATE sessions SET status = $1, version = version + 1,
			updated_at = $2, updated_by = $3
		WHERE id = $4 AND version = $5`,
		status, now, userID, id, expectedVersion,
	)
	if err != nil {
		return fmt.Errorf("updating session status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return service.ErrSessionVersionConflict
	}

	return nil
}

// List returns paginated sessions matching the given filter.
func (r *PostgresSessionRepository) List(ctx context.Context, filter service.SessionFilter) (*service.SessionListResult, error) {
	baseQuery := `SELECT id, patient_id, clinician_id, status, scheduled_at, started_at,
		completed_at, patient_goal, fitzpatrick_type, is_tanned, is_pregnant,
		on_anticoagulants, photo_consent, notes, version,
		created_at, created_by, updated_at, updated_by
	FROM sessions`
	countQuery := "SELECT COUNT(*) FROM sessions"

	var conditions []string
	args := []interface{}{}
	argIndex := 1

	if filter.PatientID > 0 {
		conditions = append(conditions, fmt.Sprintf("patient_id = $%d", argIndex))
		args = append(args, filter.PatientID)
		argIndex++
	}

	if filter.ClinicianID > 0 {
		conditions = append(conditions, fmt.Sprintf("clinician_id = $%d", argIndex))
		args = append(args, filter.ClinicianID)
		argIndex++
	}

	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, filter.Status)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total matching records.
	var total int

	err := r.db.QueryRowContext(ctx, countQuery+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("counting sessions: %w", err)
	}

	// Apply ordering and pagination.
	orderClause := " ORDER BY created_at DESC"
	offset := (filter.Page - 1) * filter.PerPage
	limitClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, filter.PerPage, offset)

	rows, err := r.db.QueryContext(ctx, baseQuery+whereClause+orderClause+limitClause, args...)
	if err != nil {
		return nil, fmt.Errorf("querying sessions: %w", err)
	}
	defer rows.Close()

	sessions, err := scanSessions(rows)
	if err != nil {
		return nil, err
	}

	return &service.SessionListResult{
		Sessions: sessions,
		Total:    total,
	}, nil
}

// ListByPatient returns session summaries for a patient ordered by creation date descending.
func (r *PostgresSessionRepository) ListByPatient(ctx context.Context, patientID int64) ([]domain.SessionSummary, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, created_at, status FROM sessions
		WHERE patient_id = $1 ORDER BY created_at DESC`, patientID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying patient sessions: %w", err)
	}
	defer rows.Close()

	summaries := []domain.SessionSummary{}

	for rows.Next() {
		var s domain.SessionSummary
		if err := rows.Scan(&s.ID, &s.CreatedAt, &s.Status); err != nil {
			return nil, fmt.Errorf("scanning session summary: %w", err)
		}

		summaries = append(summaries, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating session summaries: %w", err)
	}

	return summaries, nil
}

// SetIndicationCodes replaces the indication code associations for a session.
// It deletes all existing associations and inserts the new ones.
func (r *PostgresSessionRepository) SetIndicationCodes(ctx context.Context, sessionID int64, codeIDs []int64) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM session_indication_codes WHERE session_id = $1`, sessionID,
	)
	if err != nil {
		return fmt.Errorf("deleting indication codes: %w", err)
	}

	for _, codeID := range codeIDs {
		_, err := r.db.ExecContext(ctx,
			`INSERT INTO session_indication_codes (session_id, indication_code_id)
			VALUES ($1, $2)`, sessionID, codeID,
		)
		if err != nil {
			return fmt.Errorf("inserting indication code %d: %w", codeID, err)
		}
	}

	return nil
}

// loadIndicationCodes retrieves indication code IDs for a session.
func (r *PostgresSessionRepository) loadIndicationCodes(ctx context.Context, sessionID int64) ([]int64, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT indication_code_id FROM session_indication_codes
		WHERE session_id = $1 ORDER BY indication_code_id`, sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying indication codes: %w", err)
	}
	defer rows.Close()

	var codes []int64

	for rows.Next() {
		var codeID int64
		if err := rows.Scan(&codeID); err != nil {
			return nil, fmt.Errorf("scanning indication code: %w", err)
		}

		codes = append(codes, codeID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating indication codes: %w", err)
	}

	return codes, nil
}

// scanSessions scans rows into a Session slice.
func scanSessions(rows *sql.Rows) ([]domain.Session, error) {
	sessions := []domain.Session{}

	for rows.Next() {
		var s domain.Session
		err := rows.Scan(
			&s.ID, &s.PatientID, &s.ClinicianID, &s.Status,
			&s.ScheduledAt, &s.StartedAt, &s.CompletedAt,
			&s.PatientGoal, &s.FitzpatrickType,
			&s.IsTanned, &s.IsPregnant, &s.OnAnticoagulants,
			&s.PhotoConsent, &s.Notes,
			&s.Version, &s.CreatedAt, &s.CreatedBy, &s.UpdatedAt, &s.UpdatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning session row: %w", err)
		}

		sessions = append(sessions, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating session rows: %w", err)
	}

	return sessions, nil
}
