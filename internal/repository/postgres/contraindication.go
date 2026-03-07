package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"dermify-api/internal/domain"
	"dermify-api/internal/service"
)

// PostgresContraindicationRepository implements service.ContraindicationRepository
// using PostgreSQL.
type PostgresContraindicationRepository struct {
	db *sql.DB
}

// NewPostgresContraindicationRepository creates a new PostgresContraindicationRepository.
func NewPostgresContraindicationRepository(db *sql.DB) *PostgresContraindicationRepository {
	return &PostgresContraindicationRepository{db: db}
}

// Create inserts a new screening record and sets the ID on the provided struct.
func (r *PostgresContraindicationRepository) Create(ctx context.Context, screening *domain.ContraindicationScreening) error {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO contraindication_screenings (session_id, pregnant, breastfeeding,
			active_infection, active_cold_sores, isotretinoin, photosensitivity,
			autoimmune_disorder, keloid_history, anticoagulants, recent_tan,
			has_flags, mitigation_notes, notes, version,
			created_at, created_by, updated_at, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
		RETURNING id`,
		screening.SessionID, screening.Pregnant, screening.Breastfeeding,
		screening.ActiveInfection, screening.ActiveColdSores, screening.Isotretinoin,
		screening.Photosensitivity, screening.AutoimmuneDisorder, screening.KeloidHistory,
		screening.Anticoagulants, screening.RecentTan,
		screening.HasFlags, screening.MitigationNotes, screening.Notes, screening.Version,
		screening.CreatedAt, screening.CreatedBy, screening.UpdatedAt, screening.UpdatedBy,
	).Scan(&screening.ID)
	if err != nil {
		return fmt.Errorf("inserting screening: %w", err)
	}

	return nil
}

// GetBySessionID retrieves the screening record for a session.
func (r *PostgresContraindicationRepository) GetBySessionID(ctx context.Context, sessionID int64) (*domain.ContraindicationScreening, error) {
	var s domain.ContraindicationScreening

	err := r.db.QueryRowContext(ctx,
		`SELECT id, session_id, pregnant, breastfeeding, active_infection, active_cold_sores,
			isotretinoin, photosensitivity, autoimmune_disorder, keloid_history,
			anticoagulants, recent_tan, has_flags, mitigation_notes, notes, version,
			created_at, created_by, updated_at, updated_by
		FROM contraindication_screenings WHERE session_id = $1`, sessionID,
	).Scan(
		&s.ID, &s.SessionID, &s.Pregnant, &s.Breastfeeding, &s.ActiveInfection,
		&s.ActiveColdSores, &s.Isotretinoin, &s.Photosensitivity, &s.AutoimmuneDisorder,
		&s.KeloidHistory, &s.Anticoagulants, &s.RecentTan,
		&s.HasFlags, &s.MitigationNotes, &s.Notes, &s.Version,
		&s.CreatedAt, &s.CreatedBy, &s.UpdatedAt, &s.UpdatedBy,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrScreeningNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("querying screening: %w", err)
	}

	return &s, nil
}

// Update modifies a screening record using optimistic locking on the version field.
func (r *PostgresContraindicationRepository) Update(ctx context.Context, screening *domain.ContraindicationScreening) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE contraindication_screenings SET pregnant = $1, breastfeeding = $2,
			active_infection = $3, active_cold_sores = $4, isotretinoin = $5,
			photosensitivity = $6, autoimmune_disorder = $7, keloid_history = $8,
			anticoagulants = $9, recent_tan = $10, has_flags = $11,
			mitigation_notes = $12, notes = $13, version = version + 1,
			updated_at = $14, updated_by = $15
		WHERE id = $16 AND version = $17`,
		screening.Pregnant, screening.Breastfeeding,
		screening.ActiveInfection, screening.ActiveColdSores, screening.Isotretinoin,
		screening.Photosensitivity, screening.AutoimmuneDisorder, screening.KeloidHistory,
		screening.Anticoagulants, screening.RecentTan, screening.HasFlags,
		screening.MitigationNotes, screening.Notes,
		screening.UpdatedAt, screening.UpdatedBy,
		screening.ID, screening.Version,
	)
	if err != nil {
		return fmt.Errorf("updating screening: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("screening version conflict: %w", service.ErrScreeningNotFound)
	}

	return nil
}
