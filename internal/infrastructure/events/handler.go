package events

import (
	"context"
)

// Handler defines the interface for handling consumed events
type Handler[T Event] interface {
	Handle(ctx context.Context, event T) error
}

type BatchHandler[T Event] interface {
	HandleBatch(ctx context.Context, events []T) error
}
