package events

import (
	"context"
)

// Consumer defines the interface for consuming events
type Consumer interface {
	Start(ctx context.Context, router *Router)
}
