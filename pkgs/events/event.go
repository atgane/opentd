package events

import cloudevents "github.com/cloudevents/sdk-go/v2"

const (
	NATS = "nats"
)

type NewEventClient[T any] func(T) (cloudevents.Client, error)
