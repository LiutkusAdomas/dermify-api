package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"dermify-api/internal/domain"
	"dermify-api/internal/service"
)

// ErrPatientNotFound is returned when no patient matches the given ID.
var ErrPatientNotFound = errors.New("patient not found") //nolint:gochecknoglobals // sentinel error

// ErrPatientVersionConflict is returned when an update fails due to version mismatch.
var ErrPatientVersionConflict = errors.New("patient version conflict") //nolint:gochecknoglobals // sentinel error

// PostgresPatientRepository implements service.PatientRepository using PostgreSQL.
type PostgresPatientRepository struct {
	db *sql.DB
}

// NewPostgresPatientRepository creates a new PostgresPatientRepository.
func NewPostgresPatientRepository(db *sql.DB) *PostgresPatientRepository {
	return &PostgresPatientRepository{db: db}
}

// Create inserts a new patient record and sets the ID on the provided struct.
func (r *PostgresPatientRepository) Create(ctx context.Context, patient *domain.Patient) error {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO patients (first_name, last_name, date_of_birth, sex, phone, email,
			external_reference, version, created_at, created_by, updated_at, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id`,
		patient.FirstName, patient.LastName, patient.DateOfBirth, patient.Sex,
		patient.Phone, patient.Email, patient.ExternalReference,
		patient.Version, patient.CreatedAt, patient.CreatedBy, patient.UpdatedAt, patient.UpdatedBy,
	).Scan(&patient.ID)
	if err != nil {
		return fmt.Errorf("inserting patient: %w", err)
	}

	return nil
}

// GetByID retrieves a patient by ID.
func (r *PostgresPatientRepository) GetByID(ctx context.Context, id int64) (*domain.Patient, error) {
	var p domain.Patient

	err := r.db.QueryRowContext(ctx,
		`SELECT id, first_name, last_name, date_of_birth, sex, phone, email,
			external_reference, version, created_at, created_by, updated_at, updated_by
		FROM patients WHERE id = $1`, id,
	).Scan(
		&p.ID, &p.FirstName, &p.LastName, &p.DateOfBirth, &p.Sex,
		&p.Phone, &p.Email, &p.ExternalReference,
		&p.Version, &p.CreatedAt, &p.CreatedBy, &p.UpdatedAt, &p.UpdatedBy,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrPatientNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("querying patient: %w", err)
	}

	return &p, nil
}

// Update modifies a patient using optimistic locking on the version field.
func (r *PostgresPatientRepository) Update(ctx context.Context, patient *domain.Patient) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE patients SET first_name = $1, last_name = $2, date_of_birth = $3, sex = $4,
			phone = $5, email = $6, external_reference = $7,
			version = version + 1, updated_at = $8, updated_by = $9
		WHERE id = $10 AND version = $11`,
		patient.FirstName, patient.LastName, patient.DateOfBirth, patient.Sex,
		patient.Phone, patient.Email, patient.ExternalReference,
		patient.UpdatedAt, patient.UpdatedBy,
		patient.ID, patient.Version,
	)
	if err != nil {
		return fmt.Errorf("updating patient: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return service.ErrPatientVersionConflict
	}

	return nil
}

// List returns paginated patients matching the given filter.
func (r *PostgresPatientRepository) List(ctx context.Context, filter service.PatientFilter) (*service.PatientListResult, error) {
	baseQuery := `SELECT p.id, p.first_name, p.last_name, p.date_of_birth, p.sex,
		p.phone, p.email, p.external_reference, p.version,
		p.created_at, p.created_by, p.updated_at, p.updated_by,
		COALESCE(sess.session_count, 0), sess.last_session_date
	FROM patients p
	LEFT JOIN (
		SELECT patient_id,
			   COUNT(*) AS session_count,
			   MAX(created_at) AS last_session_date
		FROM sessions
		GROUP BY patient_id
	) sess ON sess.patient_id = p.id`
	countQuery := "SELECT COUNT(*) FROM patients p"

	whereClause := ""
	args := []interface{}{}
	argIndex := 1

	if filter.Search != "" {
		searchPattern := strings.ToLower(filter.Search) + "%"
		whereClause = fmt.Sprintf(
			` WHERE LOWER(p.last_name) LIKE $%d OR LOWER(p.first_name) LIKE $%d OR LOWER(p.email) LIKE $%d OR p.phone LIKE $%d`,
			argIndex, argIndex, argIndex, argIndex,
		)
		args = append(args, searchPattern)
		argIndex++
	}

	// Count total matching records.
	var total int
	err := r.db.QueryRowContext(ctx, countQuery+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("counting patients: %w", err)
	}

	// Apply ordering and pagination.
	orderClause := " ORDER BY p.last_name ASC"
	offset := (filter.Page - 1) * filter.PerPage
	limitClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, filter.PerPage, offset)

	rows, err := r.db.QueryContext(ctx, baseQuery+whereClause+orderClause+limitClause, args...)
	if err != nil {
		return nil, fmt.Errorf("querying patients: %w", err)
	}
	defer rows.Close()

	patients, err := scanPatientListItems(rows)
	if err != nil {
		return nil, err
	}

	return &service.PatientListResult{
		Patients: patients,
		Total:    total,
	}, nil
}

// GetSessionHistory returns session summaries for a patient.
func (r *PostgresPatientRepository) GetSessionHistory(ctx context.Context, patientID int64) ([]domain.SessionSummary, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, created_at, status FROM sessions
		WHERE patient_id = $1
		ORDER BY created_at DESC`, patientID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying session history: %w", err)
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

// scanPatientListItems scans rows into PatientListItem slice.
func scanPatientListItems(rows *sql.Rows) ([]service.PatientListItem, error) {
	patients := []service.PatientListItem{}

	for rows.Next() {
		var item service.PatientListItem
		err := rows.Scan(
			&item.ID, &item.FirstName, &item.LastName, &item.DateOfBirth, &item.Sex,
			&item.Phone, &item.Email, &item.ExternalReference,
			&item.Version, &item.CreatedAt, &item.CreatedBy, &item.UpdatedAt, &item.UpdatedBy,
			&item.SessionCount, &item.LastSessionDate,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning patient row: %w", err)
		}
		patients = append(patients, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating patient rows: %w", err)
	}

	return patients, nil
}
