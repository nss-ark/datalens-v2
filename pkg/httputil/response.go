// Package httputil provides HTTP response helpers, middleware
// utilities, and shared request/response types.
package httputil

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/complyark/datalens/pkg/types"
)

// =============================================================================
// Response Helpers
// =============================================================================

// Response is the standard API response envelope.
type Response struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   *Error `json:"error,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
}

// Error represents an API error.
type Error struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

// Meta holds pagination metadata.
type Meta struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// JSON writes a JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Response{
		Success: status >= 200 && status < 300,
		Data:    data,
	})
}

// JSONWithPagination writes a paginated JSON response.
func JSONWithPagination(w http.ResponseWriter, data any, page, pageSize, total int) {
	totalPages := total / pageSize
	if total%pageSize > 0 {
		totalPages++
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{
		Success: true,
		Data:    data,
		Meta: &Meta{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// ErrorResponse writes an error response.
func ErrorResponse(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Response{
		Success: false,
		Error: &Error{
			Code:    code,
			Message: message,
		},
	})
}

// ErrorFromDomain maps domain errors to HTTP responses.
func ErrorFromDomain(w http.ResponseWriter, err error) {
	var de *types.DomainError
	if errors.As(err, &de) {
		status := domainErrorToHTTPStatus(de.Err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Error: &Error{
				Code:    de.Code,
				Message: de.Message,
				Details: de.Details,
			},
		})
		return
	}

	ErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred")
}

func domainErrorToHTTPStatus(err error) int {
	switch {
	case errors.Is(err, types.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, types.ErrConflict):
		return http.StatusConflict
	case errors.Is(err, types.ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, types.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, types.ErrValidation):
		return http.StatusBadRequest
	case errors.Is(err, types.ErrRateLimited):
		return http.StatusTooManyRequests
	case errors.Is(err, types.ErrUnavailable):
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// =============================================================================
// Request Helpers
// =============================================================================

// ParsePagination extracts pagination params from query string.
func ParsePagination(r *http.Request) types.Pagination {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return types.Pagination{
		Page:     page,
		PageSize: pageSize,
	}
}

// ParseID extracts and validates an ID from a URL parameter.
func ParseID(s string) (types.ID, error) {
	if s == "" {
		return types.ID{}, types.NewValidationError("id is required", nil)
	}
	id, err := types.ParseID(s)
	if err != nil {
		return types.ID{}, types.NewValidationError("invalid id format", nil)
	}
	return id, nil
}

// DecodeJSON reads and validates a JSON request body.
func DecodeJSON(r *http.Request, v any) error {
	if r.Body == nil {
		return types.NewValidationError("request body is required", nil)
	}
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(v); err != nil {
		return types.NewValidationError("invalid request body: "+err.Error(), nil)
	}
	return nil
}
