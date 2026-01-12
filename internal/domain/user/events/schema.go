package events

import (
	"github.com/cgund98/go-postgres-api-template/internal/infrastructure/events"
)

/** -------------------------------- UserCreatedEvent -------------------------------- */

// UserCreatedEvent represents a user created event
type UserCreatedEvent struct {
	events.EventMetadata
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

// Type implements events.Event interface
func (e *UserCreatedEvent) Type() string {
	return e.EventType
}

// EventID implements events.Event interface
func (e *UserCreatedEvent) EventID() string {
	return e.EventMetadata.EventID
}

// AggregateID implements events.Event interface
func (e *UserCreatedEvent) AggregateID() string {
	return e.UserID
}

// Make sure the event implements the events.Event interface
var _ events.Event = &UserCreatedEvent{}

/** -------------------------------- UserUpdatedEvent -------------------------------- */

// UserUpdatedEvent represents a user updated event
type UserUpdatedEvent struct {
	events.EventMetadata
	UserID  string         `json:"user_id"`
	Changes map[string]any `json:"changes"`
}

// Type implements events.Event interface
func (e *UserUpdatedEvent) Type() string {
	return e.EventType
}

// EventID implements events.Event interface
func (e *UserUpdatedEvent) EventID() string {
	return e.EventMetadata.EventID
}

// AggregateID implements events.Event interface
func (e *UserUpdatedEvent) AggregateID() string {
	return e.UserID
}

// Make sure the event implements the events.Event interface
var _ events.Event = &UserUpdatedEvent{}

/** -------------------------------- UserDeletedEvent -------------------------------- */

// UserDeletedEvent represents a user deleted event
type UserDeletedEvent struct {
	events.EventMetadata
	UserID string `json:"user_id"`
}

// Type implements events.Event interface
func (e *UserDeletedEvent) Type() string {
	return e.EventType
}

// EventID implements events.Event interface
func (e *UserDeletedEvent) EventID() string {
	return e.EventMetadata.EventID
}

// AggregateID implements events.Event interface
func (e *UserDeletedEvent) AggregateID() string {
	return e.UserID
}

// Make sure the event implements the events.Event interface
var _ events.Event = &UserDeletedEvent{}

/** -------------------------------- Constructors -------------------------------- */

// NewUserCreatedEvent creates a new user created event
func NewUserCreatedEvent(userID, email string) *UserCreatedEvent {
	return &UserCreatedEvent{
		EventMetadata: events.NewBaseEvent(EventTypeUserCreated),
		UserID:        userID,
		Email:         email,
	}
}

// NewUserUpdatedEvent creates a new user updated event
func NewUserUpdatedEvent(userID string, changes map[string]any) *UserUpdatedEvent {
	return &UserUpdatedEvent{
		EventMetadata: events.NewBaseEvent(EventTypeUserUpdated),
		UserID:        userID,
		Changes:       changes,
	}
}

// NewUserDeletedEvent creates a new user deleted event
func NewUserDeletedEvent(userID string) *UserDeletedEvent {
	return &UserDeletedEvent{
		EventMetadata: events.NewBaseEvent(EventTypeUserDeleted),
		UserID:        userID,
	}
}
