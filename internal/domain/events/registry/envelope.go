package registry

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const specVersion = "1.0"

// EventAttributes are additional metadata for the event.
type EventAttributes struct {
	CorrelationID string `json:"correlation_id"`
	AggregateID   string `json:"aggregate_id"`
}

// Envelope is a wrapper struct that implements the CloudEvents spec.
// This is the transit format for events.
type Envelope struct {
	ID          string          `json:"id"`
	Source      string          `json:"source"`
	SpecVersion string          `json:"specversion" default:"1.0"`
	Type        string          `json:"type"`
	Time        string          `json:"time"`
	Data        []byte          `json:"data"`
	Attributes  EventAttributes `json:"attributes"`
}

// EventEnvelope creates a new EventEnvelope
func NewEnvelope(payload Payload, source string, correlationID string) (Envelope, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return Envelope{}, fmt.Errorf("failed to marshal payload: %w", err)
	}

	return Envelope{
		ID:          uuid.New().String(),
		Source:      source,
		SpecVersion: specVersion,
		Type:        payload.EventType(),
		Time:        time.Now().Format(time.RFC3339),
		Data:        data,
		Attributes: EventAttributes{
			CorrelationID: correlationID,
			AggregateID:   payload.AggregateID(),
		},
	}, nil
}

func (e *Envelope) Unmarshal(data []byte) error {
	return json.Unmarshal(data, e)
}

func (e *Envelope) Marshal() ([]byte, error) {
	return json.Marshal(e)
}

func (e *Envelope) CorrelationID() string {
	return e.Attributes.CorrelationID
}

func (e *Envelope) AggregateID() string {
	return e.Attributes.AggregateID
}
