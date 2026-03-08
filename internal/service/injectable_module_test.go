package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"dermify-api/internal/domain"
	"dermify-api/internal/service"
	"dermify-api/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// injectableTestDeps holds all mocked dependencies used by InjectableModuleService tests.
type injectableTestDeps struct {
	svc          *service.InjectableModuleService
	sessionRepo  *testutil.MockSessionRepository
	consentRepo  *testutil.MockConsentRepository
	moduleRepo   *testutil.MockModuleRepository
	registryRepo *testutil.MockRegistryRepository
	fillerRepo   *testutil.MockFillerModuleRepository
	botulinumRepo *testutil.MockBotulinumModuleRepository
}

func newInjectableTestDeps() injectableTestDeps {
	sessionRepo := &testutil.MockSessionRepository{}
	consentRepo := &testutil.MockConsentRepository{}
	moduleRepo := &testutil.MockModuleRepository{}
	registryRepo := &testutil.MockRegistryRepository{}
	fillerRepo := &testutil.MockFillerModuleRepository{}
	botulinumRepo := &testutil.MockBotulinumModuleRepository{}

	sessionSvc := service.NewSessionService(sessionRepo, consentRepo, moduleRepo)
	registrySvc := service.NewRegistryService(registryRepo)
	svc := service.NewInjectableModuleService(sessionSvc, registrySvc, fillerRepo, botulinumRepo)

	return injectableTestDeps{
		svc:           svc,
		sessionRepo:   sessionRepo,
		consentRepo:   consentRepo,
		moduleRepo:    moduleRepo,
		registryRepo:  registryRepo,
		fillerRepo:    fillerRepo,
		botulinumRepo: botulinumRepo,
	}
}

// setupEditableSession configures mocks so AddModule succeeds: editable session,
// consent exists, and module Create assigns an ID.
func (d *injectableTestDeps) setupEditableSession(sessionID int64, moduleID int64) {
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

// setupFillerProduct configures the registry mock to return a product with type filler.
func (d *injectableTestDeps) setupFillerProduct(productID int64) {
	d.registryRepo.GetProductByIDFn = func(_ context.Context, id int64) (*domain.Product, error) {
		if id == productID {
			return &domain.Product{
				ID:          productID,
				Name:        "Test Filler",
				ProductType: domain.ModuleTypeFiller,
				Active:      true,
			}, nil
		}
		return nil, service.ErrProductNotFound
	}
}

// setupBotulinumProduct configures the registry mock to return a product with type botulinum_toxin.
func (d *injectableTestDeps) setupBotulinumProduct(productID int64) {
	d.registryRepo.GetProductByIDFn = func(_ context.Context, id int64) (*domain.Product, error) {
		if id == productID {
			return &domain.Product{
				ID:          productID,
				Name:        "Test Botulinum",
				ProductType: domain.ModuleTypeBotulinum,
				Active:      true,
			}, nil
		}
		return nil, service.ErrProductNotFound
	}
}

// ---------------------------------------------------------------------------
// Filler Create tests
// ---------------------------------------------------------------------------

func TestCreateFillerModule_Success(t *testing.T) {
	deps := newInjectableTestDeps()
	deps.setupEditableSession(1, 42)
	deps.setupFillerProduct(100)

	createCalled := false
	deps.fillerRepo.CreateFn = func(_ context.Context, detail *domain.FillerModuleDetail) error {
		createCalled = true
		detail.ID = 99
		return nil
	}

	detail := &domain.FillerModuleDetail{
		ProductID:   100,
		BatchNumber: strPtr("LOT-001"),
	}

	result, err := deps.svc.CreateFillerModule(context.Background(), 1, detail, 5)

	require.NoError(t, err)
	assert.True(t, createCalled, "filler repo Create should be called")
	assert.NotNil(t, result)
	assert.Equal(t, int64(42), result.ModuleID)
	assert.Equal(t, int64(5), result.CreatedBy)
	assert.Equal(t, int64(5), result.UpdatedBy)
	assert.Equal(t, 1, result.Version)
	assert.Equal(t, int64(99), result.ID)
}

func TestCreateFillerModule_ProductNotFound(t *testing.T) {
	deps := newInjectableTestDeps()

	deps.registryRepo.GetProductByIDFn = func(_ context.Context, _ int64) (*domain.Product, error) {
		return nil, service.ErrProductNotFound
	}

	detail := &domain.FillerModuleDetail{ProductID: 999}
	_, err := deps.svc.CreateFillerModule(context.Background(), 1, detail, 5)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrProductNotFound))
}

func TestCreateFillerModule_ProductTypeMismatch(t *testing.T) {
	deps := newInjectableTestDeps()

	// Return a botulinum product when filler is expected.
	deps.registryRepo.GetProductByIDFn = func(_ context.Context, _ int64) (*domain.Product, error) {
		return &domain.Product{
			ID:          100,
			Name:        "Botox",
			ProductType: domain.ModuleTypeBotulinum,
			Active:      true,
		}, nil
	}

	detail := &domain.FillerModuleDetail{ProductID: 100}
	_, err := deps.svc.CreateFillerModule(context.Background(), 1, detail, 5)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrProductTypeMismatch))
}

