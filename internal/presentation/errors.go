package presentation

import (
	"errors"

	"github.com/danielgtaylor/huma/v2"

	"github.com/cgund98/go-postgres-api-template/internal/domain"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// SanitizeError returns a safe error message for clients
// Domain errors are safe to expose, but internal errors are sanitized
func SanitizeError(err error) string {
	// Check if it's a known domain error - these are safe to expose
	if errors.Is(err, domain.ErrNotFound) {
		return err.Error()
	}
	if errors.Is(err, domain.ErrAlreadyExists) {
		return err.Error()
	}
	if errors.Is(err, domain.ErrInvalidInput) {
		return err.Error()
	}

	// For internal errors, log the full error but return a generic message
	logger.Error("internal server error occurred", "error", err)
	return "An internal error occurred"
}

// NewHumaError creates a new Huma error with the appropriate HTTP status code and sanitized error message
func NewHumaError(err error) huma.StatusError {
	return huma.NewError(GetHTTPStatus(err), SanitizeError(err))
}
