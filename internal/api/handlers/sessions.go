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

type createSessionRequest struct {
	PatientID        int64      `json:"patient_id"`
	ScheduledAt      *time.Time `json:"scheduled_at"`
	PatientGoal      *string    `json:"patient_goal"`
	FitzpatrickType  *int       `json:"fitzpatrick_type"`
	IsTanned         bool       `json:"is_tanned"`
	IsPregnant       bool       `json:"is_pregnant"`
	OnAnticoagulants bool       `json:"on_anticoagulants"`
	PhotoConsent     *string    `json:"photo_consent"`
	Notes            *string    `json:"notes"`
	IndicationCodes  []int64    `json:"indication_code_ids"`
}

type updateSessionRequest struct {
	ScheduledAt      *time.Time `json:"scheduled_at"`
	PatientGoal      *string    `json:"patient_goal"`
	FitzpatrickType  *int       `json:"fitzpatrick_type"`
	IsTanned         bool       `json:"is_tanned"`
	IsPregnant       bool       `json:"is_pregnant"`
	OnAnticoagulants bool       `json:"on_anticoagulants"`
	PhotoConsent     *string    `json:"photo_consent"`
	Notes            *string    `json:"notes"`
	IndicationCodes  []int64    `json:"indication_code_ids"`
	Version          int        `json:"version"`
}

type transitionSessionRequest struct {
	Status string `json:"status"`
}

type addModuleRequest struct {
	ModuleType string `json:"module_type"`
}

// HandleCreateSession creates a new treatment session.
//
//	@Summary		Create session
//	@Description	Creates a new treatment session for a patient.
//	@Tags			sessions
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		createSessionRequest	true	"Session details"
//	@Success		201		{object}	SessionResponse
//	@Failure		400		{object}	apierrors.ErrorResponse
//	@Failure		404		{object}	apierrors.ErrorResponse
//	@Failure		500		{object}	apierrors.ErrorResponse
//	@Router			/sessions [post]
func HandleCreateSession(svc *service.SessionService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		claims := middleware.GetUserClaims(r.Context())
		if claims == nil {
			apierrors.WriteError(w, http.StatusUnauthorized,
				apierrors.AuthNotAuthenticated, "not authenticated")
			return
		}

		var req createSessionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		session := &domain.Session{
			PatientID:        req.PatientID,
			ClinicianID:      claims.UserID,
			ScheduledAt:      req.ScheduledAt,
			PatientGoal:      req.PatientGoal,
			FitzpatrickType:  req.FitzpatrickType,
			IsTanned:         req.IsTanned,
			IsPregnant:       req.IsPregnant,
			OnAnticoagulants: req.OnAnticoagulants,
			PhotoConsent:     req.PhotoConsent,
			Notes:            req.Notes,
			IndicationCodes:  req.IndicationCodes,
		}

		if err := svc.Create(r.Context(), session); err != nil {
			handleSessionCreateError(w, err)
			return
		}

		m.IncrementSessionCreatedCount()

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(toSessionResponse(session)) //nolint:errcheck // response write
	}
}

// HandleGetSession returns a single session by ID.
//
//	@Summary		Get session
//	@Description	Returns a single session by ID.
//	@Tags			sessions
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Session ID"
//	@Success		200	{object}	SessionResponse
//	@Failure		400	{object}	apierrors.ErrorResponse
//	@Failure		404	{object}	apierrors.ErrorResponse
//	@Router			/sessions/{id} [get]
func HandleGetSession(svc *service.SessionService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		id, err := parseIDParam(r)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.SessionInvalidData, "invalid session ID")
			return
		}

		session, err := svc.GetByID(r.Context(), id)
		if err != nil {
			handleSessionLookupError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(toSessionResponse(session)) //nolint:errcheck // response write
	}
}

