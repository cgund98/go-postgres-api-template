package v1

import (
	"github.com/cgund98/go-postgres-api-template/internal/domain/events/registry"
)

const (
	EventTypeUserCreated = "user.created.v1"
	EventTypeUserUpdated = "user.updated.v1"
	EventTypeUserDeleted = "user.deleted.v1"
)

/** -------------------------------- UserCreatedEvent -------------------------------- */

// UserCreatedEvent represents a user created event
type UserCreatedEvent struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

// EventType implements registry.Payload interface
func (e *UserCreatedEvent) EventType() string {
	return EventTypeUserCreated
}

// AggregateID implements registry.Payload interface
func (e *UserCreatedEvent) AggregateID() string {
	return e.UserID
}

// Make sure the event implements the registry.Payload interface
var _ registry.Payload = &UserCreatedEvent{}

/** -------------------------------- UserUpdatedEvent -------------------------------- */

// UserUpdatedEvent represents a user updated event
type UserUpdatedEvent struct {
	UserID  string         `json:"user_id"`
	Changes map[string]any `json:"changes"`
}

// EventType implements registry.Payload interface
func (e *UserUpdatedEvent) EventType() string {
	return EventTypeUserUpdated
}

// AggregateID implements registry.Payload interface
func (e *UserUpdatedEvent) AggregateID() string {
	return e.UserID
}

// Make sure the event implements the registry.Payload interface
var _ registry.Payload = &UserUpdatedEvent{}

/** -------------------------------- UserDeletedEvent -------------------------------- */

// UserDeletedEvent represents a user deleted event
type UserDeletedEvent struct {
	UserID string `json:"user_id"`
}

// EventType implements registry.Payload interface
func (e *UserDeletedEvent) EventType() string {
	return EventTypeUserDeleted
}

// AggregateID implements registry.Payload interface
func (e *UserDeletedEvent) AggregateID() string {
	return e.UserID
}

// Make sure the event implements the registry.Payload interface
var _ registry.Payload = &UserDeletedEvent{}
