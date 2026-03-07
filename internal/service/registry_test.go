package service_test

import (
	"context"
	"testing"
	"time"

	"dermify-api/internal/domain"
	"dermify-api/internal/service"
	"dermify-api/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestListDevices_All tests listing all devices from the registry.
func TestListDevices_All(t *testing.T) {
	now := time.Now()
	allDevices := []domain.Device{
		{ID: 1, Name: "Lumenis M22", Manufacturer: "Lumenis", Model: "M22", DeviceType: "ipl", Active: true, CreatedAt: now},
		{ID: 2, Name: "Candela GentleMax Pro", Manufacturer: "Candela", Model: "GentleMax Pro", DeviceType: "ndyag", Active: true, CreatedAt: now},
		{ID: 3, Name: "Lutronic Genius", Manufacturer: "Lutronic", Model: "Genius", DeviceType: "rf", Active: true, CreatedAt: now},
	}

	mock := &testutil.MockRegistryRepository{
		ListDevicesFn: func(_ context.Context, deviceType string) ([]domain.Device, error) {
			if deviceType == "" {
				return allDevices, nil
			}
			var filtered []domain.Device
			for _, d := range allDevices {
				if d.DeviceType == deviceType {
					filtered = append(filtered, d)
				}
			}
			return filtered, nil
		},
	}

	svc := service.NewRegistryService(mock)
	devices, err := svc.ListDevices(context.Background(), "")

	require.NoError(t, err)
	assert.Len(t, devices, 3)
	assert.Equal(t, "Lumenis M22", devices[0].Name)
	assert.Equal(t, "Candela GentleMax Pro", devices[1].Name)
	assert.Equal(t, "Lutronic Genius", devices[2].Name)
}

// TestListDevices_ByType tests listing devices filtered by device type.
func TestListDevices_ByType(t *testing.T) {
	now := time.Now()
	allDevices := []domain.Device{
		{ID: 1, Name: "Lumenis M22", Manufacturer: "Lumenis", Model: "M22", DeviceType: "ipl", Active: true, CreatedAt: now},
		{ID: 2, Name: "Candela Nordlys", Manufacturer: "Candela", Model: "Nordlys", DeviceType: "ipl", Active: true, CreatedAt: now},
		{ID: 3, Name: "Candela GentleMax Pro", Manufacturer: "Candela", Model: "GentleMax Pro", DeviceType: "ndyag", Active: true, CreatedAt: now},
	}

	mock := &testutil.MockRegistryRepository{
		ListDevicesFn: func(_ context.Context, deviceType string) ([]domain.Device, error) {
			if deviceType == "" {
				return allDevices, nil
			}
			var filtered []domain.Device
			for _, d := range allDevices {
				if d.DeviceType == deviceType {
					filtered = append(filtered, d)
				}
			}
			return filtered, nil
		},
	}

	svc := service.NewRegistryService(mock)
	devices, err := svc.ListDevices(context.Background(), "ipl")

	require.NoError(t, err)
	assert.Len(t, devices, 2)
	assert.Equal(t, "ipl", devices[0].DeviceType)
	assert.Equal(t, "ipl", devices[1].DeviceType)
}

// TestListProducts_All tests listing all products from the registry.
func TestListProducts_All(t *testing.T) {
	now := time.Now()
	conc := "24 mg/mL"
	allProducts := []domain.Product{
		{ID: 1, Name: "Juvederm Ultra XC", Manufacturer: "Allergan", ProductType: "filler", Concentration: &conc, Active: true, CreatedAt: now},
		{ID: 2, Name: "Botox", Manufacturer: "Allergan", ProductType: "botulinum_toxin", Concentration: nil, Active: true, CreatedAt: now},
	}

	mock := &testutil.MockRegistryRepository{
		ListProductsFn: func(_ context.Context, productType string) ([]domain.Product, error) {
			if productType == "" {
				return allProducts, nil
			}
			var filtered []domain.Product
			for _, p := range allProducts {
				if p.ProductType == productType {
					filtered = append(filtered, p)
				}
			}
			return filtered, nil
		},
	}

	svc := service.NewRegistryService(mock)
	products, err := svc.ListProducts(context.Background(), "")

	require.NoError(t, err)
	assert.Len(t, products, 2)
	assert.Equal(t, "Juvederm Ultra XC", products[0].Name)
	assert.Equal(t, "Botox", products[1].Name)
}

// TestGetDeviceByID_NotFound tests that requesting a non-existent device returns an error.
func TestGetDeviceByID_NotFound(t *testing.T) {
	mock := &testutil.MockRegistryRepository{
		GetDeviceByIDFn: func(_ context.Context, _ int64) (*domain.Device, error) {
			return nil, service.ErrDeviceNotFound
		},
	}

	svc := service.NewRegistryService(mock)
	device, err := svc.GetDeviceByID(context.Background(), 999)

	require.Error(t, err)
	assert.ErrorIs(t, err, service.ErrDeviceNotFound)
	assert.Nil(t, device)
}
