package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

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

// parseInt64Param extracts an int64 query parameter with a default value.
func parseInt64Param(r *http.Request, key string, defaultVal int64) int64 {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultVal
	}

	parsed, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return defaultVal
	}

	return parsed
}

// parseModuleIDParam extracts the module ID from the URL path parameter.
func parseModuleIDParam(r *http.Request) (int64, error) {
	idStr := chi.URLParam(r, "moduleId")
	return strconv.ParseInt(idStr, 10, 64)
}
