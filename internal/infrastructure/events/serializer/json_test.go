package serializer

import (
	"encoding/json"
	"testing"

	userEvents "github.com/cgund98/go-postgres-api-template/internal/domain/user/events"
	"github.com/cgund98/go-postgres-api-template/internal/infrastructure/events"
)

func TestJSONSerializer_Serialize(t *testing.T) {
	tests := []struct {
		name  string
		event any
	}{
		{
			name:  "UserCreatedEvent serializes to JSON",
			event: userEvents.NewUserCreatedEvent("user-123", "test@example.com"),
		},
		{
			name: "UserUpdatedEvent serializes to JSON",
			event: userEvents.NewUserUpdatedEvent("user-123", map[string]any{
				"email": "updated@example.com",
			}),
		},
		{
			name:  "UserDeletedEvent serializes to JSON",
			event: userEvents.NewUserDeletedEvent("user-123"),
		},
	}

	serializer := NewJSONSerializer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Cast event to events.Event interface
			event, ok := tt.event.(events.Event)
			if !ok {
				t.Fatalf("Event does not implement events.Event interface")
			}

			// Serialize the event using the serializer
			jsonData, err := serializer.Serialize(event)
			if err != nil {
				t.Fatalf("Failed to serialize event: %v", err)
			}

			// Verify JSON is not empty
			if len(jsonData) == 0 {
				t.Fatal("Serialized JSON is empty")
			}

			// Verify JSON is valid by attempting to parse it into a map
			var jsonMap map[string]any
			if err := json.Unmarshal(jsonData, &jsonMap); err != nil {
				t.Fatalf("Serialized JSON is not valid: %v", err)
			}

			// Verify required fields are present
			if _, ok := jsonMap["event_id"]; !ok {
				t.Error("JSON missing 'event_id' field")
			}
			if _, ok := jsonMap["event_type"]; !ok {
				t.Error("JSON missing 'event_type' field")
			}
			if _, ok := jsonMap["timestamp"]; !ok {
				t.Error("JSON missing 'timestamp' field")
			}

			// Verify event-specific fields are present
			switch event.Type() {
			case "user.created":
				if _, ok := jsonMap["user_id"]; !ok {
					t.Error("JSON missing 'user_id' field")
				}
				if _, ok := jsonMap["email"]; !ok {
					t.Error("JSON missing 'email' field")
				}
			case "user.updated":
				if _, ok := jsonMap["user_id"]; !ok {
					t.Error("JSON missing 'user_id' field")
				}
				if _, ok := jsonMap["changes"]; !ok {
					t.Error("JSON missing 'changes' field")
				}
			case "user.deleted":
				if _, ok := jsonMap["user_id"]; !ok {
					t.Error("JSON missing 'user_id' field")
				}
			}

			// Verify event_id matches
			if eventID, ok := jsonMap["event_id"].(string); ok {
				if eventID != event.EventID() {
					t.Errorf("EventID mismatch: expected %s, got %s", event.EventID(), eventID)
				}
			} else {
				t.Error("event_id is not a string")
			}

			// Verify event_type matches
			if eventType, ok := jsonMap["event_type"].(string); ok {
				if eventType != event.Type() {
					t.Errorf("EventType mismatch: expected %s, got %s", event.Type(), eventType)
				}
			} else {
				t.Error("event_type is not a string")
			}

			t.Logf("Successfully serialized event to JSON: %s", string(jsonData))
		})
	}
}

func TestJSONSerializer_SerializeErrorHandling(t *testing.T) {
	serializer := NewJSONSerializer()

	// Test with nil event (should handle gracefully or panic - depends on implementation)
	// Since json.Marshal handles nil, this should work
	var nilEvent *userEvents.UserCreatedEvent
	_, err := serializer.Serialize(nilEvent)
	if err != nil {
		t.Logf("Serializer correctly handled nil event with error: %v", err)
	} else {
		t.Log("Serializer handled nil event without error (nil serializes to 'null')")
	}
}

func TestJSONSerializer_ImplementsInterface(t *testing.T) {
	serializer := NewJSONSerializer()

	// Verify serializer implements the Serializer interface
	var _ Serializer = serializer

	// Test that we can call Serialize on the interface
	event := userEvents.NewUserCreatedEvent("user-123", "test@example.com")
	_, err := serializer.Serialize(event)
	if err != nil {
		t.Fatalf("Serializer interface implementation failed: %v", err)
	}
}
