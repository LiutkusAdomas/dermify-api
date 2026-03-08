package service_test

import (
	"context"
	"errors"
	"testing"

	"dermify-api/internal/domain"
	"dermify-api/internal/service"
	"dermify-api/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// energyTestDeps holds all mocked dependencies used by EnergyModuleService tests.
type energyTestDeps struct {
	energySvc    *service.EnergyModuleService
	sessionRepo  *testutil.MockSessionRepository
	consentRepo  *testutil.MockConsentRepository
	moduleRepo   *testutil.MockModuleRepository
	registryRepo *testutil.MockRegistryRepository
	iplRepo      *testutil.MockIPLModuleRepository
	ndyagRepo    *testutil.MockNdYAGModuleRepository
	co2Repo      *testutil.MockCO2ModuleRepository
	rfRepo       *testutil.MockRFModuleRepository
}

func newEnergyTestDeps() energyTestDeps {
	sessionRepo := &testutil.MockSessionRepository{}
	consentRepo := &testutil.MockConsentRepository{}
	moduleRepo := &testutil.MockModuleRepository{}
	registryRepo := &testutil.MockRegistryRepository{}
	iplRepo := &testutil.MockIPLModuleRepository{}
	ndyagRepo := &testutil.MockNdYAGModuleRepository{}
	co2Repo := &testutil.MockCO2ModuleRepository{}
	rfRepo := &testutil.MockRFModuleRepository{}

	sessionSvc := service.NewSessionService(sessionRepo, consentRepo, moduleRepo)
	registrySvc := service.NewRegistryService(registryRepo)
	energySvc := service.NewEnergyModuleService(sessionSvc, registrySvc, iplRepo, ndyagRepo, co2Repo, rfRepo)

	return energyTestDeps{
		energySvc:    energySvc,
		sessionRepo:  sessionRepo,
		consentRepo:  consentRepo,
		moduleRepo:   moduleRepo,
		registryRepo: registryRepo,
		iplRepo:      iplRepo,
		ndyagRepo:    ndyagRepo,
		co2Repo:      co2Repo,
		rfRepo:       rfRepo,
	}
}

// setupEditableSession configures mocks so AddModule succeeds: editable session,
// consent exists, and module Create assigns an ID.
func (d *energyTestDeps) setupEditableSession(sessionID int64, moduleID int64) {
	d.sessionRepo.GetByIDFn = func(_ context.Context, id int64) (*domain.Session, error) {
		return &domain.Session{ID: id, Status: domain.SessionStatusDraft}, nil
	}
	d.consentRepo.ExistsForSessionFn = func(_ context.Context, _ int64) (bool, error) {
		return true, nil
	}
	d.moduleRepo.NextSortOrderFn = func(_ context.Context, _ int64) (int, error) {
		return 1, nil
	}
	d.moduleRepo.CreateFn = func(_ context.Context, mod *domain.SessionModule, _ int64) error {
		mod.ID = moduleID
		return nil
	}
}

// setupIPLDevice configures the registry mock to return an IPL device with given handpieces.
func (d *energyTestDeps) setupIPLDevice(deviceID int64, handpieceIDs ...int64) {
	handpieces := make([]domain.Handpiece, len(handpieceIDs))
	for i, hpID := range handpieceIDs {
		handpieces[i] = domain.Handpiece{ID: hpID, DeviceID: deviceID, Name: "HP", Active: true}
	}

	d.registryRepo.GetDeviceByIDFn = func(_ context.Context, id int64) (*domain.Device, error) {
		if id == deviceID {
			return &domain.Device{
				ID:         deviceID,
				Name:       "Test IPL Device",
				DeviceType: domain.ModuleTypeIPL,
				Active:     true,
				Handpieces: handpieces,
			}, nil
		}
		return nil, service.ErrDeviceNotFound
	}
}

// setupDeviceWithType configures the registry mock to return a device of the specified type.
func (d *energyTestDeps) setupDeviceWithType(deviceID int64, deviceType string, handpieceIDs ...int64) {
	handpieces := make([]domain.Handpiece, len(handpieceIDs))
	for i, hpID := range handpieceIDs {
		handpieces[i] = domain.Handpiece{ID: hpID, DeviceID: deviceID, Name: "HP", Active: true}
	}

	d.registryRepo.GetDeviceByIDFn = func(_ context.Context, id int64) (*domain.Device, error) {
		if id == deviceID {
			return &domain.Device{
				ID:         deviceID,
				Name:       "Test Device",
				DeviceType: deviceType,
				Active:     true,
				Handpieces: handpieces,
			}, nil
		}
		return nil, service.ErrDeviceNotFound
	}
}

// ---------------------------------------------------------------------------
// IPL Create tests
// ---------------------------------------------------------------------------

func TestCreateIPLModule(t *testing.T) {
	deps := newEnergyTestDeps()
	deps.setupEditableSession(1, 42)
	deps.setupIPLDevice(100, 10)

	createCalled := false
	deps.iplRepo.CreateFn = func(_ context.Context, detail *domain.IPLModuleDetail) error {
		createCalled = true
		detail.ID = 99
		return nil
	}

	hpID := int64(10)
	detail := &domain.IPLModuleDetail{
		DeviceID:    100,
		HandpieceID: &hpID,
		FilterBand:  strPtr("560nm"),
	}

	result, err := deps.energySvc.CreateIPLModule(context.Background(), 1, detail, 5)

	require.NoError(t, err)
	assert.True(t, createCalled, "IPL repo Create should be called")
	assert.NotNil(t, result)
	assert.Equal(t, int64(42), result.ModuleID)
	assert.Equal(t, int64(5), result.CreatedBy)
	assert.Equal(t, int64(5), result.UpdatedBy)
	assert.Equal(t, 1, result.Version)
	assert.Equal(t, int64(99), result.ID)
}

func TestCreateIPLModule_DeviceNotFound(t *testing.T) {
	deps := newEnergyTestDeps()

	deps.registryRepo.GetDeviceByIDFn = func(_ context.Context, _ int64) (*domain.Device, error) {
		return nil, service.ErrDeviceNotFound
	}

	detail := &domain.IPLModuleDetail{DeviceID: 999}
	_, err := deps.energySvc.CreateIPLModule(context.Background(), 1, detail, 5)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrDeviceNotFound))
}

