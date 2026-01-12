package handlers

import (
	"context"

	userEvents "github.com/cgund98/go-postgres-api-template/internal/domain/user/events"
	"github.com/cgund98/go-postgres-api-template/internal/infrastructure/events"
	"github.com/cgund98/go-postgres-api-template/internal/observability"
)

var logger = observability.Logger

// UserCreatedHandler handles user created events
type UserCreatedHandler struct{}

// NewUserCreatedHandler creates a new user created handler
func NewUserCreatedHandler() *UserCreatedHandler {
	return &UserCreatedHandler{}
}

// Handle processes a user created event
func (h *UserCreatedHandler) Handle(_ context.Context, event *userEvents.UserCreatedEvent) error {
	// TODO: Implement event handling logic
	logger.Info("Handling user created event", "event", event)
	return nil
}

// UserUpdatedHandler handles user updated events
type UserUpdatedHandler struct{}

// NewUserUpdatedHandler creates a new user updated handler
func NewUserUpdatedHandler() *UserUpdatedHandler {
	return &UserUpdatedHandler{}
}

// Handle processes a user updated event
func (h *UserUpdatedHandler) Handle(_ context.Context, event *userEvents.UserUpdatedEvent) error {
	// TODO: Implement event handling logic
	logger.Info("Handling user updated event", "event", event)
	return nil
}

// UserDeletedHandler handles user deleted events
type UserDeletedHandler struct{}

// NewUserDeletedHandler creates a new user deleted handler
func NewUserDeletedHandler() *UserDeletedHandler {
	return &UserDeletedHandler{}
}

// Handle processes a user deleted event
func (h *UserDeletedHandler) Handle(_ context.Context, event *userEvents.UserDeletedEvent) error {
	// TODO: Implement event handling logic
	logger.Info("Handling user deleted event", "event", event)
	return nil
}

// Make sure the handler implements the events.Handler interface
var _ events.Handler[*userEvents.UserCreatedEvent] = &UserCreatedHandler{}
var _ events.Handler[*userEvents.UserUpdatedEvent] = &UserUpdatedHandler{}
var _ events.Handler[*userEvents.UserDeletedEvent] = &UserDeletedHandler{}
