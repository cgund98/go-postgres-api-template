package events

import (
	"context"
	"fmt"

	"github.com/cgund98/go-postgres-api-template/internal/domain/events/registry"
)

// Router defines the interface for routing events to a specific handler
type Router struct {
	handlers map[string]Handler
}

// NewRouter creates a new router
func NewRouter() *Router {
	return &Router{
		handlers: make(map[string]Handler),
	}
}

// RegisterHandler registers a new handler for a specific event type
func (r *Router) RegisterHandler(eventType string, handler Handler) {
	r.handlers[eventType] = handler
}

// Route routes an event to the appropriate handler
func (r *Router) Route(ctx context.Context, envelope registry.Envelope) error {
	handler, ok := r.handlers[envelope.Type]
	if !ok {
		return fmt.Errorf("no handler registered for event type: %s", envelope.Type)
	}
	ctx = WithCorrelationID(ctx, envelope.CorrelationID())
	return handler.Handle(ctx, envelope)
}