// HandleListSessions returns a paginated list of sessions with optional filters.
//
//	@Summary		List sessions
//	@Description	Returns a paginated list of sessions with optional filters.
//	@Tags			sessions
//	@Produce		json
//	@Security		BearerAuth
//	@Param			patient_id		query	int		false	"Filter by patient ID"
//	@Param			clinician_id	query	int		false	"Filter by clinician ID"
//	@Param			status			query	string	false	"Filter by status"
//	@Param			page			query	int		false	"Page number"	default(1)
//	@Param			per_page		query	int		false	"Items per page"	default(20)
//	@Success		200	{object}	PaginatedResponse
//	@Failure		500	{object}	apierrors.ErrorResponse
//	@Router			/sessions [get]
func HandleListSessions(svc *service.SessionService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		filter := service.SessionFilter{
			PatientID:   parseInt64Param(r, "patient_id", 0),
			ClinicianID: parseInt64Param(r, "clinician_id", 0),
			Status:      r.URL.Query().Get("status"),
			Page:        parseIntParam(r, "page", 1),
			PerPage:     parseIntParam(r, "per_page", 20),
		}

		result, err := svc.List(r.Context(), filter)
		if err != nil {
			slog.Error("failed to list sessions", "error", err)
			apierrors.WriteError(w, http.StatusInternalServerError,
				apierrors.SessionLookupFailed, "failed to list sessions")
			return
		}

		sessions := make([]SessionResponse, len(result.Sessions))
		for i, s := range result.Sessions {
			sessions[i] = toSessionResponse(&s)
		}

		totalPages := 0
		if filter.PerPage > 0 {
			totalPages = int(math.Ceil(float64(result.Total) / float64(filter.PerPage)))
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(PaginatedResponse{ //nolint:errcheck // response write
			Data:       sessions,
			Total:      result.Total,
			Page:       filter.Page,
			PerPage:    filter.PerPage,
			TotalPages: totalPages,
		})
	}
}

// HandleUpdateSession updates a session's header fields.
//
//	@Summary		Update session
//	@Description	Updates a session's header fields. Requires version for optimistic locking.
//	@Tags			sessions
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	int						true	"Session ID"
//	@Param			request	body	updateSessionRequest	true	"Updated session details"
//	@Success		200		{object}	SessionResponse
//	@Failure		400		{object}	apierrors.ErrorResponse
//	@Failure		404		{object}	apierrors.ErrorResponse
//	@Failure		409		{object}	apierrors.ErrorResponse
//	@Router			/sessions/{id} [put]
func HandleUpdateSession(svc *service.SessionService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
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
				apierrors.SessionInvalidData, "invalid session ID")
			return
		}

		var req updateSessionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		session := &domain.Session{
			ID:               id,
			ScheduledAt:      req.ScheduledAt,
			PatientGoal:      req.PatientGoal,
			FitzpatrickType:  req.FitzpatrickType,
			IsTanned:         req.IsTanned,
			IsPregnant:       req.IsPregnant,
			OnAnticoagulants: req.OnAnticoagulants,
			PhotoConsent:     req.PhotoConsent,
			Notes:            req.Notes,
			IndicationCodes:  req.IndicationCodes,
			Version:          req.Version,
			UpdatedBy:        claims.UserID,
		}

		if err := svc.Update(r.Context(), session); err != nil {
			handleSessionUpdateError(w, err)
			return
		}

		// Re-fetch to get the full updated record.
		updated, err := svc.GetByID(r.Context(), id)
		if err != nil {
			slog.Error("failed to fetch updated session", "id", id, "error", err)
			apierrors.WriteError(w, http.StatusInternalServerError,
				apierrors.SessionLookupFailed, "session updated but failed to retrieve")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(toSessionResponse(updated)) //nolint:errcheck // response write
	}
}

// HandleTransitionSession transitions a session to a new state.
//
//	@Summary		Transition session state
//	@Description	Transitions a session to a new state (draft → in_progress → awaiting_signoff).
//	@Tags			sessions
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	int							true	"Session ID"
//	@Param			request	body	transitionSessionRequest	true	"Target status"
//	@Success		200		{object}	MessageResponse
//	@Failure		400		{object}	apierrors.ErrorResponse
//	@Failure		404		{object}	apierrors.ErrorResponse
//	@Failure		409		{object}	apierrors.ErrorResponse
//	@Router			/sessions/{id}/transition [post]
func HandleTransitionSession(svc *service.SessionService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
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
				apierrors.SessionInvalidData, "invalid session ID")
			return
		}

		var req transitionSessionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		if err := svc.TransitionState(r.Context(), id, req.Status, claims.UserID); err != nil {
			handleSessionTransitionError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(MessageResponse{Message: "state transitioned"}) //nolint:errcheck // response write
	}
}

