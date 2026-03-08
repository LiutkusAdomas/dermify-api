package service

import (
	"context"
	"errors"
	"fmt"

	"dermify-api/internal/domain"
)

// Sentinel errors for energy module operations.
var (
	ErrModuleDetailNotFound      = errors.New("module detail not found")       //nolint:gochecknoglobals // sentinel error
	ErrModuleDetailVersionConflict = errors.New("module detail version conflict") //nolint:gochecknoglobals // sentinel error
	ErrDeviceTypeMismatch        = errors.New("device type does not match module type") //nolint:gochecknoglobals // sentinel error
	ErrHandpieceMismatch         = errors.New("handpiece does not belong to device")    //nolint:gochecknoglobals // sentinel error
	ErrInvalidModuleData         = errors.New("invalid module data")                     //nolint:gochecknoglobals // sentinel error
)

// IPLModuleRepository defines the data access contract for IPL module details.
type IPLModuleRepository interface {
	// Create inserts a new IPL module detail record.
	Create(ctx context.Context, detail *domain.IPLModuleDetail) error
	// GetByModuleID retrieves an IPL module detail by its parent module ID.
	GetByModuleID(ctx context.Context, moduleID int64) (*domain.IPLModuleDetail, error)
	// Update modifies an IPL module detail using optimistic locking.
	Update(ctx context.Context, detail *domain.IPLModuleDetail) error
}

// NdYAGModuleRepository defines the data access contract for Nd:YAG module details.
type NdYAGModuleRepository interface {
	// Create inserts a new Nd:YAG module detail record.
	Create(ctx context.Context, detail *domain.NdYAGModuleDetail) error
	// GetByModuleID retrieves an Nd:YAG module detail by its parent module ID.
	GetByModuleID(ctx context.Context, moduleID int64) (*domain.NdYAGModuleDetail, error)
	// Update modifies an Nd:YAG module detail using optimistic locking.
	Update(ctx context.Context, detail *domain.NdYAGModuleDetail) error
}

// CO2ModuleRepository defines the data access contract for CO2 module details.
type CO2ModuleRepository interface {
	// Create inserts a new CO2 module detail record.
	Create(ctx context.Context, detail *domain.CO2ModuleDetail) error
	// GetByModuleID retrieves a CO2 module detail by its parent module ID.
	GetByModuleID(ctx context.Context, moduleID int64) (*domain.CO2ModuleDetail, error)
	// Update modifies a CO2 module detail using optimistic locking.
	Update(ctx context.Context, detail *domain.CO2ModuleDetail) error
}

// RFModuleRepository defines the data access contract for RF module details.
type RFModuleRepository interface {
	// Create inserts a new RF module detail record.
	Create(ctx context.Context, detail *domain.RFModuleDetail) error
	// GetByModuleID retrieves an RF module detail by its parent module ID.
	GetByModuleID(ctx context.Context, moduleID int64) (*domain.RFModuleDetail, error)
	// Update modifies an RF module detail using optimistic locking.
	Update(ctx context.Context, detail *domain.RFModuleDetail) error
}

// EnergyModuleService handles business logic for energy-based treatment
// module details (IPL, Nd:YAG, CO2, RF).
type EnergyModuleService struct {
	sessionSvc  *SessionService
	registrySvc *RegistryService
	iplRepo     IPLModuleRepository
	ndyagRepo   NdYAGModuleRepository
	co2Repo     CO2ModuleRepository
	rfRepo      RFModuleRepository
}

// NewEnergyModuleService creates a new EnergyModuleService with the given
// dependencies.
func NewEnergyModuleService(
	sessionSvc *SessionService,
	registrySvc *RegistryService,
	iplRepo IPLModuleRepository,
	ndyagRepo NdYAGModuleRepository,
	co2Repo CO2ModuleRepository,
	rfRepo RFModuleRepository,
) *EnergyModuleService {
	return &EnergyModuleService{
		sessionSvc:  sessionSvc,
		registrySvc: registrySvc,
		iplRepo:     iplRepo,
		ndyagRepo:   ndyagRepo,
		co2Repo:     co2Repo,
		rfRepo:      rfRepo,
	}
}

