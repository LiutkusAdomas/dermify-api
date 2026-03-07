//go:build ignore
// +build ignore

// Remove //go:build ignore after Plan 01-01 creates service interfaces.

package testutil

import (
	"context"
	"time"
)

// MockRoleRepository is a test double for the RoleRepository interface.
type MockRoleRepository struct {
	UpdateUserRoleFn func(ctx context.Context, userID int64, role string) error
	GetUserRoleFn    func(ctx context.Context, userID int64) (string, error)
	CountUsersFn     func(ctx context.Context) (int64, error)
}

// UpdateUserRole delegates to UpdateUserRoleFn if set, otherwise returns nil.
func (m *MockRoleRepository) UpdateUserRole(ctx context.Context, userID int64, role string) error {
	if m.UpdateUserRoleFn != nil {
		return m.UpdateUserRoleFn(ctx, userID, role)
	}
	return nil
}

// GetUserRole delegates to GetUserRoleFn if set, otherwise returns empty string and nil.
func (m *MockRoleRepository) GetUserRole(ctx context.Context, userID int64) (string, error) {
	if m.GetUserRoleFn != nil {
		return m.GetUserRoleFn(ctx, userID)
	}
	return "", nil
}

// CountUsers delegates to CountUsersFn if set, otherwise returns 0 and nil.
func (m *MockRoleRepository) CountUsers(ctx context.Context) (int64, error) {
	if m.CountUsersFn != nil {
		return m.CountUsersFn(ctx)
	}
	return 0, nil
}

// PatientListItem represents a patient in list results with session metadata.
type PatientListItem struct {
	ID                int64
	FirstName         string
	LastName          string
	DateOfBirth       time.Time
	Sex               string
	Phone             string
	Email             string
	ExternalReference string
	Version           int
	CreatedAt         time.Time
	CreatedBy         int64
	UpdatedAt         time.Time
	UpdatedBy         int64
	SessionCount      int
	LastSessionDate   *time.Time
}

// Patient represents a patient record for mock operations.
type Patient struct {
	ID                int64
	FirstName         string
	LastName          string
	DateOfBirth       time.Time
	Sex               string
	Phone             string
	Email             string
	ExternalReference string
	Version           int
	CreatedAt         time.Time
	CreatedBy         int64
	UpdatedAt         time.Time
	UpdatedBy         int64
}

// PatientFilter defines filtering and pagination options for patient listing.
type PatientFilter struct {
	Search  string
	Page    int
	PerPage int
}

// PatientListResult holds paginated patient results.
type PatientListResult struct {
	Patients []PatientListItem
	Total    int
}

// SessionSummary represents a summary of a patient session.
type SessionSummary struct {
	ID        int64
	CreatedAt time.Time
	CreatedBy int64
}

// MockPatientRepository is a test double for the PatientRepository interface.
type MockPatientRepository struct {
	CreateFn            func(ctx context.Context, patient *Patient) error
	GetByIDFn           func(ctx context.Context, id int64) (*Patient, error)
	UpdateFn            func(ctx context.Context, patient *Patient) error
	ListFn              func(ctx context.Context, filter PatientFilter) (*PatientListResult, error)
	GetSessionHistoryFn func(ctx context.Context, patientID int64) ([]SessionSummary, error)
}

// Create delegates to CreateFn if set, otherwise returns nil.
func (m *MockPatientRepository) Create(ctx context.Context, patient *Patient) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, patient)
	}
	return nil
}

// GetByID delegates to GetByIDFn if set, otherwise returns nil and nil.
func (m *MockPatientRepository) GetByID(ctx context.Context, id int64) (*Patient, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, nil
}

// Update delegates to UpdateFn if set, otherwise returns nil.
func (m *MockPatientRepository) Update(ctx context.Context, patient *Patient) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, patient)
	}
	return nil
}

// List delegates to ListFn if set, otherwise returns empty result.
func (m *MockPatientRepository) List(ctx context.Context, filter PatientFilter) (*PatientListResult, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, filter)
	}
	return &PatientListResult{Patients: []PatientListItem{}, Total: 0}, nil
}

