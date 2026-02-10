// Package types provides universal types and errors for DataLens.
package types

import (
	"errors"
	"fmt"
)

// =============================================================================
// Sentinel Errors
// =============================================================================

var (
	// ErrNotFound indicates the requested resource does not exist.
	ErrNotFound = errors.New("resource not found")

	// ErrConflict indicates a conflict with the current state (e.g., duplicate).
	ErrConflict = errors.New("resource conflict")

	// ErrUnauthorized indicates missing or invalid authentication.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden indicates the caller lacks permission.
	ErrForbidden = errors.New("forbidden")

	// ErrValidation indicates invalid input.
	ErrValidation = errors.New("validation error")

	// ErrInternal indicates an unexpected internal error.
	ErrInternal = errors.New("internal error")

	// ErrTimeout indicates an operation timed out.
	ErrTimeout = errors.New("operation timed out")

	// ErrRateLimited indicates the caller exceeded rate limits.
	ErrRateLimited = errors.New("rate limited")

	// ErrUnavailable indicates a service is temporarily unavailable.
	ErrUnavailable = errors.New("service unavailable")
)

// =============================================================================
// Domain Errors
// =============================================================================

// DomainError wraps a sentinel error with context.
type DomainError struct {
	Err     error
	Message string
	Code    string
	Details map[string]any
}

func (e *DomainError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %s", e.Err.Error(), e.Message)
	}
	return e.Err.Error()
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

// NewNotFoundError creates a not-found error with context.
func NewNotFoundError(resource string, id any) *DomainError {
	return &DomainError{
		Err:     ErrNotFound,
		Message: fmt.Sprintf("%s with id '%v' not found", resource, id),
		Code:    "NOT_FOUND",
	}
}

// NewConflictError creates a conflict error.
func NewConflictError(resource, field string, value any) *DomainError {
	return &DomainError{
		Err:     ErrConflict,
		Message: fmt.Sprintf("%s with %s '%v' already exists", resource, field, value),
		Code:    "CONFLICT",
	}
}

// NewValidationError creates a validation error.
func NewValidationError(message string, details map[string]any) *DomainError {
	return &DomainError{
		Err:     ErrValidation,
		Message: message,
		Code:    "VALIDATION_ERROR",
		Details: details,
	}
}

// NewUnauthorizedError creates an unauthorized error.
func NewUnauthorizedError(message string) *DomainError {
	return &DomainError{
		Err:     ErrUnauthorized,
		Message: message,
		Code:    "UNAUTHORIZED",
	}
}

// NewForbiddenError creates a forbidden error.
func NewForbiddenError(message string) *DomainError {
	return &DomainError{
		Err:     ErrForbidden,
		Message: message,
		Code:    "FORBIDDEN",
	}
}

// =============================================================================
// Validation Helpers
// =============================================================================

// ValidationErrors collects multiple validation failures.
type ValidationErrors struct {
	Errors []FieldError `json:"errors"`
}

// FieldError describes a single field validation failure.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (v *ValidationErrors) Add(field, message string) {
	v.Errors = append(v.Errors, FieldError{Field: field, Message: message})
}

func (v *ValidationErrors) HasErrors() bool {
	return len(v.Errors) > 0
}

func (v *ValidationErrors) Error() string {
	if len(v.Errors) == 0 {
		return ""
	}
	return fmt.Sprintf("validation failed: %d field(s) invalid", len(v.Errors))
}

func (v *ValidationErrors) ToDomainError() *DomainError {
	details := make(map[string]any)
	for _, e := range v.Errors {
		details[e.Field] = e.Message
	}
	return NewValidationError(v.Error(), details)
}
