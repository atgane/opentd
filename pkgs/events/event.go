package events

import (
	"fmt"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

const (
	NATS = "nats"
)

type EventConfig struct {
	EventType  string
	NATSConfig NATSConfig
}

func NewConsumerEvent(conf EventConfig) (c cloudevents.Client, err error) {
	if conf.EventType == NATS {
		if c, err = newNATSConsumerEventClient(conf.NATSConfig); err != nil {
			return nil, err
		}
		return c, nil
	}

	return nil, fmt.Errorf("undefined event")
}
