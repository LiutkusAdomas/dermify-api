package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"dermify-api/internal/domain"
	"dermify-api/internal/service"
)

// PostgresRegistryRepository implements service.RegistryRepository using PostgreSQL.
type PostgresRegistryRepository struct {
	db *sql.DB
}

// NewPostgresRegistryRepository creates a new PostgresRegistryRepository.
func NewPostgresRegistryRepository(db *sql.DB) *PostgresRegistryRepository {
	return &PostgresRegistryRepository{db: db}
}

// ListDevices returns all active devices, optionally filtered by device type.
// Handpieces are NOT loaded for list performance.
func (r *PostgresRegistryRepository) ListDevices(ctx context.Context, deviceType string) ([]domain.Device, error) {
	query := `SELECT id, name, manufacturer, model, device_type, active, created_at
		FROM devices WHERE active = true`
	args := []interface{}{}

	if deviceType != "" {
		query += " AND device_type = $1"
		args = append(args, deviceType)
	}

	query += " ORDER BY name ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying devices: %w", err)
	}
	defer rows.Close()

	return scanDevices(rows)
}

// GetDeviceByID returns a single device with its associated handpieces.
func (r *PostgresRegistryRepository) GetDeviceByID(ctx context.Context, id int64) (*domain.Device, error) {
	var d domain.Device

	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, manufacturer, model, device_type, active, created_at
		FROM devices WHERE id = $1`, id,
	).Scan(&d.ID, &d.Name, &d.Manufacturer, &d.Model, &d.DeviceType, &d.Active, &d.CreatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrDeviceNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("querying device: %w", err)
	}

	handpieces, err := r.loadHandpieces(ctx, id)
	if err != nil {
		return nil, err
	}

	d.Handpieces = handpieces

	return &d, nil
}

// ListProducts returns all active products, optionally filtered by product type.
func (r *PostgresRegistryRepository) ListProducts(ctx context.Context, productType string) ([]domain.Product, error) {
	query := `SELECT id, name, manufacturer, product_type, concentration, active, created_at
		FROM products WHERE active = true`
	args := []interface{}{}

	if productType != "" {
		query += " AND product_type = $1"
		args = append(args, productType)
	}

	query += " ORDER BY name ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying products: %w", err)
	}
	defer rows.Close()

	return scanProducts(rows)
}

// GetProductByID returns a single product by ID.
func (r *PostgresRegistryRepository) GetProductByID(ctx context.Context, id int64) (*domain.Product, error) {
	var p domain.Product

	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, manufacturer, product_type, concentration, active, created_at
		FROM products WHERE id = $1`, id,
	).Scan(&p.ID, &p.Name, &p.Manufacturer, &p.ProductType, &p.Concentration, &p.Active, &p.CreatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrProductNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("querying product: %w", err)
	}

	return &p, nil
}

// ListIndicationCodes returns indication codes, optionally filtered by module type.
func (r *PostgresRegistryRepository) ListIndicationCodes(ctx context.Context, moduleType string) ([]domain.IndicationCode, error) {
	query := "SELECT id, code, name, module_type, active FROM indication_codes WHERE active = true"
	args := []interface{}{}

	if moduleType != "" {
		query += " AND module_type = $1"
		args = append(args, moduleType)
	}

	query += " ORDER BY code ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying indication codes: %w", err)
	}
	defer rows.Close()

	return scanIndicationCodes(rows)
}

// ListClinicalEndpoints returns clinical endpoints, optionally filtered by module type.
func (r *PostgresRegistryRepository) ListClinicalEndpoints(ctx context.Context, moduleType string) ([]domain.ClinicalEndpoint, error) {
	query := "SELECT id, code, name, module_type, active FROM clinical_endpoints WHERE active = true"
	args := []interface{}{}

	if moduleType != "" {
		query += " AND module_type = $1"
		args = append(args, moduleType)
	}

	query += " ORDER BY code ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying clinical endpoints: %w", err)
	}
	defer rows.Close()

	return scanClinicalEndpoints(rows)
}

// loadHandpieces loads all handpieces for a given device.
func (r *PostgresRegistryRepository) loadHandpieces(ctx context.Context, deviceID int64) ([]domain.Handpiece, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, device_id, name, active, created_at
		FROM handpieces WHERE device_id = $1 AND active = true
		ORDER BY name ASC`, deviceID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying handpieces: %w", err)
	}
	defer rows.Close()

	handpieces := []domain.Handpiece{}

	for rows.Next() {
		var h domain.Handpiece
		err := rows.Scan(&h.ID, &h.DeviceID, &h.Name, &h.Active, &h.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scanning handpiece row: %w", err)
		}
		handpieces = append(handpieces, h)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating handpiece rows: %w", err)
	}

	return handpieces, nil
}

// scanDevices scans rows into a Device slice.
func scanDevices(rows *sql.Rows) ([]domain.Device, error) {
	devices := []domain.Device{}

	for rows.Next() {
		var d domain.Device
		err := rows.Scan(&d.ID, &d.Name, &d.Manufacturer, &d.Model, &d.DeviceType, &d.Active, &d.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scanning device row: %w", err)
		}
		devices = append(devices, d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating device rows: %w", err)
	}

	return devices, nil
}

// scanProducts scans rows into a Product slice.
func scanProducts(rows *sql.Rows) ([]domain.Product, error) {
	products := []domain.Product{}

	for rows.Next() {
		var p domain.Product
		err := rows.Scan(&p.ID, &p.Name, &p.Manufacturer, &p.ProductType, &p.Concentration, &p.Active, &p.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scanning product row: %w", err)
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating product rows: %w", err)
	}

	return products, nil
}

// scanIndicationCodes scans rows into an IndicationCode slice.
func scanIndicationCodes(rows *sql.Rows) ([]domain.IndicationCode, error) {
	codes := []domain.IndicationCode{}

	for rows.Next() {
		var c domain.IndicationCode
		err := rows.Scan(&c.ID, &c.Code, &c.Name, &c.ModuleType, &c.Active)
		if err != nil {
			return nil, fmt.Errorf("scanning indication code row: %w", err)
		}
		codes = append(codes, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating indication code rows: %w", err)
	}

	return codes, nil
}

// scanClinicalEndpoints scans rows into a ClinicalEndpoint slice.
func scanClinicalEndpoints(rows *sql.Rows) ([]domain.ClinicalEndpoint, error) {
	endpoints := []domain.ClinicalEndpoint{}

	for rows.Next() {
		var e domain.ClinicalEndpoint
		err := rows.Scan(&e.ID, &e.Code, &e.Name, &e.ModuleType, &e.Active)
		if err != nil {
			return nil, fmt.Errorf("scanning clinical endpoint row: %w", err)
		}
		endpoints = append(endpoints, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating clinical endpoint rows: %w", err)
	}

	return endpoints, nil
}
