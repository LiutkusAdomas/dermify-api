//go:build ignore
// +build ignore

// Remove //go:build ignore after Plan 01-01 creates RequireRole middleware.

package middleware_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Suppress unused import lint.
var _ = assert.Equal

// TestRequireRole_AllowedRole tests that a user with an allowed role can access the endpoint.
func TestRequireRole_AllowedRole(t *testing.T) {
	t.Skip("Implemented in Plan 01-01")
}

// TestRequireRole_DeniedRole tests that a user with a non-allowed role receives 403.
func TestRequireRole_DeniedRole(t *testing.T) {
	t.Skip("Implemented in Plan 01-01")
}

// TestRequireRole_NoRole tests that a user with no role assigned receives 403.
func TestRequireRole_NoRole(t *testing.T) {
	t.Skip("Implemented in Plan 01-01")
}

// TestRequireRole_NoClaims tests that a request without auth claims receives 401.
func TestRequireRole_NoClaims(t *testing.T) {
	t.Skip("Implemented in Plan 01-01")
}
