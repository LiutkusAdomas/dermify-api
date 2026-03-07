package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"time"

	"dermify-api/internal/api/apierrors"
	"dermify-api/internal/api/metrics"
	"dermify-api/internal/api/middleware"
	"dermify-api/internal/domain"
	"dermify-api/internal/service"

	"github.com/go-chi/chi/v5"
)

type createPatientRequest struct {
	FirstName         string  `json:"first_name"`
	LastName          string  `json:"last_name"`
	DateOfBirth       string  `json:"date_of_birth"`
	Sex               string  `json:"sex"`
	Phone             *string `json:"phone"`
	Email             *string `json:"email"`
	ExternalReference *string `json:"external_reference"`
}

type updatePatientRequest struct {
	FirstName         string  `json:"first_name"`
	LastName          string  `json:"last_name"`
	DateOfBirth       string  `json:"date_of_birth"`
	Sex               string  `json:"sex"`
	Phone             *string `json:"phone"`
	Email             *string `json:"email"`
	ExternalReference *string `json:"external_reference"`
	Version           int     `json:"version"`
}

// HandleCreatePatient creates a new patient record.
//
//	@Summary		Create patient
//	@Description	Creates a new patient record with demographics.
//	@Tags			patients
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		createPatientRequest	true	"Patient details"
//	@Success		201		{object}	PatientResponse
//	@Failure		400		{object}	apierrors.ErrorResponse
//	@Failure		500		{object}	apierrors.ErrorResponse
//	@Router			/patients [post]
func HandleCreatePatient(svc *service.PatientService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		claims := middleware.GetUserClaims(r.Context())
		if claims == nil {
			apierrors.WriteError(w, http.StatusUnauthorized,
				apierrors.AuthNotAuthenticated, "not authenticated")
			return
		}

		var req createPatientRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		dob, err := time.Parse("2006-01-02", req.DateOfBirth)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.PatientInvalidData, "invalid date_of_birth format, expected YYYY-MM-DD")
			return
		}

		patient := &domain.Patient{
			FirstName:         req.FirstName,
			LastName:          req.LastName,
			DateOfBirth:       dob,
			Sex:               req.Sex,
			Phone:             req.Phone,
			Email:             req.Email,
			ExternalReference: req.ExternalReference,
			CreatedBy:         claims.UserID,
			UpdatedBy:         claims.UserID,
		}

		if err := svc.Create(r.Context(), patient); err != nil {
			handlePatientCreateError(w, err)
			return
		}

		m.IncrementPatientCreatedCount()

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(toPatientResponse(patient, 0, nil)) //nolint:errcheck // response write
	}
}

// HandleListPatients returns a paginated list of patients with optional search.
//
//	@Summary		List patients
//	@Description	Returns paginated list of patients with optional search by name, phone, or email.
//	@Tags			patients
//	@Produce		json
//	@Security		BearerAuth
//	@Param			search		query	string	false	"Search term"
//	@Param			page		query	int		false	"Page number"	default(1)
//	@Param			per_page	query	int		false	"Items per page"	default(20)
//	@Success		200	{object}	PaginatedResponse
//	@Failure		500	{object}	apierrors.ErrorResponse
//	@Router			/patients [get]
func HandleListPatients(svc *service.PatientService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		filter := service.PatientFilter{
			Search:  r.URL.Query().Get("search"),
			Page:    parseIntParam(r, "page", 1),
			PerPage: parseIntParam(r, "per_page", 20),
		}

		result, err := svc.List(r.Context(), filter)
		if err != nil {
			slog.Error("failed to list patients", "error", err)
			apierrors.WriteError(w, http.StatusInternalServerError,
				apierrors.PatientLookupFailed, "failed to list patients")
			return
		}

		patients := make([]PatientResponse, len(result.Patients))
		for i, p := range result.Patients {
			patients[i] = toPatientResponse(&p.Patient, p.SessionCount, p.LastSessionDate)
		}

		totalPages := 0
		if filter.PerPage > 0 {
			totalPages = int(math.Ceil(float64(result.Total) / float64(filter.PerPage)))
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(PaginatedResponse{ //nolint:errcheck // response write
			Data:       patients,
			Total:      result.Total,
			Page:       filter.Page,
			PerPage:    filter.PerPage,
			TotalPages: totalPages,
		})
	}
}

// HandleGetPatient returns a single patient by ID.
//
//	@Summary		Get patient
//	@Description	Returns a single patient record by ID.
//	@Tags			patients
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Patient ID"
//	@Success		200	{object}	PatientResponse
//	@Failure		400	{object}	apierrors.ErrorResponse
//	@Failure		404	{object}	apierrors.ErrorResponse
//	@Router			/patients/{id} [get]
func HandleGetPatient(svc *service.PatientService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		id, err := parseIDParam(r)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.PatientInvalidData, "invalid patient ID")
			return
		}

		patient, err := svc.GetByID(r.Context(), id)
		if err != nil {
			handlePatientLookupError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(toPatientResponse(patient, 0, nil)) //nolint:errcheck // response write
	}
}