// validateDeviceForModule checks that the device exists, matches the expected
// type, and (if provided) that the handpiece belongs to the device.
func (s *EnergyModuleService) validateDeviceForModule(ctx context.Context, deviceID int64, handpieceID *int64, expectedDeviceType string) error {
	device, err := s.registrySvc.GetDeviceByID(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("validating device: %w", err)
	}

	if device.DeviceType != expectedDeviceType {
		return ErrDeviceTypeMismatch
	}

	if handpieceID != nil {
		found := false
		for _, hp := range device.Handpieces {
			if hp.ID == *handpieceID {
				found = true
				break
			}
		}

		if !found {
			return ErrHandpieceMismatch
		}
	}

	return nil
}

// ---------------------------------------------------------------------------
// IPL methods
// ---------------------------------------------------------------------------

// CreateIPLModule validates the device, creates the base session module row
// (enforcing consent gate and editability), and inserts the IPL detail record.
func (s *EnergyModuleService) CreateIPLModule(ctx context.Context, sessionID int64, detail *domain.IPLModuleDetail, userID int64) (*domain.IPLModuleDetail, error) {
	if err := s.validateDeviceForModule(ctx, detail.DeviceID, detail.HandpieceID, domain.ModuleTypeIPL); err != nil {
		return nil, err
	}

	baseModule, err := s.sessionSvc.AddModule(ctx, sessionID, domain.ModuleTypeIPL, userID)
	if err != nil {
		return nil, fmt.Errorf("creating base module: %w", err)
	}

	detail.ModuleID = baseModule.ID
	detail.CreatedBy = userID
	detail.UpdatedBy = userID
	detail.Version = 1

	if err := s.iplRepo.Create(ctx, detail); err != nil {
		return nil, fmt.Errorf("creating IPL detail: %w", err)
	}

	return detail, nil
}

// GetIPLModule retrieves an IPL module detail by its parent module ID.
func (s *EnergyModuleService) GetIPLModule(ctx context.Context, moduleID int64) (*domain.IPLModuleDetail, error) {
	return s.iplRepo.GetByModuleID(ctx, moduleID)
}

// UpdateIPLModule validates the device (if set) and updates the IPL detail record.
func (s *EnergyModuleService) UpdateIPLModule(ctx context.Context, detail *domain.IPLModuleDetail, userID int64) error {
	detail.UpdatedBy = userID

	if detail.DeviceID != 0 {
		if err := s.validateDeviceForModule(ctx, detail.DeviceID, detail.HandpieceID, domain.ModuleTypeIPL); err != nil {
			return err
		}
	}

	return s.iplRepo.Update(ctx, detail)
}

// ---------------------------------------------------------------------------
// Nd:YAG methods
// ---------------------------------------------------------------------------

// CreateNdYAGModule validates the device, creates the base session module row,
// and inserts the Nd:YAG detail record.
func (s *EnergyModuleService) CreateNdYAGModule(ctx context.Context, sessionID int64, detail *domain.NdYAGModuleDetail, userID int64) (*domain.NdYAGModuleDetail, error) {
	if err := s.validateDeviceForModule(ctx, detail.DeviceID, detail.HandpieceID, domain.ModuleTypeNdYAG); err != nil {
		return nil, err
	}

	baseModule, err := s.sessionSvc.AddModule(ctx, sessionID, domain.ModuleTypeNdYAG, userID)
	if err != nil {
		return nil, fmt.Errorf("creating base module: %w", err)
	}

	detail.ModuleID = baseModule.ID
	detail.CreatedBy = userID
	detail.UpdatedBy = userID
	detail.Version = 1

	if err := s.ndyagRepo.Create(ctx, detail); err != nil {
		return nil, fmt.Errorf("creating NdYAG detail: %w", err)
	}

	return detail, nil
}

// GetNdYAGModule retrieves an Nd:YAG module detail by its parent module ID.
func (s *EnergyModuleService) GetNdYAGModule(ctx context.Context, moduleID int64) (*domain.NdYAGModuleDetail, error) {
	return s.ndyagRepo.GetByModuleID(ctx, moduleID)
}

