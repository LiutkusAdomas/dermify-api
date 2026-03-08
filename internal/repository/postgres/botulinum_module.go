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

// PostgresBotulinumModuleRepository implements service.BotulinumModuleRepository using PostgreSQL.
type PostgresBotulinumModuleRepository struct {
	db *sql.DB
}

// NewPostgresBotulinumModuleRepository creates a new PostgresBotulinumModuleRepository.
func NewPostgresBotulinumModuleRepository(db *sql.DB) *PostgresBotulinumModuleRepository {
	return &PostgresBotulinumModuleRepository{db: db}
}

// Create inserts a new botulinum module detail record and sets the generated ID.
func (r *PostgresBotulinumModuleRepository) Create(ctx context.Context, detail *domain.BotulinumModuleDetail) error {
	now := time.Now()
	detail.CreatedAt = now
	detail.UpdatedAt = now

	err := r.db.QueryRowContext(ctx,
		`INSERT INTO botulinum_module_details (
			module_id, product_id, batch_number, expiry_date,
			diluent, dilution_volume, resulting_concentration,
			total_units, injection_sites, notes, version,
			created_at, created_by, updated_at, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id`,
		detail.ModuleID, detail.ProductID, detail.BatchNumber, detail.ExpiryDate,
		detail.Diluent, detail.DilutionVolume, detail.ResultingConcentration,
		detail.TotalUnits, detail.InjectionSites, detail.Notes, detail.Version,
		detail.CreatedAt, detail.CreatedBy, detail.UpdatedAt, detail.UpdatedBy,
	).Scan(&detail.ID)
	if err != nil {
		return fmt.Errorf("inserting botulinum module detail: %w", err)
	}

	return nil
}

// GetByModuleID retrieves a botulinum module detail by its parent module ID.
func (r *PostgresBotulinumModuleRepository) GetByModuleID(ctx context.Context, moduleID int64) (*domain.BotulinumModuleDetail, error) {
	var d domain.BotulinumModuleDetail

	// Use a nullable byte slice for JSONB that may be NULL.
	var injectionSites *[]byte

	err := r.db.QueryRowContext(ctx,
		`SELECT id, module_id, product_id, batch_number, expiry_date,
			diluent, dilution_volume, resulting_concentration,
			total_units, injection_sites, notes, version,
			created_at, created_by, updated_at, updated_by
		FROM botulinum_module_details WHERE module_id = $1`, moduleID,
	).Scan(
		&d.ID, &d.ModuleID, &d.ProductID, &d.BatchNumber, &d.ExpiryDate,
		&d.Diluent, &d.DilutionVolume, &d.ResultingConcentration,
		&d.TotalUnits, &injectionSites, &d.Notes, &d.Version,
		&d.CreatedAt, &d.CreatedBy, &d.UpdatedAt, &d.UpdatedBy,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrModuleDetailNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("querying botulinum module detail: %w", err)
	}

	if injectionSites != nil {
		d.InjectionSites = *injectionSites
	}

	return &d, nil
}

// Update modifies a botulinum module detail using optimistic locking on the version field.
func (r *PostgresBotulinumModuleRepository) Update(ctx context.Context, detail *domain.BotulinumModuleDetail) error {
	now := time.Now()

	result, err := r.db.ExecContext(ctx,
		`UPDATE botulinum_module_details SET
			product_id = $1, batch_number = $2, expiry_date = $3,
			diluent = $4, dilution_volume = $5, resulting_concentration = $6,
			total_units = $7, injection_sites = $8, notes = $9,
			version = version + 1, updated_at = $10, updated_by = $11
		WHERE id = $12 AND version = $13`,
		detail.ProductID, detail.BatchNumber, detail.ExpiryDate,
		detail.Diluent, detail.DilutionVolume, detail.ResultingConcentration,
		detail.TotalUnits, detail.InjectionSites, detail.Notes,
		now, detail.UpdatedBy,
		detail.ID, detail.Version,
	)
	if err != nil {
		return fmt.Errorf("updating botulinum module detail: %w", err)
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
