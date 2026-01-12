package deserializer

import (
	"encoding/json"
	"reflect"

	"github.com/cgund98/go-postgres-api-template/internal/infrastructure/events"
)

// JSONDeserializer is a deserializer for JSON events
// It uses a function to create a new instance of the event type
// This works for both pointer and non-pointer types
type JSONDeserializer[T events.Event] struct {
	// new is a function to create a new instance of the event type
	// It is used to create a new instance of the event type
	// This works for both pointer and non-pointer types
	new func() T
}

func NewJSONDeserializer[T events.Event]() JSONDeserializer[T] {
	return JSONDeserializer[T]{
		new: func() T {
			// Get the type of T by using a pointer to T and then getting its element type
			// This works for both pointer and non-pointer types
			var ptrToT *T
			t := reflect.TypeOf(ptrToT).Elem()

			// If T is already a pointer type, create a new instance of the element type
			if t.Kind() == reflect.Ptr {
				elemType := t.Elem()
				return reflect.New(elemType).Interface().(T)
			}

			// If T is not a pointer, create a new instance
			return reflect.New(t).Elem().Interface().(T)
		},
	}
}

func (d JSONDeserializer[T]) Deserialize(data []byte) (T, error) {
	evt := d.new()
	err := json.Unmarshal(data, evt)
	return evt, err
}

// Make sure the deserializer implements the Deserializer interface
var _ Deserializer[events.Event] = JSONDeserializer[events.Event]{}