// UpdateNdYAGModule validates the device (if set) and updates the Nd:YAG detail record.
func (s *EnergyModuleService) UpdateNdYAGModule(ctx context.Context, detail *domain.NdYAGModuleDetail, userID int64) error {
	detail.UpdatedBy = userID

	if detail.DeviceID != 0 {
		if err := s.validateDeviceForModule(ctx, detail.DeviceID, detail.HandpieceID, domain.ModuleTypeNdYAG); err != nil {
			return err
		}
	}

	return s.ndyagRepo.Update(ctx, detail)
}

// ---------------------------------------------------------------------------
// CO2 methods
// ---------------------------------------------------------------------------

// CreateCO2Module validates the device, creates the base session module row,
// and inserts the CO2 detail record.
func (s *EnergyModuleService) CreateCO2Module(ctx context.Context, sessionID int64, detail *domain.CO2ModuleDetail, userID int64) (*domain.CO2ModuleDetail, error) {
	if err := s.validateDeviceForModule(ctx, detail.DeviceID, detail.HandpieceID, domain.ModuleTypeCO2); err != nil {
		return nil, err
	}

	baseModule, err := s.sessionSvc.AddModule(ctx, sessionID, domain.ModuleTypeCO2, userID)
	if err != nil {
		return nil, fmt.Errorf("creating base module: %w", err)
	}

	detail.ModuleID = baseModule.ID
	detail.CreatedBy = userID
	detail.UpdatedBy = userID
	detail.Version = 1

	if err := s.co2Repo.Create(ctx, detail); err != nil {
		return nil, fmt.Errorf("creating CO2 detail: %w", err)
	}

	return detail, nil
}

// GetCO2Module retrieves a CO2 module detail by its parent module ID.
func (s *EnergyModuleService) GetCO2Module(ctx context.Context, moduleID int64) (*domain.CO2ModuleDetail, error) {
	return s.co2Repo.GetByModuleID(ctx, moduleID)
}

// UpdateCO2Module validates the device (if set) and updates the CO2 detail record.
func (s *EnergyModuleService) UpdateCO2Module(ctx context.Context, detail *domain.CO2ModuleDetail, userID int64) error {
	detail.UpdatedBy = userID

	if detail.DeviceID != 0 {
		if err := s.validateDeviceForModule(ctx, detail.DeviceID, detail.HandpieceID, domain.ModuleTypeCO2); err != nil {
			return err
		}
	}

	return s.co2Repo.Update(ctx, detail)
}

// ---------------------------------------------------------------------------
// RF methods
// ---------------------------------------------------------------------------

// CreateRFModule validates the device, creates the base session module row,
// and inserts the RF detail record.
func (s *EnergyModuleService) CreateRFModule(ctx context.Context, sessionID int64, detail *domain.RFModuleDetail, userID int64) (*domain.RFModuleDetail, error) {
	if err := s.validateDeviceForModule(ctx, detail.DeviceID, detail.HandpieceID, domain.ModuleTypeRF); err != nil {
		return nil, err
	}

	baseModule, err := s.sessionSvc.AddModule(ctx, sessionID, domain.ModuleTypeRF, userID)
	if err != nil {
		return nil, fmt.Errorf("creating base module: %w", err)
	}

	detail.ModuleID = baseModule.ID
	detail.CreatedBy = userID
	detail.UpdatedBy = userID
	detail.Version = 1

	if err := s.rfRepo.Create(ctx, detail); err != nil {
		return nil, fmt.Errorf("creating RF detail: %w", err)
	}

	return detail, nil
}

// GetRFModule retrieves an RF module detail by its parent module ID.
func (s *EnergyModuleService) GetRFModule(ctx context.Context, moduleID int64) (*domain.RFModuleDetail, error) {
	return s.rfRepo.GetByModuleID(ctx, moduleID)
}

// UpdateRFModule validates the device (if set) and updates the RF detail record.
func (s *EnergyModuleService) UpdateRFModule(ctx context.Context, detail *domain.RFModuleDetail, userID int64) error {
	detail.UpdatedBy = userID

	if detail.DeviceID != 0 {
		if err := s.validateDeviceForModule(ctx, detail.DeviceID, detail.HandpieceID, domain.ModuleTypeRF); err != nil {
			return err
		}
	}

	return s.rfRepo.Update(ctx, detail)
}
