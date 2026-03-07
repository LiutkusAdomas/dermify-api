package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"dermify-api/config"
	"dermify-api/internal/api/apierrors"
	"dermify-api/internal/api/auth"
	"dermify-api/internal/api/middleware"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestConfig creates a minimal Configuration for testing RequireAuth.
func newTestConfig(secret string) *config.Configuration {
	return &config.Configuration{
		Auth: config.AuthConfig{
			JWTSecret:         secret,
			AccessTokenExpiry: 1 * time.Hour,
		},
	}
}

// TestRequireRole_AllowedRole tests that a user with an allowed role can access the endpoint.
func TestRequireRole_AllowedRole(t *testing.T) {
	// Generate a real token with role="doctor" to pass through RequireAuth.
	secret := "test-secret-key-for-unit-tests"
	token, err := auth.GenerateAccessToken(1, "doc@test.com", "doctor", secret, 60000000000)
	require.NoError(t, err)

	// Build a config-like setup for RequireAuth.
	cfg := newTestConfig(secret)

	// Handler that RequireRole should pass to.
	handler := middleware.RequireRole("doctor", "admin")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Wrap with RequireAuth to populate claims in context.
	fullChain := middleware.RequireAuth(cfg)(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	fullChain.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestRequireRole_DeniedRole tests that a user with a non-allowed role receives 403.
func TestRequireRole_DeniedRole(t *testing.T) {
	secret := "test-secret-key-for-unit-tests"
	token, err := auth.GenerateAccessToken(1, "doc@test.com", "doctor", secret, 60000000000)
	require.NoError(t, err)

	cfg := newTestConfig(secret)

	handler := middleware.RequireRole("admin")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	fullChain := middleware.RequireAuth(cfg)(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	fullChain.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	var errResp apierrors.ErrorResponse
	err = json.NewDecoder(rec.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, apierrors.AuthInsufficientRole, errResp.Code)
}

// TestRequireRole_NoRole tests that a user with no role assigned receives 403.
func TestRequireRole_NoRole(t *testing.T) {
	secret := "test-secret-key-for-unit-tests"
	// Empty role string simulates a user with no role assigned.
	token, err := auth.GenerateAccessToken(1, "new@test.com", "", secret, 60000000000)
	require.NoError(t, err)

	cfg := newTestConfig(secret)

	handler := middleware.RequireRole("doctor")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	fullChain := middleware.RequireAuth(cfg)(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	fullChain.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	var errResp apierrors.ErrorResponse
	err = json.NewDecoder(rec.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, apierrors.AuthInsufficientRole, errResp.Code)
}

// TestRequireRole_NoClaims tests that a request without auth claims receives 401.
func TestRequireRole_NoClaims(t *testing.T) {
	handler := middleware.RequireRole("doctor")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var errResp apierrors.ErrorResponse
	err := json.NewDecoder(rec.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, apierrors.AuthNotAuthenticated, errResp.Code)
}
