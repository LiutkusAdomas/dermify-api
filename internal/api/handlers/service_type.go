package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"dermify-api/internal/api/apierrors"
	"dermify-api/internal/api/metrics"
	"dermify-api/internal/api/middleware"
	"dermify-api/internal/domain"
	"dermify-api/internal/service"
)

type createServiceTypeRequest struct {
	Name            string `json:"name"`
	DefaultDuration int    `json:"default_duration_minutes"`
	Description     string `json:"description"`
}

type updateServiceTypeRequest struct {
	Name            string `json:"name"`
	DefaultDuration int    `json:"default_duration_minutes"`
	Description     string `json:"description"`
	Active          *bool  `json:"active"`
}

// HandleCreateServiceType creates a new service type within an organization.
func HandleCreateServiceType(svc *service.ServiceTypeService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		membership := middleware.GetOrgMembership(r.Context())
		if membership == nil {
			apierrors.WriteError(w, http.StatusForbidden, apierrors.OrgNotAdmin, "admin access required")
			return
		}

		var req createServiceTypeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		if req.Name == "" {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationRequiredFields, "name is required")
			return
		}

		st := &domain.ServiceType{
			OrgID:           membership.OrgID,
			Name:            req.Name,
			DefaultDuration: req.DefaultDuration,
			Description:     req.Description,
		}

		if err := svc.Create(r.Context(), st); err != nil {
			handleServiceTypeError(w, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(st) //nolint:errcheck // response write
	}
}

// HandleListServiceTypes lists service types for an organization.
func HandleListServiceTypes(svc *service.ServiceTypeService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		membership := middleware.GetOrgMembership(r.Context())
		if membership == nil {
			apierrors.WriteError(w, http.StatusForbidden, apierrors.OrgNotMember, "not a member")
			return
		}

		activeOnly := r.URL.Query().Get("active") == "true"
		types, err := svc.ListByOrg(r.Context(), membership.OrgID, activeOnly)
		if err != nil {
			slog.Error("failed to list service types", "error", err)
			apierrors.WriteError(w, http.StatusInternalServerError, apierrors.ServiceTypeLookupFailed, "failed to query service types")
			return
		}

		if types == nil {
			types = []*domain.ServiceType{}
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(types) //nolint:errcheck // response write
	}
}

// HandleGetServiceType returns a single service type.
func HandleGetServiceType(svc *service.ServiceTypeService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		membership := middleware.GetOrgMembership(r.Context())
		if membership == nil {
			apierrors.WriteError(w, http.StatusForbidden, apierrors.OrgNotMember, "not a member")
			return
		}

		id, err := parseIDParam(r)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationInvalidRequestBody, "invalid service type id")
			return
		}

		st, err := svc.GetByID(r.Context(), id, membership.OrgID)
		if err != nil {
			handleServiceTypeError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(st) //nolint:errcheck // response write
	}
}

// HandleUpdateServiceType updates a service type.
func HandleUpdateServiceType(svc *service.ServiceTypeService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		membership := middleware.GetOrgMembership(r.Context())
		if membership == nil {
			apierrors.WriteError(w, http.StatusForbidden, apierrors.OrgNotAdmin, "admin access required")
			return
		}

		id, err := parseIDParam(r)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationInvalidRequestBody, "invalid service type id")
			return
		}

		var req updateServiceTypeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		existing, err := svc.GetByID(r.Context(), id, membership.OrgID)
		if err != nil {
			handleServiceTypeError(w, err)
			return
		}

		existing.Name = req.Name
		existing.DefaultDuration = req.DefaultDuration
		existing.Description = req.Description
		if req.Active != nil {
			existing.Active = *req.Active
		}

		if err := svc.Update(r.Context(), existing); err != nil {
			handleServiceTypeError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(existing) //nolint:errcheck // response write
	}
}

// HandleDeleteServiceType deletes a service type.
func HandleDeleteServiceType(svc *service.ServiceTypeService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		membership := middleware.GetOrgMembership(r.Context())
		if membership == nil {
			apierrors.WriteError(w, http.StatusForbidden, apierrors.OrgNotAdmin, "admin access required")
			return
		}

		id, err := parseIDParam(r)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationInvalidRequestBody, "invalid service type id")
			return
		}

		if err := svc.Delete(r.Context(), id, membership.OrgID); err != nil {
			handleServiceTypeError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "service type deleted"}) //nolint:errcheck // response write
	}
}

func handleServiceTypeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrServiceTypeNotFound):
		apierrors.WriteError(w, http.StatusNotFound, apierrors.ServiceTypeNotFound, "service type not found")
	case errors.Is(err, service.ErrServiceTypeNameExists):
		apierrors.WriteError(w, http.StatusConflict, apierrors.ServiceTypeNameExists, "service type name already exists")
	case errors.Is(err, service.ErrInvalidServiceTypeData):
		apierrors.WriteError(w, http.StatusBadRequest, apierrors.ServiceTypeInvalidData, "invalid service type data")
	default:
		slog.Error("service type operation failed", "error", err)
		apierrors.WriteError(w, http.StatusInternalServerError, apierrors.ServiceTypeLookupFailed, "internal error")
	}
}
