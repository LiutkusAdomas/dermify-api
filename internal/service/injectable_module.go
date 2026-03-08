package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"dermify-api/internal/domain"
)

// Sentinel errors for injectable module operations.
var (
	ErrProductTypeMismatch  = errors.New("product type does not match module type") //nolint:gochecknoglobals // sentinel error
	ErrInvalidInjectionSites = errors.New("invalid injection sites data")           //nolint:gochecknoglobals // sentinel error
)

// FillerModuleRepository defines the data access contract for filler module details.
type FillerModuleRepository interface {
	// Create inserts a new filler module detail record.
	Create(ctx context.Context, detail *domain.FillerModuleDetail) error
	// GetByModuleID retrieves a filler module detail by its parent module ID.
	GetByModuleID(ctx context.Context, moduleID int64) (*domain.FillerModuleDetail, error)
	// Update modifies a filler module detail using optimistic locking.
	Update(ctx context.Context, detail *domain.FillerModuleDetail) error
}

// BotulinumModuleRepository defines the data access contract for botulinum module details.
type BotulinumModuleRepository interface {
	// Create inserts a new botulinum module detail record.
	Create(ctx context.Context, detail *domain.BotulinumModuleDetail) error
	// GetByModuleID retrieves a botulinum module detail by its parent module ID.
	GetByModuleID(ctx context.Context, moduleID int64) (*domain.BotulinumModuleDetail, error)
	// Update modifies a botulinum module detail using optimistic locking.
	Update(ctx context.Context, detail *domain.BotulinumModuleDetail) error
}

// InjectableModuleService handles business logic for injectable treatment
// module details (filler and botulinum toxin).
type InjectableModuleService struct {
	sessionSvc   *SessionService
	registrySvc  *RegistryService
	fillerRepo   FillerModuleRepository
	botulinumRepo BotulinumModuleRepository
}

// NewInjectableModuleService creates a new InjectableModuleService with the
// given dependencies.
func NewInjectableModuleService(
	sessionSvc *SessionService,
	registrySvc *RegistryService,
	fillerRepo FillerModuleRepository,
	botulinumRepo BotulinumModuleRepository,
) *InjectableModuleService {
	return &InjectableModuleService{
		sessionSvc:    sessionSvc,
		registrySvc:   registrySvc,
		fillerRepo:    fillerRepo,
		botulinumRepo: botulinumRepo,
	}
}

// validateProductForModule checks that the product exists and matches the
// expected product type for the module being created or updated.
func (s *InjectableModuleService) validateProductForModule(ctx context.Context, productID int64, expectedProductType string) error {
	product, err := s.registrySvc.GetProductByID(ctx, productID)
	if err != nil {
		return fmt.Errorf("validating product: %w", err)
	}

	if product.ProductType != expectedProductType {
		return ErrProductTypeMismatch
	}

	return nil
}

// validateInjectionSites validates that the injection sites JSON data is
// well-formed and each entry has a non-empty site name and non-negative units.
func validateInjectionSites(sites json.RawMessage) error {
	if len(sites) == 0 {
		return nil
	}

	var parsed []domain.InjectionSite
	if err := json.Unmarshal(sites, &parsed); err != nil {
		return ErrInvalidInjectionSites
	}

	for _, site := range parsed {
		if site.Site == "" {
			return ErrInvalidInjectionSites
		}

		if site.Units < 0 {
			return ErrInvalidInjectionSites
		}
	}

	return nil
}

// ---------------------------------------------------------------------------
// Filler methods
// ---------------------------------------------------------------------------

// CreateFillerModule validates the product, creates the base session module
// row (enforcing consent gate and editability), and inserts the filler detail
// record.
func (s *InjectableModuleService) CreateFillerModule(ctx context.Context, sessionID int64, detail *domain.FillerModuleDetail, userID int64) (*domain.FillerModuleDetail, error) {
	if err := s.validateProductForModule(ctx, detail.ProductID, domain.ModuleTypeFiller); err != nil {
		return nil, err
	}

	baseModule, err := s.sessionSvc.AddModule(ctx, sessionID, domain.ModuleTypeFiller, userID)
	if err != nil {
		return nil, fmt.Errorf("creating base module: %w", err)
	}

	detail.ModuleID = baseModule.ID
	detail.CreatedBy = userID
	detail.UpdatedBy = userID
	detail.Version = 1

	if err := s.fillerRepo.Create(ctx, detail); err != nil {
		return nil, fmt.Errorf("creating filler detail: %w", err)
	}

	return detail, nil
}

// GetFillerModule retrieves a filler module detail by its parent module ID.
func (s *InjectableModuleService) GetFillerModule(ctx context.Context, moduleID int64) (*domain.FillerModuleDetail, error) {
	return s.fillerRepo.GetByModuleID(ctx, moduleID)
}

// UpdateFillerModule validates the product (if set) and updates the filler
// detail record.
func (s *InjectableModuleService) UpdateFillerModule(ctx context.Context, detail *domain.FillerModuleDetail, userID int64) error {
	detail.UpdatedBy = userID

	if detail.ProductID != 0 {
		if err := s.validateProductForModule(ctx, detail.ProductID, domain.ModuleTypeFiller); err != nil {
			return err
		}
	}

	return s.fillerRepo.Update(ctx, detail)
}

// ---------------------------------------------------------------------------
// Botulinum methods
// ---------------------------------------------------------------------------

// CreateBotulinumModule validates the product and injection sites, creates the
// base session module row, and inserts the botulinum detail record.
func (s *InjectableModuleService) CreateBotulinumModule(ctx context.Context, sessionID int64, detail *domain.BotulinumModuleDetail, userID int64) (*domain.BotulinumModuleDetail, error) {
	if err := s.validateProductForModule(ctx, detail.ProductID, domain.ModuleTypeBotulinum); err != nil {
		return nil, err
	}

	if err := validateInjectionSites(detail.InjectionSites); err != nil {
		return nil, err
	}

	baseModule, err := s.sessionSvc.AddModule(ctx, sessionID, domain.ModuleTypeBotulinum, userID)
	if err != nil {
		return nil, fmt.Errorf("creating base module: %w", err)
	}

	detail.ModuleID = baseModule.ID
	detail.CreatedBy = userID
	detail.UpdatedBy = userID
	detail.Version = 1

	if err := s.botulinumRepo.Create(ctx, detail); err != nil {
		return nil, fmt.Errorf("creating botulinum detail: %w", err)
	}

	return detail, nil
}

// GetBotulinumModule retrieves a botulinum module detail by its parent module ID.
func (s *InjectableModuleService) GetBotulinumModule(ctx context.Context, moduleID int64) (*domain.BotulinumModuleDetail, error) {
	return s.botulinumRepo.GetByModuleID(ctx, moduleID)
}

// UpdateBotulinumModule validates the product (if set) and injection sites
// (if non-nil), and updates the botulinum detail record.
func (s *InjectableModuleService) UpdateBotulinumModule(ctx context.Context, detail *domain.BotulinumModuleDetail, userID int64) error {
	detail.UpdatedBy = userID

	if detail.ProductID != 0 {
		if err := s.validateProductForModule(ctx, detail.ProductID, domain.ModuleTypeBotulinum); err != nil {
			return err
		}
	}

	if len(detail.InjectionSites) > 0 {
		if err := validateInjectionSites(detail.InjectionSites); err != nil {
			return err
		}
	}

	return s.botulinumRepo.Update(ctx, detail)
}
