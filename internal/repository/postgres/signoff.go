package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"dermify-api/internal/service"
)

// PostgresSignoffRepository implements service.SignoffRepository using PostgreSQL.
type PostgresSignoffRepository struct {
	db *sql.DB
}

// NewPostgresSignoffRepository creates a new PostgresSignoffRepository.
func NewPostgresSignoffRepository(db *sql.DB) *PostgresSignoffRepository {
	return &PostgresSignoffRepository{db: db}
}

// SignOff atomically transitions a session from awaiting_signoff to signed status,
// setting signed_at and signed_by with optimistic locking on the version field.
func (r *PostgresSignoffRepository) SignOff(ctx context.Context, id int64, clinicianID int64, expectedVersion int) error {
	now := time.Now()

	result, err := r.db.ExecContext(ctx,
		`UPDATE sessions SET
			status = 'signed', signed_at = $1, signed_by = $2,
			version = version + 1, updated_at = $1, updated_by = $2
		WHERE id = $3 AND version = $4 AND status = 'awaiting_signoff'`,
		now, clinicianID, id, expectedVersion,
	)
	if err != nil {
		return fmt.Errorf("signing off session: %w", err)
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

// LockSession transitions a signed session to the locked state
// with optimistic locking on the version field.
func (r *PostgresSignoffRepository) LockSession(ctx context.Context, id int64, expectedVersion int, userID int64) error {
	now := time.Now()

	result, err := r.db.ExecContext(ctx,
		`UPDATE sessions SET
			status = 'locked', version = version + 1,
			updated_at = $1, updated_by = $2
		WHERE id = $3 AND version = $4 AND status = 'signed'`,
		now, userID, id, expectedVersion,
	)
	if err != nil {
		return fmt.Errorf("locking session: %w", err)
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
