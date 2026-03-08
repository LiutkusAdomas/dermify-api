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

// PostgresCO2ModuleRepository implements service.CO2ModuleRepository using PostgreSQL.
type PostgresCO2ModuleRepository struct {
	db *sql.DB
}

// NewPostgresCO2ModuleRepository creates a new PostgresCO2ModuleRepository.
func NewPostgresCO2ModuleRepository(db *sql.DB) *PostgresCO2ModuleRepository {
	return &PostgresCO2ModuleRepository{db: db}
}

// Create inserts a new CO2 module detail record and sets the generated ID.
func (r *PostgresCO2ModuleRepository) Create(ctx context.Context, detail *domain.CO2ModuleDetail) error {
	now := time.Now()
	detail.CreatedAt = now
	detail.UpdatedAt = now

	err := r.db.QueryRowContext(ctx,
		`INSERT INTO co2_module_details (
			module_id, device_id, handpiece_id, mode, scanner_pattern,
			power, pulse_energy, pulse_duration, density, pattern,
			passes, anaesthesia_used, notes, version,
			created_at, created_by, updated_at, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		RETURNING id`,
		detail.ModuleID, detail.DeviceID, detail.HandpieceID,
		detail.Mode, detail.ScannerPattern,
		detail.Power, detail.PulseEnergy, detail.PulseDuration,
		detail.Density, detail.Pattern, detail.Passes,
		detail.AnaesthesiaUsed, detail.Notes, detail.Version,
		detail.CreatedAt, detail.CreatedBy, detail.UpdatedAt, detail.UpdatedBy,
	).Scan(&detail.ID)
	if err != nil {
		return fmt.Errorf("inserting CO2 module detail: %w", err)
	}

	return nil
}

// GetByModuleID retrieves a CO2 module detail by its parent module ID.
func (r *PostgresCO2ModuleRepository) GetByModuleID(ctx context.Context, moduleID int64) (*domain.CO2ModuleDetail, error) {
	var d domain.CO2ModuleDetail

	err := r.db.QueryRowContext(ctx,
		`SELECT id, module_id, device_id, handpiece_id, mode, scanner_pattern,
			power, pulse_energy, pulse_duration, density, pattern,
			passes, anaesthesia_used, notes, version,
			created_at, created_by, updated_at, updated_by
		FROM co2_module_details WHERE module_id = $1`, moduleID,
	).Scan(
		&d.ID, &d.ModuleID, &d.DeviceID, &d.HandpieceID,
		&d.Mode, &d.ScannerPattern,
		&d.Power, &d.PulseEnergy, &d.PulseDuration,
		&d.Density, &d.Pattern, &d.Passes,
		&d.AnaesthesiaUsed, &d.Notes, &d.Version,
		&d.CreatedAt, &d.CreatedBy, &d.UpdatedAt, &d.UpdatedBy,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrModuleDetailNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("querying CO2 module detail: %w", err)
	}

	return &d, nil
}

// Update modifies a CO2 module detail using optimistic locking on the version field.
func (r *PostgresCO2ModuleRepository) Update(ctx context.Context, detail *domain.CO2ModuleDetail) error {
	now := time.Now()

	result, err := r.db.ExecContext(ctx,
		`UPDATE co2_module_details SET
			device_id = $1, handpiece_id = $2, mode = $3, scanner_pattern = $4,
			power = $5, pulse_energy = $6, pulse_duration = $7, density = $8,
			pattern = $9, passes = $10, anaesthesia_used = $11, notes = $12,
			version = version + 1, updated_at = $13, updated_by = $14
		WHERE id = $15 AND version = $16`,
		detail.DeviceID, detail.HandpieceID, detail.Mode, detail.ScannerPattern,
		detail.Power, detail.PulseEnergy, detail.PulseDuration, detail.Density,
		detail.Pattern, detail.Passes, detail.AnaesthesiaUsed, detail.Notes,
		now, detail.UpdatedBy,
		detail.ID, detail.Version,
	)
	if err != nil {
		return fmt.Errorf("updating CO2 module detail: %w", err)
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
