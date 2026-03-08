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

// PostgresIPLModuleRepository implements service.IPLModuleRepository using PostgreSQL.
type PostgresIPLModuleRepository struct {
	db *sql.DB
}

// NewPostgresIPLModuleRepository creates a new PostgresIPLModuleRepository.
func NewPostgresIPLModuleRepository(db *sql.DB) *PostgresIPLModuleRepository {
	return &PostgresIPLModuleRepository{db: db}
}

// Create inserts a new IPL module detail record and sets the generated ID.
func (r *PostgresIPLModuleRepository) Create(ctx context.Context, detail *domain.IPLModuleDetail) error {
	now := time.Now()
	detail.CreatedAt = now
	detail.UpdatedAt = now

	err := r.db.QueryRowContext(ctx,
		`INSERT INTO ipl_module_details (
			module_id, device_id, handpiece_id, filter_band, lightguide_size,
			fluence, pulse_duration, pulse_delay, pulse_count, passes,
			total_pulses, cooling_mode, notes, version,
			created_at, created_by, updated_at, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		RETURNING id`,
		detail.ModuleID, detail.DeviceID, detail.HandpieceID,
		detail.FilterBand, detail.LightguideSize,
		detail.Fluence, detail.PulseDuration, detail.PulseDelay,
		detail.PulseCount, detail.Passes, detail.TotalPulses,
		detail.CoolingMode, detail.Notes, detail.Version,
		detail.CreatedAt, detail.CreatedBy, detail.UpdatedAt, detail.UpdatedBy,
	).Scan(&detail.ID)
	if err != nil {
		return fmt.Errorf("inserting IPL module detail: %w", err)
	}

	return nil
}

// GetByModuleID retrieves an IPL module detail by its parent module ID.
func (r *PostgresIPLModuleRepository) GetByModuleID(ctx context.Context, moduleID int64) (*domain.IPLModuleDetail, error) {
	var d domain.IPLModuleDetail

	err := r.db.QueryRowContext(ctx,
		`SELECT id, module_id, device_id, handpiece_id, filter_band, lightguide_size,
			fluence, pulse_duration, pulse_delay, pulse_count, passes,
			total_pulses, cooling_mode, notes, version,
			created_at, created_by, updated_at, updated_by
		FROM ipl_module_details WHERE module_id = $1`, moduleID,
	).Scan(
		&d.ID, &d.ModuleID, &d.DeviceID, &d.HandpieceID,
		&d.FilterBand, &d.LightguideSize,
		&d.Fluence, &d.PulseDuration, &d.PulseDelay,
		&d.PulseCount, &d.Passes, &d.TotalPulses,
		&d.CoolingMode, &d.Notes, &d.Version,
		&d.CreatedAt, &d.CreatedBy, &d.UpdatedAt, &d.UpdatedBy,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrModuleDetailNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("querying IPL module detail: %w", err)
	}

	return &d, nil
}

// Update modifies an IPL module detail using optimistic locking on the version field.
func (r *PostgresIPLModuleRepository) Update(ctx context.Context, detail *domain.IPLModuleDetail) error {
	now := time.Now()

	result, err := r.db.ExecContext(ctx,
		`UPDATE ipl_module_details SET
			device_id = $1, handpiece_id = $2, filter_band = $3, lightguide_size = $4,
			fluence = $5, pulse_duration = $6, pulse_delay = $7, pulse_count = $8,
			passes = $9, total_pulses = $10, cooling_mode = $11, notes = $12,
			version = version + 1, updated_at = $13, updated_by = $14
		WHERE id = $15 AND version = $16`,
		detail.DeviceID, detail.HandpieceID, detail.FilterBand, detail.LightguideSize,
		detail.Fluence, detail.PulseDuration, detail.PulseDelay, detail.PulseCount,
		detail.Passes, detail.TotalPulses, detail.CoolingMode, detail.Notes,
		now, detail.UpdatedBy,
		detail.ID, detail.Version,
	)
	if err != nil {
		return fmt.Errorf("updating IPL module detail: %w", err)
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
