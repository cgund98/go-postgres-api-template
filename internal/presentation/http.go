package presentation

import (
	"errors"

	"github.com/cgund98/go-postgres-api-template/internal/domain"
	"github.com/cgund98/go-postgres-api-template/internal/observability"
)

var logger = observability.Logger

// GetHTTPStatus converts domain errors to HTTP status codes
// Uses errors.Is() to handle wrapped errors (e.g., fmt.Errorf("%w: ...", domain.ErrInvalidInput))
func GetHTTPStatus(err error) int {
	if errors.Is(err, domain.ErrNotFound) {
		return 404
	}
	if errors.Is(err, domain.ErrAlreadyExists) {
		return 409
	}
	if errors.Is(err, domain.ErrInvalidInput) {
		return 400
	}
	return 500
}
