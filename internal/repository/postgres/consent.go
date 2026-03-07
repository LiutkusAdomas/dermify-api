package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"dermify-api/internal/domain"
	"dermify-api/internal/service"
)

// PostgresConsentRepository implements service.ConsentRepository using PostgreSQL.
type PostgresConsentRepository struct {
	db *sql.DB
}

// NewPostgresConsentRepository creates a new PostgresConsentRepository.
func NewPostgresConsentRepository(db *sql.DB) *PostgresConsentRepository {
	return &PostgresConsentRepository{db: db}
}

// Create inserts a new consent record and sets the ID on the provided struct.
func (r *PostgresConsentRepository) Create(ctx context.Context, consent *domain.Consent) error {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO session_consents (session_id, consent_type, consent_method, obtained_at,
			risks_discussed, notes, version, created_at, created_by, updated_at, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id`,
		consent.SessionID, consent.ConsentType, consent.ConsentMethod, consent.ObtainedAt,
		consent.RisksDiscussed, consent.Notes, consent.Version,
		consent.CreatedAt, consent.CreatedBy, consent.UpdatedAt, consent.UpdatedBy,
	).Scan(&consent.ID)
	if err != nil {
		return fmt.Errorf("inserting consent: %w", err)
	}

	return nil
}

// GetBySessionID retrieves the consent record for a session.
func (r *PostgresConsentRepository) GetBySessionID(ctx context.Context, sessionID int64) (*domain.Consent, error) {
	var c domain.Consent

	err := r.db.QueryRowContext(ctx,
		`SELECT id, session_id, consent_type, consent_method, obtained_at,
			risks_discussed, notes, version, created_at, created_by, updated_at, updated_by
		FROM session_consents WHERE session_id = $1`, sessionID,
	).Scan(
		&c.ID, &c.SessionID, &c.ConsentType, &c.ConsentMethod, &c.ObtainedAt,
		&c.RisksDiscussed, &c.Notes, &c.Version,
		&c.CreatedAt, &c.CreatedBy, &c.UpdatedAt, &c.UpdatedBy,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrConsentNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("querying consent: %w", err)
	}

	return &c, nil
}

// Update modifies a consent record using optimistic locking on the version field.
func (r *PostgresConsentRepository) Update(ctx context.Context, consent *domain.Consent) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE session_consents SET consent_type = $1, consent_method = $2, obtained_at = $3,
			risks_discussed = $4, notes = $5, version = version + 1,
			updated_at = $6, updated_by = $7
		WHERE id = $8 AND version = $9`,
		consent.ConsentType, consent.ConsentMethod, consent.ObtainedAt,
		consent.RisksDiscussed, consent.Notes,
		consent.UpdatedAt, consent.UpdatedBy,
		consent.ID, consent.Version,
	)
	if err != nil {
		return fmt.Errorf("updating consent: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("consent version conflict: %w", service.ErrConsentNotFound)
	}

	return nil
}

// ExistsForSession checks whether a consent record exists for a session.
func (r *PostgresConsentRepository) ExistsForSession(ctx context.Context, sessionID int64) (bool, error) {
	var exists bool

	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM session_consents WHERE session_id = $1)`, sessionID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking consent existence: %w", err)
	}

	return exists, nil
}
