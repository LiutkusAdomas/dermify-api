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

// PostgresNdYAGModuleRepository implements service.NdYAGModuleRepository using PostgreSQL.
type PostgresNdYAGModuleRepository struct {
	db *sql.DB
}

// NewPostgresNdYAGModuleRepository creates a new PostgresNdYAGModuleRepository.
func NewPostgresNdYAGModuleRepository(db *sql.DB) *PostgresNdYAGModuleRepository {
	return &PostgresNdYAGModuleRepository{db: db}
}

// Create inserts a new Nd:YAG module detail record and sets the generated ID.
func (r *PostgresNdYAGModuleRepository) Create(ctx context.Context, detail *domain.NdYAGModuleDetail) error {
	now := time.Now()
	detail.CreatedAt = now
	detail.UpdatedAt = now

	err := r.db.QueryRowContext(ctx,
		`INSERT INTO ndyag_module_details (
			module_id, device_id, handpiece_id, wavelength, spot_size,
			fluence, pulse_duration, repetition_rate, cooling_type,
			total_pulses, notes, version,
			created_at, created_by, updated_at, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id`,
		detail.ModuleID, detail.DeviceID, detail.HandpieceID,
		detail.Wavelength, detail.SpotSize,
		detail.Fluence, detail.PulseDuration, detail.RepetitionRate,
		detail.CoolingType, detail.TotalPulses, detail.Notes, detail.Version,
		detail.CreatedAt, detail.CreatedBy, detail.UpdatedAt, detail.UpdatedBy,
	).Scan(&detail.ID)
	if err != nil {
		return fmt.Errorf("inserting NdYAG module detail: %w", err)
	}

	return nil
}

// GetByModuleID retrieves an Nd:YAG module detail by its parent module ID.
func (r *PostgresNdYAGModuleRepository) GetByModuleID(ctx context.Context, moduleID int64) (*domain.NdYAGModuleDetail, error) {
	var d domain.NdYAGModuleDetail

	err := r.db.QueryRowContext(ctx,
		`SELECT id, module_id, device_id, handpiece_id, wavelength, spot_size,
			fluence, pulse_duration, repetition_rate, cooling_type,
			total_pulses, notes, version,
			created_at, created_by, updated_at, updated_by
		FROM ndyag_module_details WHERE module_id = $1`, moduleID,
	).Scan(
		&d.ID, &d.ModuleID, &d.DeviceID, &d.HandpieceID,
		&d.Wavelength, &d.SpotSize,
		&d.Fluence, &d.PulseDuration, &d.RepetitionRate,
		&d.CoolingType, &d.TotalPulses, &d.Notes, &d.Version,
		&d.CreatedAt, &d.CreatedBy, &d.UpdatedAt, &d.UpdatedBy,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrModuleDetailNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("querying NdYAG module detail: %w", err)
	}

	return &d, nil
}

// Update modifies an Nd:YAG module detail using optimistic locking on the version field.
func (r *PostgresNdYAGModuleRepository) Update(ctx context.Context, detail *domain.NdYAGModuleDetail) error {
	now := time.Now()

	result, err := r.db.ExecContext(ctx,
		`UPDATE ndyag_module_details SET
			device_id = $1, handpiece_id = $2, wavelength = $3, spot_size = $4,
			fluence = $5, pulse_duration = $6, repetition_rate = $7, cooling_type = $8,
			total_pulses = $9, notes = $10,
			version = version + 1, updated_at = $11, updated_by = $12
		WHERE id = $13 AND version = $14`,
		detail.DeviceID, detail.HandpieceID, detail.Wavelength, detail.SpotSize,
		detail.Fluence, detail.PulseDuration, detail.RepetitionRate, detail.CoolingType,
		detail.TotalPulses, detail.Notes,
		now, detail.UpdatedBy,
		detail.ID, detail.Version,
	)
	if err != nil {
		return fmt.Errorf("updating NdYAG module detail: %w", err)
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
