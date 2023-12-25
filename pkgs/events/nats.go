package events

import (
	"context"

	cenats "github.com/cloudevents/sdk-go/protocol/nats/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/nats-io/nats.go"
)

type NATSConfig struct {
	NATSServer  string
	Subject     string
	NATSOptions []nats.Option
}

func NewNATSEventClient(conf NATSConfig) (cloudevents.Client, error) {
	ctx := context.Background()
	p, err := cenats.NewConsumer(conf.NATSServer, conf.Subject, conf.NATSOptions)
	if err != nil {
		return nil, err
	}

	defer p.Close(ctx)
	c, err := cloudevents.NewClient(p)
	if err != nil {
		return nil, err
	}

	return c, nil
}
