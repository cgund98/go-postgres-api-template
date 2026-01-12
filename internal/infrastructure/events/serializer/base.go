package serializer

import "github.com/cgund98/go-postgres-api-template/internal/infrastructure/events"

type Serializer interface {
	Serialize(event events.Event) ([]byte, error)
}
