package main

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/atgane/opentd/apis"
	"github.com/atgane/opentd/pkgs/events"
	"github.com/atgane/opentd/pkgs/logging"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
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
	var producerClient cloudevents.Client
	producerClient, err = events.NewProducerEvent(conf.EventConfig)
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

	gs, err := newFrontend(conf, producerClient)
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
	producerClient cloudevents.Client
	apis.UnimplementedFrontendServer
}

func newFrontend(c config, ec cloudevents.Client) (*grpc.Server, error) {
	gs := grpc.NewServer()
	fs := new(frontend)
	fs.producerClient = ec
	apis.RegisterFrontendServer(gs, fs)
	return gs, nil
}

func (f *frontend) Buy(ctx context.Context, req *apis.BuyRequest) (*apis.BuyResponse, error) {
	// TODO: add otel tracing
	log.Debug().Interface("req", req).Msg("buy order accepted")

	rid := uuid.New()

	e := cloudevents.NewEvent()
	e.SetID(rid.String())
	e.SetType(events.BuyType)
	e.SetTime(time.Now())
	e.SetSource(events.FrontendSource)
	_ = e.SetData(cloudevents.ApplicationJSON, req)

	if result := f.producerClient.Send(ctx, e); cloudevents.IsUndelivered(result) {
		err := fmt.Errorf("cloud event message send failed")
		log.Error().
			Err(err).
			Str("user_id", req.UserId).
			Str("target", req.Target).
			Int64("amount", req.Amount).
			Int64("price", req.Price).
			Msg("failed to f.producerClient.Send()")
		return nil, err
	}

	res := new(apis.BuyResponse)
	res.RequestId = rid.String()
	res.Target = req.Target
	res.Amount = req.Amount
	res.Price = req.Price

	return res, nil
}

func (f *frontend) Sell(ctx context.Context, req *apis.SellRequest) (*apis.SellResponse, error) {
	log.Debug().Interface("req", req).Msg("sell order accepted")

	rid := uuid.New()

	e := cloudevents.NewEvent()
	e.SetID(rid.String())
	e.SetType(events.SellType)
	e.SetTime(time.Now())
	e.SetSource(events.FrontendSource)
	_ = e.SetData(cloudevents.ApplicationJSON, req)

	if result := f.producerClient.Send(ctx, e); cloudevents.IsUndelivered(result) {
		err := fmt.Errorf("cloud event message send failed")
		log.Error().
			Err(err).
			Str("user_id", req.UserId).
			Str("target", req.Target).
			Int64("amount", req.Amount).
			Int64("price", req.Price).
			Msg("failed to f.producerClient.Send()")
		return nil, err
	}

	res := new(apis.SellResponse)
	res.RequestId = rid.String()
	res.Target = req.Target
	res.Amount = req.Amount
	res.Price = req.Price

	return res, nil
}

func (f *frontend) Cancel(ctx context.Context, req *apis.CancelRequest) (*apis.CancelResponse, error) {
	// TODO: impl logic
	return nil, nil
}

func (f *frontend) Update(ctx context.Context, req *apis.UpdateRequest) (*apis.UpdateResponse, error) {
	return nil, nil
}