// GetSessionHistory delegates to GetSessionHistoryFn if set, otherwise returns empty slice.
func (m *MockPatientRepository) GetSessionHistory(ctx context.Context, patientID int64) ([]SessionSummary, error) {
	if m.GetSessionHistoryFn != nil {
		return m.GetSessionHistoryFn(ctx, patientID)
	}
	return []SessionSummary{}, nil
}

// Device represents a device in the registry.
type Device struct {
	ID           int64
	Name         string
	Manufacturer string
	Model        string
	DeviceType   string
	Active       bool
	CreatedAt    time.Time
}

// Handpiece represents a handpiece attached to a device.
type Handpiece struct {
	ID        int64
	DeviceID  int64
	Name      string
	Active    bool
	CreatedAt time.Time
}

// Product represents a product in the registry.
type Product struct {
	ID            int64
	Name          string
	Manufacturer  string
	ProductType   string
	Concentration string
	Active        bool
	CreatedAt     time.Time
}

// IndicationCode represents a clinical indication code.
type IndicationCode struct {
	ID         int64
	Code       string
	Name       string
	ModuleType string
	Active     bool
}

// ClinicalEndpoint represents a clinical endpoint.
type ClinicalEndpoint struct {
	ID         int64
	Code       string
	Name       string
	ModuleType string
	Active     bool
}

// MockRegistryRepository is a test double for the RegistryRepository interface.
type MockRegistryRepository struct {
	ListDevicesFn            func(ctx context.Context) ([]Device, error)
	GetDeviceByIDFn          func(ctx context.Context, id int64) (*Device, error)
	ListProductsFn           func(ctx context.Context) ([]Product, error)
	GetProductByIDFn         func(ctx context.Context, id int64) (*Product, error)
	ListIndicationCodesFn    func(ctx context.Context) ([]IndicationCode, error)
	ListClinicalEndpointsFn  func(ctx context.Context) ([]ClinicalEndpoint, error)
}

// ListDevices delegates to ListDevicesFn if set, otherwise returns empty slice.
func (m *MockRegistryRepository) ListDevices(ctx context.Context) ([]Device, error) {
	if m.ListDevicesFn != nil {
		return m.ListDevicesFn(ctx)
	}
	return []Device{}, nil
}

// GetDeviceByID delegates to GetDeviceByIDFn if set, otherwise returns nil and nil.
func (m *MockRegistryRepository) GetDeviceByID(ctx context.Context, id int64) (*Device, error) {
	if m.GetDeviceByIDFn != nil {
		return m.GetDeviceByIDFn(ctx, id)
	}
	return nil, nil
}

// ListProducts delegates to ListProductsFn if set, otherwise returns empty slice.
func (m *MockRegistryRepository) ListProducts(ctx context.Context) ([]Product, error) {
	if m.ListProductsFn != nil {
		return m.ListProductsFn(ctx)
	}
	return []Product{}, nil
}

// GetProductByID delegates to GetProductByIDFn if set, otherwise returns nil and nil.
func (m *MockRegistryRepository) GetProductByID(ctx context.Context, id int64) (*Product, error) {
	if m.GetProductByIDFn != nil {
		return m.GetProductByIDFn(ctx, id)
	}
	return nil, nil
}

// ListIndicationCodes delegates to ListIndicationCodesFn if set, otherwise returns empty slice.
func (m *MockRegistryRepository) ListIndicationCodes(ctx context.Context) ([]IndicationCode, error) {
	if m.ListIndicationCodesFn != nil {
		return m.ListIndicationCodesFn(ctx)
	}
	return []IndicationCode{}, nil
}

// ListClinicalEndpoints delegates to ListClinicalEndpointsFn if set, otherwise returns empty slice.
func (m *MockRegistryRepository) ListClinicalEndpoints(ctx context.Context) ([]ClinicalEndpoint, error) {
	if m.ListClinicalEndpointsFn != nil {
		return m.ListClinicalEndpointsFn(ctx)
	}
	return []ClinicalEndpoint{}, nil
}
