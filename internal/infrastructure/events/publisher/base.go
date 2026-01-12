package publisher

import (
	"context"

	"github.com/cgund98/go-postgres-api-template/internal/infrastructure/events"
)

// Publisher defines the interface for publishing events
type Publisher interface {
	Publish(ctx context.Context, event events.Event) error
	PublishBatch(ctx context.Context, events []events.Event) error
}