// HandleAddModule adds a procedure module to a session.
//
//	@Summary		Add module to session
//	@Description	Adds a treatment module (ipl, ndyag, co2, rf, filler, botulinum_toxin) to a session.
//	@Tags			sessions
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	int				true	"Session ID"
//	@Param			request	body	addModuleRequest	true	"Module type"
//	@Success		201		{object}	ModuleResponse
//	@Failure		400		{object}	apierrors.ErrorResponse
//	@Failure		404		{object}	apierrors.ErrorResponse
//	@Failure		409		{object}	apierrors.ErrorResponse
//	@Failure		422		{object}	apierrors.ErrorResponse
//	@Router			/sessions/{id}/modules [post]
func HandleAddModule(svc *service.SessionService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
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
				apierrors.SessionInvalidData, "invalid session ID")
			return
		}

		var req addModuleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		module, err := svc.AddModule(r.Context(), id, req.ModuleType, claims.UserID)
		if err != nil {
			handleModuleCreateError(w, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(toModuleResponse(module)) //nolint:errcheck // response write
	}
}

// HandleListModules returns all modules for a session.
//
//	@Summary		List session modules
//	@Description	Returns all treatment modules for a session.
//	@Tags			sessions
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Session ID"
//	@Success		200	{array}		ModuleResponse
//	@Failure		400	{object}	apierrors.ErrorResponse
//	@Failure		500	{object}	apierrors.ErrorResponse
//	@Router			/sessions/{id}/modules [get]
func HandleListModules(svc *service.SessionService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		id, err := parseIDParam(r)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.SessionInvalidData, "invalid session ID")
			return
		}

		modules, err := svc.ListModules(r.Context(), id)
		if err != nil {
			slog.Error("failed to list modules", "session_id", id, "error", err)
			apierrors.WriteError(w, http.StatusInternalServerError,
				apierrors.ModuleCreationFailed, "failed to list modules")
			return
		}

		resp := make([]ModuleResponse, len(modules))
		for i, mod := range modules {
			resp[i] = toModuleResponse(&mod)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp) //nolint:errcheck // response write
	}
}

// HandleRemoveModule removes a module from a session.
//
//	@Summary		Remove module from session
//	@Description	Removes a treatment module from a session.
//	@Tags			sessions
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id			path	int	true	"Session ID"
//	@Param			moduleId	path	int	true	"Module ID"
//	@Success		204
//	@Failure		400	{object}	apierrors.ErrorResponse
//	@Failure		404	{object}	apierrors.ErrorResponse
//	@Failure		409	{object}	apierrors.ErrorResponse
//	@Router			/sessions/{id}/modules/{moduleId} [delete]
func HandleRemoveModule(svc *service.SessionService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
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
				apierrors.SessionInvalidData, "invalid session ID")
			return
		}

		moduleIDStr := chi.URLParam(r, "moduleId")
		moduleID, err := strconv.ParseInt(moduleIDStr, 10, 64)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ModuleInvalidData, "invalid module ID")
			return
		}

		if err := svc.RemoveModule(r.Context(), id, moduleID, claims.UserID); err != nil {
			handleModuleRemoveError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// toSessionResponse converts a domain session to an API response.
func toSessionResponse(s *domain.Session) SessionResponse {
	return SessionResponse{
		ID:               s.ID,
		PatientID:        s.PatientID,
		ClinicianID:      s.ClinicianID,
		Status:           s.Status,
		ScheduledAt:      s.ScheduledAt,
		StartedAt:        s.StartedAt,
		CompletedAt:      s.CompletedAt,
		PatientGoal:      s.PatientGoal,
		FitzpatrickType:  s.FitzpatrickType,
		IsTanned:         s.IsTanned,
		IsPregnant:       s.IsPregnant,
		OnAnticoagulants: s.OnAnticoagulants,
		PhotoConsent:     s.PhotoConsent,
		Notes:            s.Notes,
		IndicationCodes:  s.IndicationCodes,
		Version:          s.Version,
		CreatedAt:        s.CreatedAt,
		CreatedBy:        s.CreatedBy,
		UpdatedAt:        s.UpdatedAt,
		UpdatedBy:        s.UpdatedBy,
	}
}

// toModuleResponse converts a domain session module to an API response.
func toModuleResponse(m *domain.SessionModule) ModuleResponse {
	return ModuleResponse{
		ID:         m.ID,
		SessionID:  m.SessionID,
		ModuleType: m.ModuleType,
		SortOrder:  m.SortOrder,
		Version:    m.Version,
		CreatedAt:  m.CreatedAt,
		CreatedBy:  m.CreatedBy,
		UpdatedAt:  m.UpdatedAt,
		UpdatedBy:  m.UpdatedBy,
	}
}


// handleSessionCreateError maps service create errors to HTTP responses.
func handleSessionCreateError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidSessionData):
		apierrors.WriteError(w, http.StatusBadRequest,
			apierrors.SessionInvalidData, "invalid session data: check required fields")
	case errors.Is(err, service.ErrPatientNotFound):
		apierrors.WriteError(w, http.StatusNotFound,
			apierrors.PatientNotFound, "patient not found")
	default:
		slog.Error("failed to create session", "error", err)
		apierrors.WriteError(w, http.StatusInternalServerError,
			apierrors.SessionCreationFailed, "failed to create session")
	}
}

