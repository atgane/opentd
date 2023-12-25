package main

import (
	"context"
	"fmt"
	"net"

	"github.com/atgane/opentd/apis"
	"github.com/atgane/opentd/pkgs/events"
	"github.com/atgane/opentd/pkgs/logging"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type config struct {
	GRPCPort   int
	EventType  string
	NATSConfig events.NATSConfig
	LogLevel   string
}

func main() {
	// TODO: make env loader
	conf := config{
		GRPCPort:   17011,
		EventType:  "",
		NATSConfig: events.NATSConfig{},
		LogLevel:   "debug",
	}

	logging.SetLevel(conf.LogLevel)

	if conf.EventType == events.NATS {
	} else {
		err := fmt.Errorf("unidentified event")
		log.Fatal().Err(err).Msg("failed to initialize event")
		return
	}

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", conf.GRPCPort))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to net.Listen()")
		return
	}
	gs, err := newFrontend(conf)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize frontend")
		return
	}
	if err := gs.Serve(l); err != nil {
		log.Fatal().Err(err).Msg("failed to gs.Serve()")
		return
	}
}

type frontend struct {
	apis.UnimplementedFrontendServer
}

func newFrontend(c config) (*grpc.Server, error) {
	gs := grpc.NewServer()
	fs := new(frontend)
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
