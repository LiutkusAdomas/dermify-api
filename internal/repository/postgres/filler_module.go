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

// PostgresFillerModuleRepository implements service.FillerModuleRepository using PostgreSQL.
type PostgresFillerModuleRepository struct {
	db *sql.DB
}

// NewPostgresFillerModuleRepository creates a new PostgresFillerModuleRepository.
func NewPostgresFillerModuleRepository(db *sql.DB) *PostgresFillerModuleRepository {
	return &PostgresFillerModuleRepository{db: db}
}

// Create inserts a new filler module detail record and sets the generated ID.
func (r *PostgresFillerModuleRepository) Create(ctx context.Context, detail *domain.FillerModuleDetail) error {
	now := time.Now()
	detail.CreatedAt = now
	detail.UpdatedAt = now

	err := r.db.QueryRowContext(ctx,
		`INSERT INTO filler_module_details (
			module_id, product_id, batch_number, expiry_date,
			syringe_volume, total_volume, needle_type, injection_plane,
			anatomical_sites, endpoint, notes, version,
			created_at, created_by, updated_at, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id`,
		detail.ModuleID, detail.ProductID, detail.BatchNumber, detail.ExpiryDate,
		detail.SyringeVolume, detail.TotalVolume, detail.NeedleType, detail.InjectionPlane,
		detail.AnatomicalSites, detail.Endpoint, detail.Notes, detail.Version,
		detail.CreatedAt, detail.CreatedBy, detail.UpdatedAt, detail.UpdatedBy,
	).Scan(&detail.ID)
	if err != nil {
		return fmt.Errorf("inserting filler module detail: %w", err)
	}

	return nil
}

// GetByModuleID retrieves a filler module detail by its parent module ID.
func (r *PostgresFillerModuleRepository) GetByModuleID(ctx context.Context, moduleID int64) (*domain.FillerModuleDetail, error) {
	var d domain.FillerModuleDetail

	err := r.db.QueryRowContext(ctx,
		`SELECT id, module_id, product_id, batch_number, expiry_date,
			syringe_volume, total_volume, needle_type, injection_plane,
			anatomical_sites, endpoint, notes, version,
			created_at, created_by, updated_at, updated_by
		FROM filler_module_details WHERE module_id = $1`, moduleID,
	).Scan(
		&d.ID, &d.ModuleID, &d.ProductID, &d.BatchNumber, &d.ExpiryDate,
		&d.SyringeVolume, &d.TotalVolume, &d.NeedleType, &d.InjectionPlane,
		&d.AnatomicalSites, &d.Endpoint, &d.Notes, &d.Version,
		&d.CreatedAt, &d.CreatedBy, &d.UpdatedAt, &d.UpdatedBy,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrModuleDetailNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("querying filler module detail: %w", err)
	}

	return &d, nil
}

// Update modifies a filler module detail using optimistic locking on the version field.
func (r *PostgresFillerModuleRepository) Update(ctx context.Context, detail *domain.FillerModuleDetail) error {
	now := time.Now()

	result, err := r.db.ExecContext(ctx,
		`UPDATE filler_module_details SET
			product_id = $1, batch_number = $2, expiry_date = $3,
			syringe_volume = $4, total_volume = $5, needle_type = $6,
			injection_plane = $7, anatomical_sites = $8, endpoint = $9,
			notes = $10, version = version + 1,
			updated_at = $11, updated_by = $12
		WHERE id = $13 AND version = $14`,
		detail.ProductID, detail.BatchNumber, detail.ExpiryDate,
		detail.SyringeVolume, detail.TotalVolume, detail.NeedleType,
		detail.InjectionPlane, detail.AnatomicalSites, detail.Endpoint,
		detail.Notes, now, detail.UpdatedBy,
		detail.ID, detail.Version,
	)
	if err != nil {
		return fmt.Errorf("updating filler module detail: %w", err)
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