func TestCreateFillerModule_ConsentRequired(t *testing.T) {
	deps := newInjectableTestDeps()
	deps.setupFillerProduct(100)

	// Session is editable but no consent.
	deps.sessionRepo.GetByIDFn = func(_ context.Context, id int64) (*domain.Session, error) {
		return &domain.Session{ID: id, Status: domain.SessionStatusDraft}, nil
	}
	deps.consentRepo.ExistsForSessionFn = func(_ context.Context, _ int64) (bool, error) {
		return false, nil
	}

	detail := &domain.FillerModuleDetail{ProductID: 100}
	_, err := deps.svc.CreateFillerModule(context.Background(), 1, detail, 5)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrConsentRequired))
}

func TestCreateFillerModule_SessionNotEditable(t *testing.T) {
	deps := newInjectableTestDeps()
	deps.setupFillerProduct(100)

	// Session is locked.
	deps.sessionRepo.GetByIDFn = func(_ context.Context, id int64) (*domain.Session, error) {
		return &domain.Session{ID: id, Status: domain.SessionStatusLocked}, nil
	}

	detail := &domain.FillerModuleDetail{ProductID: 100}
	_, err := deps.svc.CreateFillerModule(context.Background(), 1, detail, 5)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrSessionNotEditable))
}

// ---------------------------------------------------------------------------
// Filler Get tests
// ---------------------------------------------------------------------------

func TestGetFillerModule_Success(t *testing.T) {
	deps := newInjectableTestDeps()

	expected := &domain.FillerModuleDetail{
		ID:        99,
		ModuleID:  42,
		ProductID: 100,
		Version:   1,
	}
	deps.fillerRepo.GetByModuleIDFn = func(_ context.Context, moduleID int64) (*domain.FillerModuleDetail, error) {
		assert.Equal(t, int64(42), moduleID)
		return expected, nil
	}

	result, err := deps.svc.GetFillerModule(context.Background(), 42)

	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestGetFillerModule_NotFound(t *testing.T) {
	deps := newInjectableTestDeps()

	deps.fillerRepo.GetByModuleIDFn = func(_ context.Context, _ int64) (*domain.FillerModuleDetail, error) {
		return nil, service.ErrModuleDetailNotFound
	}

	_, err := deps.svc.GetFillerModule(context.Background(), 999)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrModuleDetailNotFound))
}

// ---------------------------------------------------------------------------
// Filler Update tests
// ---------------------------------------------------------------------------

func TestUpdateFillerModule_Success(t *testing.T) {
	deps := newInjectableTestDeps()
	deps.setupFillerProduct(100)

	updateCalled := false
	deps.fillerRepo.UpdateFn = func(_ context.Context, detail *domain.FillerModuleDetail) error {
		updateCalled = true
		assert.Equal(t, int64(7), detail.UpdatedBy)
		return nil
	}

	detail := &domain.FillerModuleDetail{
		ID:        99,
		ProductID: 100,
		Version:   1,
	}
	err := deps.svc.UpdateFillerModule(context.Background(), detail, 7)

	require.NoError(t, err)
	assert.True(t, updateCalled, "filler repo Update should be called")
}

func TestUpdateFillerModule_VersionConflict(t *testing.T) {
	deps := newInjectableTestDeps()
	deps.setupFillerProduct(100)

	deps.fillerRepo.UpdateFn = func(_ context.Context, _ *domain.FillerModuleDetail) error {
		return service.ErrModuleDetailVersionConflict
	}

	detail := &domain.FillerModuleDetail{ID: 99, ProductID: 100, Version: 1}
	err := deps.svc.UpdateFillerModule(context.Background(), detail, 7)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrModuleDetailVersionConflict))
}

// ---------------------------------------------------------------------------
// Botulinum Create tests
// ---------------------------------------------------------------------------

func TestCreateBotulinumModule_Success(t *testing.T) {
	deps := newInjectableTestDeps()
	deps.setupEditableSession(1, 50)
	deps.setupBotulinumProduct(200)

	createCalled := false
	deps.botulinumRepo.CreateFn = func(_ context.Context, detail *domain.BotulinumModuleDetail) error {
		createCalled = true
		detail.ID = 101
		return nil
	}

	sites, _ := json.Marshal([]domain.InjectionSite{
		{Site: "glabella", Units: 20},
		{Site: "frontalis", Units: 10},
	})
	detail := &domain.BotulinumModuleDetail{
		ProductID:      200,
		InjectionSites: sites,
		BatchNumber:    strPtr("BTX-001"),
	}

	result, err := deps.svc.CreateBotulinumModule(context.Background(), 1, detail, 5)

	require.NoError(t, err)
	assert.True(t, createCalled, "botulinum repo Create should be called")
	assert.NotNil(t, result)
	assert.Equal(t, int64(50), result.ModuleID)
	assert.Equal(t, int64(5), result.CreatedBy)
	assert.Equal(t, int64(5), result.UpdatedBy)
	assert.Equal(t, 1, result.Version)
	assert.Equal(t, int64(101), result.ID)
}

