package consumer

import (
	"context"

	"github.com/cgund98/go-postgres-api-template/internal/adapters/events"
	"github.com/cgund98/go-postgres-api-template/internal/adapters/events/deserializer"
)

// Consumer defines the interface for consuming events
type Consumer[T events.Event] interface {
	Start(ctx context.Context, deserializer deserializer.Deserializer[T], handler events.Handler[T])
	StartBatch(ctx context.Context, deserializer deserializer.Deserializer[T], handler events.BatchHandler[T])
}
