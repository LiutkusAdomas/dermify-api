package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"dermify-api/internal/api/apierrors"
	"dermify-api/internal/api/metrics"
	"dermify-api/internal/service"

	"github.com/go-chi/chi/v5"
)

// HandleListDevices returns all active devices, optionally filtered by ?type= query param.
//
//	@Summary		List devices
//	@Description	Returns all active devices in the registry, optionally filtered by device type.
//	@Tags			registry
//	@Produce		json
//	@Security		BearerAuth
//	@Param			type	query		string	false	"Device type filter (ipl, ndyag, co2, rf)"
//	@Success		200		{array}		domain.Device
//	@Failure		500		{object}	apierrors.ErrorResponse
//	@Router			/registry/devices [get]
func HandleListDevices(svc *service.RegistryService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		deviceType := r.URL.Query().Get("type")

		devices, err := svc.ListDevices(r.Context(), deviceType)
		if err != nil {
			slog.Error("failed to list devices", "error", err)
			apierrors.WriteError(w, http.StatusInternalServerError,
				apierrors.RegistryLookupFailed, "failed to list devices")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(devices) //nolint:errcheck // response write
	}
}

// HandleGetDevice returns a single device with its associated handpieces.
//
//	@Summary		Get device by ID
//	@Description	Returns a single device with its associated handpieces.
//	@Tags			registry
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Device ID"
//	@Success		200	{object}	domain.Device
//	@Failure		400	{object}	apierrors.ErrorResponse
//	@Failure		404	{object}	apierrors.ErrorResponse
//	@Failure		500	{object}	apierrors.ErrorResponse
//	@Router			/registry/devices/{id} [get]
func HandleGetDevice(svc *service.RegistryService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ValidationInvalidRequestBody, "invalid device id")
			return
		}

		device, err := svc.GetDeviceByID(r.Context(), id)
		if err != nil {
			if errors.Is(err, service.ErrDeviceNotFound) {
				apierrors.WriteError(w, http.StatusNotFound,
					apierrors.RegistryDeviceNotFound, "device not found")
				return
			}
			slog.Error("failed to get device", "error", err, "id", id)
			apierrors.WriteError(w, http.StatusInternalServerError,
				apierrors.RegistryLookupFailed, "failed to get device")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(device) //nolint:errcheck // response write
	}
}

// HandleListProducts returns all active products, optionally filtered by ?type= query param.
//
//	@Summary		List products
//	@Description	Returns all active products in the registry, optionally filtered by product type.
//	@Tags			registry
//	@Produce		json
//	@Security		BearerAuth
//	@Param			type	query		string	false	"Product type filter (filler, botulinum_toxin)"
//	@Success		200		{array}		domain.Product
//	@Failure		500		{object}	apierrors.ErrorResponse
//	@Router			/registry/products [get]
func HandleListProducts(svc *service.RegistryService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		productType := r.URL.Query().Get("type")

		products, err := svc.ListProducts(r.Context(), productType)
		if err != nil {
			slog.Error("failed to list products", "error", err)
			apierrors.WriteError(w, http.StatusInternalServerError,
				apierrors.RegistryLookupFailed, "failed to list products")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(products) //nolint:errcheck // response write
	}
}

// HandleGetProduct returns a single product by ID.
//
//	@Summary		Get product by ID
//	@Description	Returns a single product from the registry.
//	@Tags			registry
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Product ID"
//	@Success		200	{object}	domain.Product
//	@Failure		400	{object}	apierrors.ErrorResponse
//	@Failure		404	{object}	apierrors.ErrorResponse
//	@Failure		500	{object}	apierrors.ErrorResponse
//	@Router			/registry/products/{id} [get]
func HandleGetProduct(svc *service.RegistryService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ValidationInvalidRequestBody, "invalid product id")
			return
		}

		product, err := svc.GetProductByID(r.Context(), id)
		if err != nil {
			if errors.Is(err, service.ErrProductNotFound) {
				apierrors.WriteError(w, http.StatusNotFound,
					apierrors.RegistryProductNotFound, "product not found")
				return
			}
			slog.Error("failed to get product", "error", err, "id", id)
			apierrors.WriteError(w, http.StatusInternalServerError,
				apierrors.RegistryLookupFailed, "failed to get product")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(product) //nolint:errcheck // response write
	}
}

// HandleListIndicationCodes returns indication codes, optionally filtered by ?module_type= query param.
//
//	@Summary		List indication codes
//	@Description	Returns all active indication codes, optionally filtered by module type.
//	@Tags			registry
//	@Produce		json
//	@Security		BearerAuth
//	@Param			module_type	query		string	false	"Module type filter (ipl, ndyag, co2, rf, filler, botulinum_toxin)"
//	@Success		200			{array}		domain.IndicationCode
//	@Failure		500			{object}	apierrors.ErrorResponse
//	@Router			/registry/indication-codes [get]
func HandleListIndicationCodes(svc *service.RegistryService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		moduleType := r.URL.Query().Get("module_type")

		codes, err := svc.ListIndicationCodes(r.Context(), moduleType)
		if err != nil {
			slog.Error("failed to list indication codes", "error", err)
			apierrors.WriteError(w, http.StatusInternalServerError,
				apierrors.RegistryLookupFailed, "failed to list indication codes")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(codes) //nolint:errcheck // response write
	}
}

// HandleListClinicalEndpoints returns clinical endpoints, optionally filtered by ?module_type= query param.
//
//	@Summary		List clinical endpoints
//	@Description	Returns all active clinical endpoints, optionally filtered by module type.
//	@Tags			registry
//	@Produce		json
//	@Security		BearerAuth
//	@Param			module_type	query		string	false	"Module type filter (ipl, ndyag, co2, rf, filler, botulinum_toxin)"
//	@Success		200			{array}		domain.ClinicalEndpoint
//	@Failure		500			{object}	apierrors.ErrorResponse
//	@Router			/registry/clinical-endpoints [get]
func HandleListClinicalEndpoints(svc *service.RegistryService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		moduleType := r.URL.Query().Get("module_type")

		endpoints, err := svc.ListClinicalEndpoints(r.Context(), moduleType)
		if err != nil {
			slog.Error("failed to list clinical endpoints", "error", err)
			apierrors.WriteError(w, http.StatusInternalServerError,
				apierrors.RegistryLookupFailed, "failed to list clinical endpoints")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(endpoints) //nolint:errcheck // response write
	}
}