func TestCreateIPLModule_DeviceTypeMismatch(t *testing.T) {
	deps := newEnergyTestDeps()

	// Return an RF device when IPL is expected
	deps.setupDeviceWithType(100, domain.ModuleTypeRF)

	detail := &domain.IPLModuleDetail{DeviceID: 100}
	_, err := deps.energySvc.CreateIPLModule(context.Background(), 1, detail, 5)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrDeviceTypeMismatch))
}

func TestCreateIPLModule_HandpieceMismatch(t *testing.T) {
	deps := newEnergyTestDeps()

	// Device has handpiece 10, but we request handpiece 99
	deps.setupIPLDevice(100, 10)

	hpID := int64(99)
	detail := &domain.IPLModuleDetail{
		DeviceID:    100,
		HandpieceID: &hpID,
	}
	_, err := deps.energySvc.CreateIPLModule(context.Background(), 1, detail, 5)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrHandpieceMismatch))
}

func TestCreateIPLModule_ConsentRequired(t *testing.T) {
	deps := newEnergyTestDeps()
	deps.setupIPLDevice(100)

	// Session is editable but no consent
	deps.sessionRepo.GetByIDFn = func(_ context.Context, id int64) (*domain.Session, error) {
		return &domain.Session{ID: id, Status: domain.SessionStatusDraft}, nil
	}
	deps.consentRepo.ExistsForSessionFn = func(_ context.Context, _ int64) (bool, error) {
		return false, nil
	}

	detail := &domain.IPLModuleDetail{DeviceID: 100}
	_, err := deps.energySvc.CreateIPLModule(context.Background(), 1, detail, 5)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrConsentRequired))
}

func TestCreateIPLModule_SessionNotEditable(t *testing.T) {
	deps := newEnergyTestDeps()
	deps.setupIPLDevice(100)

	// Session is locked
	deps.sessionRepo.GetByIDFn = func(_ context.Context, id int64) (*domain.Session, error) {
		return &domain.Session{ID: id, Status: domain.SessionStatusLocked}, nil
	}

	detail := &domain.IPLModuleDetail{DeviceID: 100}
	_, err := deps.energySvc.CreateIPLModule(context.Background(), 1, detail, 5)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrSessionNotEditable))
}

// ---------------------------------------------------------------------------
// IPL Get tests
// ---------------------------------------------------------------------------

