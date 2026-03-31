package events

import (
	"context"

	"github.com/cgund98/go-postgres-api-template/internal/domain/events/registry"
)

// Handler defines the interface for handling events
type Handler interface {
	Handle(ctx context.Context, envelope registry.Envelope) error
}
