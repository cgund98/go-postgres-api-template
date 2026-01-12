package serializer

import (
	"encoding/json"

	"github.com/cgund98/go-postgres-api-template/internal/infrastructure/events"
)

type JSONSerializer struct {
}

func NewJSONSerializer() JSONSerializer {
	return JSONSerializer{}
}

func (s JSONSerializer) Serialize(event events.Event) ([]byte, error) {
	return json.Marshal(event)
}

// Make sure the serializer implements the Serializer interface
var _ Serializer = JSONSerializer{}
