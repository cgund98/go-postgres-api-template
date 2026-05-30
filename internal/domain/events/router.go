package events

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/trace"

	"github.com/cgund98/go-postgres-api-template/internal/domain/events/registry"
	"github.com/cgund98/go-postgres-api-template/internal/observability"
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

func (r *Router) augmentContext(ctx context.Context, envelope registry.Envelope) context.Context {
	logger := observability.LoggerFromContext(ctx)
	if logger == nil {
		logger = observability.NewLogger("postgres-template/events")
	}
	logger = logger.With(
		"event_type", envelope.Type,
		"correlation_id", envelope.CorrelationID(),
	)
	if spanCtx := trace.SpanFromContext(ctx).SpanContext(); spanCtx.IsValid() {
		logger = logger.With(
			"trace_id", spanCtx.TraceID().String(),
			"span_id", spanCtx.SpanID().String(),
		)
	}
	ctx = observability.SetLoggerOnContext(ctx, logger)

	return ctx
}

// Route routes an event to the appropriate handler
func (r *Router) Route(ctx context.Context, envelope registry.Envelope) error {
	ctx, endSpan := observability.StartTraceFromContext(ctx, "Route")
	defer endSpan()

	handler, ok := r.handlers[envelope.Type]
	if !ok {
		return fmt.Errorf("no handler registered for event type: %s", envelope.Type)
	}
	ctx = WithCorrelationID(ctx, envelope.CorrelationID())

	ctx = r.augmentContext(ctx, envelope)
	return handler.Handle(ctx, envelope)
}