// HandleUpdatePatient updates a patient record with optimistic version locking.
//
//	@Summary		Update patient
//	@Description	Updates a patient record. Requires version field for optimistic locking.
//	@Tags			patients
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	int					true	"Patient ID"
//	@Param			request	body	updatePatientRequest	true	"Updated patient details"
//	@Success		200		{object}	PatientResponse
//	@Failure		400		{object}	apierrors.ErrorResponse
//	@Failure		404		{object}	apierrors.ErrorResponse
//	@Failure		409		{object}	apierrors.ErrorResponse
//	@Router			/patients/{id} [put]
func HandleUpdatePatient(svc *service.PatientService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		claims := middleware.GetUserClaims(r.Context())
		if claims == nil {
			apierrors.WriteError(w, http.StatusUnauthorized,
				apierrors.AuthNotAuthenticated, "not authenticated")
			return
		}

		id, err := parseIDParam(r)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.PatientInvalidData, "invalid patient ID")
			return
		}

		var req updatePatientRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		dob, err := time.Parse("2006-01-02", req.DateOfBirth)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.PatientInvalidData, "invalid date_of_birth format, expected YYYY-MM-DD")
			return
		}

		patient := &domain.Patient{
			ID:                id,
			FirstName:         req.FirstName,
			LastName:          req.LastName,
			DateOfBirth:       dob,
			Sex:               req.Sex,
			Phone:             req.Phone,
			Email:             req.Email,
			ExternalReference: req.ExternalReference,
			Version:           req.Version,
			UpdatedBy:         claims.UserID,
		}

		if err := svc.Update(r.Context(), patient); err != nil {
			handlePatientUpdateError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(toPatientResponse(patient, 0, nil)) //nolint:errcheck // response write
	}
}

// HandleGetPatientSessions returns the session history for a patient.
//
//	@Summary		Get patient sessions
//	@Description	Returns the session history for a patient (empty until Phase 2).
//	@Tags			patients
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Patient ID"
//	@Success		200	{array}		domain.SessionSummary
//	@Failure		400	{object}	apierrors.ErrorResponse
//	@Failure		500	{object}	apierrors.ErrorResponse
//	@Router			/patients/{id}/sessions [get]
func HandleGetPatientSessions(svc *service.PatientService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		id, err := parseIDParam(r)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.PatientInvalidData, "invalid patient ID")
			return
		}

		sessions, err := svc.GetSessionHistory(r.Context(), id)
		if err != nil {
			slog.Error("failed to get session history", "patient_id", id, "error", err)
			apierrors.WriteError(w, http.StatusInternalServerError,
				apierrors.PatientLookupFailed, "failed to get session history")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(sessions) //nolint:errcheck // response write
	}
}

// toPatientResponse converts a domain patient to an API response.
func toPatientResponse(p *domain.Patient, sessionCount int, lastSessionDate *time.Time) PatientResponse {
	return PatientResponse{
		ID:                p.ID,
		FirstName:         p.FirstName,
		LastName:          p.LastName,
		DateOfBirth:       p.DateOfBirth.Format("2006-01-02"),
		Sex:               p.Sex,
		Phone:             p.Phone,
		Email:             p.Email,
		ExternalReference: p.ExternalReference,
		Version:           p.Version,
		SessionCount:      sessionCount,
		LastSessionDate:   lastSessionDate,
		CreatedAt:         p.CreatedAt,
		CreatedBy:         p.CreatedBy,
		UpdatedAt:         p.UpdatedAt,
		UpdatedBy:         p.UpdatedBy,
	}
}

// parseIDParam extracts the "id" URL parameter as an int64.
func parseIDParam(r *http.Request) (int64, error) {
	idStr := chi.URLParam(r, "id")
	return strconv.ParseInt(idStr, 10, 64)
}

// parseIntParam extracts an integer query parameter with a default value.
func parseIntParam(r *http.Request, key string, defaultVal int) int {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultVal
	}

	parsed, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}

	return parsed
}

// handlePatientCreateError maps service create errors to HTTP responses.
func handlePatientCreateError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidPatientData):
		apierrors.WriteError(w, http.StatusBadRequest,
			apierrors.PatientInvalidData, "invalid patient data: check required fields")
	default:
		slog.Error("failed to create patient", "error", err)
		apierrors.WriteError(w, http.StatusInternalServerError,
			apierrors.PatientCreationFailed, "failed to create patient")
	}
}

// handlePatientLookupError maps service lookup errors to HTTP responses.
func handlePatientLookupError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrPatientNotFound):
		apierrors.WriteError(w, http.StatusNotFound,
			apierrors.PatientNotFound, "patient not found")
	default:
		slog.Error("failed to get patient", "error", err)
		apierrors.WriteError(w, http.StatusInternalServerError,
			apierrors.PatientLookupFailed, "failed to look up patient")
	}
}

// handlePatientUpdateError maps service update errors to HTTP responses.
func handlePatientUpdateError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidPatientData):
		apierrors.WriteError(w, http.StatusBadRequest,
			apierrors.PatientInvalidData, "invalid patient data: check required fields")
	case errors.Is(err, service.ErrPatientVersionConflict):
		apierrors.WriteError(w, http.StatusConflict,
			apierrors.PatientVersionConflict, "patient was modified by another user")
	default:
		slog.Error("failed to update patient", "error", err)
		apierrors.WriteError(w, http.StatusInternalServerError,
			apierrors.PatientUpdateFailed, "failed to update patient")
	}
}