func TestCreateBotulinumModule_ProductTypeMismatch(t *testing.T) {
	deps := newInjectableTestDeps()

	// Return a filler product when botulinum is expected.
	deps.registryRepo.GetProductByIDFn = func(_ context.Context, _ int64) (*domain.Product, error) {
		return &domain.Product{
			ID:          100,
			Name:        "Juvederm",
			ProductType: domain.ModuleTypeFiller,
			Active:      true,
		}, nil
	}

	detail := &domain.BotulinumModuleDetail{ProductID: 100}
	_, err := deps.svc.CreateBotulinumModule(context.Background(), 1, detail, 5)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrProductTypeMismatch))
}

func TestCreateBotulinumModule_InvalidInjectionSites(t *testing.T) {
	deps := newInjectableTestDeps()
	deps.setupBotulinumProduct(200)

	// Malformed JSON for injection sites.
	detail := &domain.BotulinumModuleDetail{
		ProductID:      200,
		InjectionSites: json.RawMessage(`[{"site":"","units":10}]`), // empty site name
	}
	_, err := deps.svc.CreateBotulinumModule(context.Background(), 1, detail, 5)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrInvalidInjectionSites))
}

func TestCreateBotulinumModule_MalformedJSON(t *testing.T) {
	deps := newInjectableTestDeps()
	deps.setupBotulinumProduct(200)

	// Completely invalid JSON.
	detail := &domain.BotulinumModuleDetail{
		ProductID:      200,
		InjectionSites: json.RawMessage(`not valid json`),
	}
	_, err := deps.svc.CreateBotulinumModule(context.Background(), 1, detail, 5)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrInvalidInjectionSites))
}

// ---------------------------------------------------------------------------
// Botulinum Get tests
// ---------------------------------------------------------------------------

func TestGetBotulinumModule_Success(t *testing.T) {
	deps := newInjectableTestDeps()

	sites, _ := json.Marshal([]domain.InjectionSite{
		{Site: "glabella", Units: 20},
	})
	expected := &domain.BotulinumModuleDetail{
		ID:             101,
		ModuleID:       50,
		ProductID:      200,
		InjectionSites: sites,
		Version:        1,
	}
	deps.botulinumRepo.GetByModuleIDFn = func(_ context.Context, moduleID int64) (*domain.BotulinumModuleDetail, error) {
		assert.Equal(t, int64(50), moduleID)
		return expected, nil
	}

	result, err := deps.svc.GetBotulinumModule(context.Background(), 50)

	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestGetBotulinumModule_NotFound(t *testing.T) {
	deps := newInjectableTestDeps()

	deps.botulinumRepo.GetByModuleIDFn = func(_ context.Context, _ int64) (*domain.BotulinumModuleDetail, error) {
		return nil, service.ErrModuleDetailNotFound
	}

	_, err := deps.svc.GetBotulinumModule(context.Background(), 999)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrModuleDetailNotFound))
}

// ---------------------------------------------------------------------------
// Botulinum Update tests
// ---------------------------------------------------------------------------

func TestUpdateBotulinumModule_Success(t *testing.T) {
	deps := newInjectableTestDeps()
	deps.setupBotulinumProduct(200)

	updateCalled := false
	deps.botulinumRepo.UpdateFn = func(_ context.Context, detail *domain.BotulinumModuleDetail) error {
		updateCalled = true
		assert.Equal(t, int64(7), detail.UpdatedBy)
		return nil
	}

	sites, _ := json.Marshal([]domain.InjectionSite{
		{Site: "glabella", Units: 25},
	})
	detail := &domain.BotulinumModuleDetail{
		ID:             101,
		ProductID:      200,
		InjectionSites: sites,
		Version:        1,
	}
	err := deps.svc.UpdateBotulinumModule(context.Background(), detail, 7)

	require.NoError(t, err)
	assert.True(t, updateCalled, "botulinum repo Update should be called")
}

func TestUpdateBotulinumModule_VersionConflict(t *testing.T) {
	deps := newInjectableTestDeps()
	deps.setupBotulinumProduct(200)

	deps.botulinumRepo.UpdateFn = func(_ context.Context, _ *domain.BotulinumModuleDetail) error {
		return service.ErrModuleDetailVersionConflict
	}

	detail := &domain.BotulinumModuleDetail{ID: 101, ProductID: 200, Version: 1}
	err := deps.svc.UpdateBotulinumModule(context.Background(), detail, 7)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrModuleDetailVersionConflict))
}