func TestGetIPLModule(t *testing.T) {
	deps := newEnergyTestDeps()

	expected := &domain.IPLModuleDetail{
		ID:       99,
		ModuleID: 42,
		DeviceID: 100,
		Version:  1,
	}
	deps.iplRepo.GetByModuleIDFn = func(_ context.Context, moduleID int64) (*domain.IPLModuleDetail, error) {
		assert.Equal(t, int64(42), moduleID)
		return expected, nil
	}

	result, err := deps.energySvc.GetIPLModule(context.Background(), 42)

	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestGetIPLModule_NotFound(t *testing.T) {
	deps := newEnergyTestDeps()

	deps.iplRepo.GetByModuleIDFn = func(_ context.Context, _ int64) (*domain.IPLModuleDetail, error) {
		return nil, service.ErrModuleDetailNotFound
	}

	_, err := deps.energySvc.GetIPLModule(context.Background(), 999)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrModuleDetailNotFound))
}

// ---------------------------------------------------------------------------
// IPL Update tests
// ---------------------------------------------------------------------------

func TestUpdateIPLModule(t *testing.T) {
	deps := newEnergyTestDeps()
	deps.setupIPLDevice(100, 10)

	updateCalled := false
	deps.iplRepo.UpdateFn = func(_ context.Context, detail *domain.IPLModuleDetail) error {
		updateCalled = true
		assert.Equal(t, int64(7), detail.UpdatedBy)
		return nil
	}

	hpID := int64(10)
	detail := &domain.IPLModuleDetail{
		ID:          99,
		DeviceID:    100,
		HandpieceID: &hpID,
		Version:     1,
	}
	err := deps.energySvc.UpdateIPLModule(context.Background(), detail, 7)

	require.NoError(t, err)
	assert.True(t, updateCalled, "IPL repo Update should be called")
}

func TestUpdateIPLModule_VersionConflict(t *testing.T) {
	deps := newEnergyTestDeps()
	deps.setupIPLDevice(100)

	deps.iplRepo.UpdateFn = func(_ context.Context, _ *domain.IPLModuleDetail) error {
		return service.ErrModuleDetailVersionConflict
	}

	detail := &domain.IPLModuleDetail{ID: 99, DeviceID: 100, Version: 1}
	err := deps.energySvc.UpdateIPLModule(context.Background(), detail, 7)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrModuleDetailVersionConflict))
}

// ---------------------------------------------------------------------------
// NdYAG Create test
// ---------------------------------------------------------------------------

func TestCreateNdYAGModule(t *testing.T) {
	deps := newEnergyTestDeps()
	deps.setupEditableSession(1, 50)
	deps.setupDeviceWithType(200, domain.ModuleTypeNdYAG, 20)

	createCalled := false
	deps.ndyagRepo.CreateFn = func(_ context.Context, detail *domain.NdYAGModuleDetail) error {
		createCalled = true
		detail.ID = 101
		return nil
	}

	hpID := int64(20)
	detail := &domain.NdYAGModuleDetail{
		DeviceID:    200,
		HandpieceID: &hpID,
		Wavelength:  strPtr("1064nm"),
	}

	result, err := deps.energySvc.CreateNdYAGModule(context.Background(), 1, detail, 5)

	require.NoError(t, err)
	assert.True(t, createCalled, "NdYAG repo Create should be called")
	assert.NotNil(t, result)
	assert.Equal(t, int64(50), result.ModuleID)
	assert.Equal(t, int64(101), result.ID)
	assert.Equal(t, int64(5), result.CreatedBy)
}

// ---------------------------------------------------------------------------
// CO2 Create test
// ---------------------------------------------------------------------------

func TestCreateCO2Module(t *testing.T) {
	deps := newEnergyTestDeps()
	deps.setupEditableSession(1, 60)
	deps.setupDeviceWithType(300, domain.ModuleTypeCO2)

	createCalled := false
	deps.co2Repo.CreateFn = func(_ context.Context, detail *domain.CO2ModuleDetail) error {
		createCalled = true
		detail.ID = 102
		return nil
	}

	detail := &domain.CO2ModuleDetail{
		DeviceID: 300,
		Mode:     strPtr("fractional"),
	}

	result, err := deps.energySvc.CreateCO2Module(context.Background(), 1, detail, 5)

	require.NoError(t, err)
	assert.True(t, createCalled, "CO2 repo Create should be called")
	assert.NotNil(t, result)
	assert.Equal(t, int64(60), result.ModuleID)
	assert.Equal(t, int64(102), result.ID)
}

