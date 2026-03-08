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

// PostgresOutcomeRepository implements service.OutcomeRepository using PostgreSQL.
type PostgresOutcomeRepository struct {
	db *sql.DB
}

// NewPostgresOutcomeRepository creates a new PostgresOutcomeRepository.
func NewPostgresOutcomeRepository(db *sql.DB) *PostgresOutcomeRepository {
	return &PostgresOutcomeRepository{db: db}
}

// Create inserts a new session outcome record and sets the generated ID.
func (r *PostgresOutcomeRepository) Create(ctx context.Context, outcome *domain.SessionOutcome) error {
	now := time.Now()
	outcome.CreatedAt = now
	outcome.UpdatedAt = now

	err := r.db.QueryRowContext(ctx,
		`INSERT INTO session_outcomes (
			session_id, outcome_status, aftercare_notes, red_flags_text,
			contact_info, follow_up_at, notes, version,
			created_at, created_by, updated_at, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id`,
		outcome.SessionID, outcome.OutcomeStatus, outcome.AftercareNotes, outcome.RedFlagsText,
		outcome.ContactInfo, outcome.FollowUpAt, outcome.Notes, outcome.Version,
		outcome.CreatedAt, outcome.CreatedBy, outcome.UpdatedAt, outcome.UpdatedBy,
	).Scan(&outcome.ID)
	if err != nil {
		return fmt.Errorf("inserting session outcome: %w", err)
	}

	return nil
}

// GetBySessionID retrieves the outcome record for a session.
func (r *PostgresOutcomeRepository) GetBySessionID(ctx context.Context, sessionID int64) (*domain.SessionOutcome, error) {
	var o domain.SessionOutcome

	err := r.db.QueryRowContext(ctx,
		`SELECT id, session_id, outcome_status, aftercare_notes, red_flags_text,
			contact_info, follow_up_at, notes, version,
			created_at, created_by, updated_at, updated_by
		FROM session_outcomes WHERE session_id = $1`, sessionID,
	).Scan(
		&o.ID, &o.SessionID, &o.OutcomeStatus, &o.AftercareNotes, &o.RedFlagsText,
		&o.ContactInfo, &o.FollowUpAt, &o.Notes, &o.Version,
		&o.CreatedAt, &o.CreatedBy, &o.UpdatedAt, &o.UpdatedBy,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrOutcomeNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("querying session outcome: %w", err)
	}

	return &o, nil
}

// Update modifies a session outcome record using optimistic locking on the version field.
func (r *PostgresOutcomeRepository) Update(ctx context.Context, outcome *domain.SessionOutcome) error {
	now := time.Now()

	result, err := r.db.ExecContext(ctx,
		`UPDATE session_outcomes SET
			outcome_status = $1, aftercare_notes = $2, red_flags_text = $3,
			contact_info = $4, follow_up_at = $5, notes = $6,
			version = version + 1, updated_at = $7, updated_by = $8
		WHERE id = $9 AND version = $10`,
		outcome.OutcomeStatus, outcome.AftercareNotes, outcome.RedFlagsText,
		outcome.ContactInfo, outcome.FollowUpAt, outcome.Notes,
		now, outcome.UpdatedBy,
		outcome.ID, outcome.Version,
	)
	if err != nil {
		return fmt.Errorf("updating session outcome: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return service.ErrOutcomeNotFound
	}

	return nil
}

// ExistsForSession checks whether an outcome record exists for a session.
func (r *PostgresOutcomeRepository) ExistsForSession(ctx context.Context, sessionID int64) (bool, error) {
	var exists bool

	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM session_outcomes WHERE session_id = $1)`, sessionID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking outcome existence: %w", err)
	}

	return exists, nil
}

// SetEndpoints replaces the endpoint associations for an outcome.
// It deletes all existing associations and inserts the new ones.
func (r *PostgresOutcomeRepository) SetEndpoints(ctx context.Context, outcomeID int64, endpointIDs []int64) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM session_outcome_endpoints WHERE outcome_id = $1`, outcomeID,
	)
	if err != nil {
		return fmt.Errorf("deleting outcome endpoints: %w", err)
	}

	for _, endpointID := range endpointIDs {
		_, err := r.db.ExecContext(ctx,
			`INSERT INTO session_outcome_endpoints (outcome_id, endpoint_id)
			VALUES ($1, $2)`, outcomeID, endpointID,
		)
		if err != nil {
			return fmt.Errorf("inserting outcome endpoint %d: %w", endpointID, err)
		}
	}

	return nil
}

// GetEndpoints returns the endpoint IDs associated with an outcome.
func (r *PostgresOutcomeRepository) GetEndpoints(ctx context.Context, outcomeID int64) ([]int64, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT endpoint_id FROM session_outcome_endpoints
		WHERE outcome_id = $1 ORDER BY endpoint_id`, outcomeID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying outcome endpoints: %w", err)
	}
	defer rows.Close()

	var endpoints []int64

	for rows.Next() {
		var endpointID int64
		if err := rows.Scan(&endpointID); err != nil {
			return nil, fmt.Errorf("scanning outcome endpoint: %w", err)
		}

		endpoints = append(endpoints, endpointID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating outcome endpoints: %w", err)
	}

	return endpoints, nil
}