// handleSessionLookupError maps service lookup errors to HTTP responses.
func handleSessionLookupError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrSessionNotFound):
		apierrors.WriteError(w, http.StatusNotFound,
			apierrors.SessionNotFound, "session not found")
	default:
		slog.Error("failed to get session", "error", err)
		apierrors.WriteError(w, http.StatusInternalServerError,
			apierrors.SessionLookupFailed, "failed to look up session")
	}
}

// handleSessionUpdateError maps service update errors to HTTP responses.
func handleSessionUpdateError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrSessionNotEditable):
		apierrors.WriteError(w, http.StatusConflict,
			apierrors.SessionNotEditable, "session is not editable in current state")
	case errors.Is(err, service.ErrSessionVersionConflict):
		apierrors.WriteError(w, http.StatusConflict,
			apierrors.SessionVersionConflict, "session was modified by another user")
	case errors.Is(err, service.ErrInvalidSessionData):
		apierrors.WriteError(w, http.StatusBadRequest,
			apierrors.SessionInvalidData, "invalid session data: check field values")
	case errors.Is(err, service.ErrSessionNotFound):
		apierrors.WriteError(w, http.StatusNotFound,
			apierrors.SessionNotFound, "session not found")
	default:
		slog.Error("failed to update session", "error", err)
		apierrors.WriteError(w, http.StatusInternalServerError,
			apierrors.SessionUpdateFailed, "failed to update session")
	}
}

// handleSessionTransitionError maps service transition errors to HTTP responses.
func handleSessionTransitionError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidStateTransition):
		apierrors.WriteError(w, http.StatusConflict,
			apierrors.SessionInvalidTransition, "invalid state transition")
	case errors.Is(err, service.ErrSessionNotFound):
		apierrors.WriteError(w, http.StatusNotFound,
			apierrors.SessionNotFound, "session not found")
	case errors.Is(err, service.ErrSessionVersionConflict):
		apierrors.WriteError(w, http.StatusConflict,
			apierrors.SessionVersionConflict, "session was modified by another user")
	default:
		slog.Error("failed to transition session state", "error", err)
		apierrors.WriteError(w, http.StatusInternalServerError,
			apierrors.SessionUpdateFailed, "failed to transition session state")
	}
}

// handleModuleCreateError maps service module creation errors to HTTP responses.
func handleModuleCreateError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrConsentRequired):
		apierrors.WriteError(w, http.StatusUnprocessableEntity,
			apierrors.ConsentRequired, "consent required before adding modules")
	case errors.Is(err, service.ErrSessionNotEditable):
		apierrors.WriteError(w, http.StatusConflict,
			apierrors.SessionNotEditable, "session is not editable in current state")
	case errors.Is(err, service.ErrInvalidSessionData):
		apierrors.WriteError(w, http.StatusBadRequest,
			apierrors.ModuleInvalidData, "invalid module type")
	case errors.Is(err, service.ErrSessionNotFound):
		apierrors.WriteError(w, http.StatusNotFound,
			apierrors.SessionNotFound, "session not found")
	default:
		slog.Error("failed to add module", "error", err)
		apierrors.WriteError(w, http.StatusInternalServerError,
			apierrors.ModuleCreationFailed, "failed to add module")
	}
}

// handleModuleRemoveError maps service module removal errors to HTTP responses.
func handleModuleRemoveError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrSessionNotEditable):
		apierrors.WriteError(w, http.StatusConflict,
			apierrors.SessionNotEditable, "session is not editable in current state")
	case errors.Is(err, service.ErrSessionNotFound):
		apierrors.WriteError(w, http.StatusNotFound,
			apierrors.SessionNotFound, "session not found")
	default:
		slog.Error("failed to remove module", "error", err)
		apierrors.WriteError(w, http.StatusInternalServerError,
			apierrors.ModuleRemovalFailed, "failed to remove module")
	}
}
