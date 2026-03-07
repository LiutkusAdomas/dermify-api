//go:build ignore
// +build ignore

// Remove //go:build ignore after Plan 01-03 creates registry service.

package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Suppress unused import lint.
var _ = assert.Equal

// TestListDevices_All tests listing all devices from the registry.
func TestListDevices_All(t *testing.T) {
	t.Skip("Implemented in Plan 01-03")
}

// TestListDevices_ByType tests listing devices filtered by device type.
func TestListDevices_ByType(t *testing.T) {
	t.Skip("Implemented in Plan 01-03")
}

// TestListProducts_All tests listing all products from the registry.
func TestListProducts_All(t *testing.T) {
	t.Skip("Implemented in Plan 01-03")
}

// TestGetDeviceByID_NotFound tests that requesting a non-existent device returns an error.
func TestGetDeviceByID_NotFound(t *testing.T) {
	t.Skip("Implemented in Plan 01-03")
}
