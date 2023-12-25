package main

import (
	"context"
	"fmt"
	"net"

	"github.com/atgane/opentd/apis"
	"github.com/atgane/opentd/pkgs/events"
	"github.com/atgane/opentd/pkgs/logging"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type config struct {
	GRPCPort    int
	EventConfig events.EventConfig
	LogLevel    string
}

func main() {
	// TODO: make env loader
	conf := config{
		GRPCPort: 17011,
		EventConfig: events.EventConfig{
			EventType: events.NATS,
			NATSConfig: events.NATSConfig{
				NATSServer: "localhost:4222",
				Subject:    "some-subject",
			},
		},
		LogLevel: "trace",
	}

	logging.SetLevel(conf.LogLevel)

	var err error
	var consumerClient cloudevents.Client
	consumerClient, err = events.NewConsumerEvent(conf.EventConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize event")
		return
	}
	log.Debug().Msg("event consumer client initializing success")

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", conf.GRPCPort))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to net.Listen()")
		return
	}

	gs, err := newFrontend(conf, consumerClient)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize frontend")
		return
	}
	log.Debug().Msg("frontend initializing success")

	if err := gs.Serve(l); err != nil {
		log.Fatal().Err(err).Msg("failed to gs.Serve()")
		return
	}
}

type frontend struct {
	consumerClient cloudevents.Client
	apis.UnimplementedFrontendServer
}

func newFrontend(c config, ec cloudevents.Client) (*grpc.Server, error) {
	gs := grpc.NewServer()
	fs := new(frontend)
	fs.consumerClient = ec
	apis.RegisterFrontendServer(gs, fs)
	return gs, nil
}

func (f *frontend) Buy(ctx context.Context, req *apis.BuyRequest) (*apis.BuyResponse, error) {
	// TODO: add otel tracing
	// TODO: impl logic
	return nil, nil
}

func (f *frontend) Sell(ctx context.Context, req *apis.SellRequest) (*apis.SellResponse, error) {
	return nil, nil
}

func (f *frontend) Cancel(ctx context.Context, req *apis.CancelRequest) (*apis.CancelResponse, error) {
	return nil, nil
}

func (f *frontend) Update(ctx context.Context, req *apis.UpdateRequest) (*apis.UpdateResponse, error) {
	return nil, nil
}
