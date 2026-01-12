package deserializer

import (
	"github.com/cgund98/go-postgres-api-template/internal/infrastructure/events"
)

// Deserializer defines the interface for deserializing events
type Deserializer[T events.Event] interface {
	Deserialize(data []byte) (T, error)
}
