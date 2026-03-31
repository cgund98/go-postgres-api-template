package registry

type Payload interface {
	EventType() string
	AggregateID() string
}
