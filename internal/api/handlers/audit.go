package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"dermify-api/internal/api/apierrors"
	"dermify-api/internal/api/metrics"
	"dermify-api/internal/service"
)

// HandleGetAuditTrail returns audit trail entries filtered by entity type and ID.
//
//	@Summary		Get audit trail
//	@Description	Returns audit trail entries filtered by entity type and ID.
//	@Tags			audit
//	@Produce		json
//	@Security		BearerAuth
//	@Param			entity_type	query	string	true	"Entity type (e.g. session, patient)"
//	@Param			entity_id	query	int		true	"Entity ID"
//	@Success		200	{array}		domain.AuditEntry
//	@Failure		400	{object}	apierrors.ErrorResponse
//	@Failure		500	{object}	apierrors.ErrorResponse
//	@Router			/sessions/{id}/audit [get]
func HandleGetAuditTrail(svc *service.AuditService, _ *metrics.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		entityType := r.URL.Query().Get("entity_type")
		entityIDStr := r.URL.Query().Get("entity_id")

		if entityType == "" || entityIDStr == "" {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ValidationRequiredFields, "entity_type and entity_id are required")
			return
		}

		entityID, err := strconv.ParseInt(entityIDStr, 10, 64)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ValidationRequiredFields, "entity_id must be a valid integer")
			return
		}

		entries, err := svc.ListByEntity(r.Context(), entityType, entityID)
		if err != nil {
			slog.Error("failed to get audit trail", "entity_type", entityType, "entity_id", entityID, "error", err)
			apierrors.WriteError(w, http.StatusInternalServerError,
				apierrors.AuditLookupFailed, "failed to retrieve audit trail")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(entries) //nolint:errcheck // response write
	}
}
