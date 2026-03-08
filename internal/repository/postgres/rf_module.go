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

// PostgresRFModuleRepository implements service.RFModuleRepository using PostgreSQL.
type PostgresRFModuleRepository struct {
	db *sql.DB
}

// NewPostgresRFModuleRepository creates a new PostgresRFModuleRepository.
func NewPostgresRFModuleRepository(db *sql.DB) *PostgresRFModuleRepository {
	return &PostgresRFModuleRepository{db: db}
}

// Create inserts a new RF module detail record and sets the generated ID.
func (r *PostgresRFModuleRepository) Create(ctx context.Context, detail *domain.RFModuleDetail) error {
	now := time.Now()
	detail.CreatedAt = now
	detail.UpdatedAt = now

	err := r.db.QueryRowContext(ctx,
		`INSERT INTO rf_module_details (
			module_id, device_id, handpiece_id, rf_mode, tip_type,
			depth, energy_level, overlap, pulses_per_zone,
			total_pulses, notes, version,
			created_at, created_by, updated_at, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id`,
		detail.ModuleID, detail.DeviceID, detail.HandpieceID,
		detail.RFMode, detail.TipType,
		detail.Depth, detail.EnergyLevel, detail.Overlap,
		detail.PulsesPerZone, detail.TotalPulses, detail.Notes, detail.Version,
		detail.CreatedAt, detail.CreatedBy, detail.UpdatedAt, detail.UpdatedBy,
	).Scan(&detail.ID)
	if err != nil {
		return fmt.Errorf("inserting RF module detail: %w", err)
	}

	return nil
}

// GetByModuleID retrieves an RF module detail by its parent module ID.
func (r *PostgresRFModuleRepository) GetByModuleID(ctx context.Context, moduleID int64) (*domain.RFModuleDetail, error) {
	var d domain.RFModuleDetail

	err := r.db.QueryRowContext(ctx,
		`SELECT id, module_id, device_id, handpiece_id, rf_mode, tip_type,
			depth, energy_level, overlap, pulses_per_zone,
			total_pulses, notes, version,
			created_at, created_by, updated_at, updated_by
		FROM rf_module_details WHERE module_id = $1`, moduleID,
	).Scan(
		&d.ID, &d.ModuleID, &d.DeviceID, &d.HandpieceID,
		&d.RFMode, &d.TipType,
		&d.Depth, &d.EnergyLevel, &d.Overlap,
		&d.PulsesPerZone, &d.TotalPulses, &d.Notes, &d.Version,
		&d.CreatedAt, &d.CreatedBy, &d.UpdatedAt, &d.UpdatedBy,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrModuleDetailNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("querying RF module detail: %w", err)
	}

	return &d, nil
}

// Update modifies an RF module detail using optimistic locking on the version field.
func (r *PostgresRFModuleRepository) Update(ctx context.Context, detail *domain.RFModuleDetail) error {
	now := time.Now()

	result, err := r.db.ExecContext(ctx,
		`UPDATE rf_module_details SET
			device_id = $1, handpiece_id = $2, rf_mode = $3, tip_type = $4,
			depth = $5, energy_level = $6, overlap = $7, pulses_per_zone = $8,
			total_pulses = $9, notes = $10,
			version = version + 1, updated_at = $11, updated_by = $12
		WHERE id = $13 AND version = $14`,
		detail.DeviceID, detail.HandpieceID, detail.RFMode, detail.TipType,
		detail.Depth, detail.EnergyLevel, detail.Overlap, detail.PulsesPerZone,
		detail.TotalPulses, detail.Notes,
		now, detail.UpdatedBy,
		detail.ID, detail.Version,
	)
	if err != nil {
		return fmt.Errorf("updating RF module detail: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return service.ErrModuleDetailVersionConflict
	}

	return nil
}
