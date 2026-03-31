package deserializer

import (
	"encoding/json"
	"testing"

	userEvents "github.com/cgund98/go-postgres-api-template/internal/domain/user/events"
)

func TestJSONDeserializer_SerializeToJSON(t *testing.T) {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize the event to JSON
			jsonData, err := json.Marshal(tt.event)
			if err != nil {
				t.Fatalf("Failed to serialize event to JSON: %v", err)
			}

			// Verify JSON is not empty
			if len(jsonData) == 0 {
				t.Fatal("Serialized JSON is empty")
			}

			// Verify JSON contains expected fields
			jsonStr := string(jsonData)
			if jsonStr == "" {
				t.Fatal("JSON string is empty")
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

			t.Logf("Successfully serialized event to JSON: %s", jsonStr)
		})
	}
}
