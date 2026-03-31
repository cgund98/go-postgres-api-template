package events

import (
	"time"

	"github.com/google/uuid"
)

// Event represents a domain event that can be serialized as JSON.
// All events must implement Type(), EventID(), Timestamp() and embed BaseEvent.
type Event interface {
	// Type returns the event type identifier (e.g., "user.created")
	Type() string

	// EventID returns the unique identifier for this event instance
	EventID() string

	// AggregateID returns the unique identifier for the aggregate that the event belongs to
	AggregateID() string
}

// EventMetadata provides common fields for all domain events.
// Embed this struct in your event types to get EventID, EventType, and Timestamp fields.
type EventMetadata struct {
	EventID   string    `json:"event_id"`
	EventType string    `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
}

// NewBaseEvent creates a new BaseEvent with a generated EventID and current timestamp.
// Use this in your event constructors to initialize the embedded BaseEvent.
func NewBaseEvent(eventType string) EventMetadata {
	return EventMetadata{
		EventID:   uuid.New().String(),
		EventType: eventType,
		Timestamp: time.Now(),
	}
}
