package handlers_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"dermify-api/config"
	"dermify-api/internal/api/apierrors"
	"dermify-api/internal/api/auth"
	"dermify-api/internal/api/handlers"
	"dermify-api/internal/api/metrics"
	"dermify-api/internal/api/middleware"
	"dermify-api/internal/domain"
	"dermify-api/internal/service"
	"dermify-api/internal/testutil"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newPatientTestRouter builds a chi router with RequireAuth and RequireRole middleware
// protecting a GET / endpoint that calls HandleListPatients.
func newPatientTestRouter(cfg *config.Configuration, svc *service.PatientService, m *metrics.Client, allowedRoles ...string) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(cfg))
	r.Use(middleware.RequireRole(allowedRoles...))
	r.Get("/", handlers.HandleListPatients(svc, m))
	return r
}

// newPatientTestDeps creates shared test dependencies: config, patient service, and metrics client.
func newPatientTestDeps(secret string) (*config.Configuration, *service.PatientService, *metrics.Client) {
	cfg := &config.Configuration{
		Auth: config.AuthConfig{
			JWTSecret:         secret,
			AccessTokenExpiry: 1 * time.Hour,
		},
	}

	repo := &testutil.MockPatientRepository{
		ListFn: func(_ context.Context, _ service.PatientFilter) (*service.PatientListResult, error) {
			return &service.PatientListResult{Patients: []service.PatientListItem{}, Total: 0}, nil
		},
	}

	svc := service.NewPatientService(repo)
	m := metrics.New(slog.Default())

	return cfg, svc, m
}

// TestDoctorAccess_PatientsEndpoint verifies that a doctor can access patient endpoints (RBAC-02).
func TestDoctorAccess_PatientsEndpoint(t *testing.T) {
	secret := "test-secret-key-for-handler-tests"
	cfg, svc, m := newPatientTestDeps(secret)

	token, err := auth.GenerateAccessToken(1, "doc@test.com", domain.RoleDoctor, secret, time.Hour)
	require.NoError(t, err)

	router := newPatientTestRouter(cfg, svc, m, domain.RoleDoctor, domain.RoleAdmin)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestAdminAccess_PatientsEndpoint verifies that an admin can access patient endpoints (RBAC-03).
func TestAdminAccess_PatientsEndpoint(t *testing.T) {
	secret := "test-secret-key-for-handler-tests"
	cfg, svc, m := newPatientTestDeps(secret)

	token, err := auth.GenerateAccessToken(2, "admin@test.com", domain.RoleAdmin, secret, time.Hour)
	require.NoError(t, err)

	router := newPatientTestRouter(cfg, svc, m, domain.RoleDoctor, domain.RoleAdmin)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestUnauthorizedAccess_PatientsEndpoint verifies that a user without a role receives 403.
func TestUnauthorizedAccess_PatientsEndpoint(t *testing.T) {
	secret := "test-secret-key-for-handler-tests"
	cfg, svc, m := newPatientTestDeps(secret)

	// Empty role simulates a user with no role assigned.
	token, err := auth.GenerateAccessToken(3, "norole@test.com", "", secret, time.Hour)
	require.NoError(t, err)

	router := newPatientTestRouter(cfg, svc, m, domain.RoleDoctor, domain.RoleAdmin)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	var errResp apierrors.ErrorResponse
	err = json.NewDecoder(rec.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, apierrors.AuthInsufficientRole, errResp.Code)
}
