package events

import (
	"context"

	"github.com/cgund98/go-postgres-api-template/internal/domain/events/registry"
)

type PublishMetadata struct {
	Source string
}

type PublishArgs struct {
	Payload  registry.Payload
	Metadata PublishMetadata
}

// Publisher defines the interface for publishing events
type Publisher interface {
	Publish(ctx context.Context, args PublishArgs) error
	PublishBatch(ctx context.Context, args []PublishArgs) error
}