// ---------------------------------------------------------------------------
// RF Create test
// ---------------------------------------------------------------------------

func TestCreateRFModule(t *testing.T) {
	deps := newEnergyTestDeps()
	deps.setupEditableSession(1, 70)
	deps.setupDeviceWithType(400, domain.ModuleTypeRF, 40)

	createCalled := false
	deps.rfRepo.CreateFn = func(_ context.Context, detail *domain.RFModuleDetail) error {
		createCalled = true
		detail.ID = 103
		return nil
	}

	hpID := int64(40)
	detail := &domain.RFModuleDetail{
		DeviceID:    400,
		HandpieceID: &hpID,
		RFMode:      strPtr("bipolar"),
	}

	result, err := deps.energySvc.CreateRFModule(context.Background(), 1, detail, 5)

	require.NoError(t, err)
	assert.True(t, createCalled, "RF repo Create should be called")
	assert.NotNil(t, result)
	assert.Equal(t, int64(70), result.ModuleID)
	assert.Equal(t, int64(103), result.ID)
	assert.Equal(t, int64(5), result.CreatedBy)
}

// ---------------------------------------------------------------------------
// Multiple module types in same session
// ---------------------------------------------------------------------------

func TestMultipleModuleTypes(t *testing.T) {
	deps := newEnergyTestDeps()

	// Session setup
	deps.sessionRepo.GetByIDFn = func(_ context.Context, id int64) (*domain.Session, error) {
		return &domain.Session{ID: id, Status: domain.SessionStatusDraft}, nil
	}
	deps.consentRepo.ExistsForSessionFn = func(_ context.Context, _ int64) (bool, error) {
		return true, nil
	}

	sortOrder := 0
	deps.moduleRepo.NextSortOrderFn = func(_ context.Context, _ int64) (int, error) {
		sortOrder++
		return sortOrder, nil
	}

	moduleIDCounter := int64(41)
	deps.moduleRepo.CreateFn = func(_ context.Context, mod *domain.SessionModule, _ int64) error {
		moduleIDCounter++
		mod.ID = moduleIDCounter
		return nil
	}

	// IPL device returns IPL type
	deps.registryRepo.GetDeviceByIDFn = func(_ context.Context, id int64) (*domain.Device, error) {
		switch id {
		case 100:
			return &domain.Device{
				ID: 100, Name: "IPL Device", DeviceType: domain.ModuleTypeIPL,
				Active: true, Handpieces: []domain.Handpiece{{ID: 10, DeviceID: 100}},
			}, nil
		case 400:
			return &domain.Device{
				ID: 400, Name: "RF Device", DeviceType: domain.ModuleTypeRF,
				Active: true, Handpieces: []domain.Handpiece{{ID: 40, DeviceID: 400}},
			}, nil
		default:
			return nil, service.ErrDeviceNotFound
		}
	}

	deps.iplRepo.CreateFn = func(_ context.Context, detail *domain.IPLModuleDetail) error {
		detail.ID = 201
		return nil
	}
	deps.rfRepo.CreateFn = func(_ context.Context, detail *domain.RFModuleDetail) error {
		detail.ID = 202
		return nil
	}

	// Create IPL module
	hpIPL := int64(10)
	iplDetail := &domain.IPLModuleDetail{DeviceID: 100, HandpieceID: &hpIPL}
	iplResult, err := deps.energySvc.CreateIPLModule(context.Background(), 1, iplDetail, 5)
	require.NoError(t, err)
	assert.Equal(t, int64(42), iplResult.ModuleID)
	assert.Equal(t, int64(201), iplResult.ID)

	// Create RF module in the same session
	hpRF := int64(40)
	rfDetail := &domain.RFModuleDetail{DeviceID: 400, HandpieceID: &hpRF}
	rfResult, err := deps.energySvc.CreateRFModule(context.Background(), 1, rfDetail, 5)
	require.NoError(t, err)
	assert.Equal(t, int64(43), rfResult.ModuleID)
	assert.Equal(t, int64(202), rfResult.ID)

	// Verify different module IDs (different base module rows)
	assert.NotEqual(t, iplResult.ModuleID, rfResult.ModuleID)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func strPtr(s string) *string {
	return &s
}
