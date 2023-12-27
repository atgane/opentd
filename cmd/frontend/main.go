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
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FrontConfig struct {
	GRPCPort         int
	EventConfig      events.EventConfig
	RedisConfig      redis.Options
	LogLevel         string
	LockExpireSecond time.Duration
}

func main() {
	// TODO: make env loader
	conf := FrontConfig{
		GRPCPort: 17011,
		EventConfig: events.EventConfig{
			EventType: events.NATS,
			NATSConfig: events.NATSConfig{
				NATSServer: "localhost:4222",
				Subject:    "some-subject",
			},
		},
		RedisConfig: redis.Options{
			Addr: "localhost:6379",
		},
		LogLevel:         "trace",
		LockExpireSecond: 300 * time.Second,
	}

	logging.SetLevel(conf.LogLevel)

	fs, err := NewFrontend(conf)
	if err != nil {
		log.Fatal().Err(err).Msg("frontend initialize error")
		return
	}

	if err := fs.Start(); err != nil {
		log.Fatal().Err(err).Msg("frontend runtime error")
	}
}

type frontend struct {
	producerClient   cloudevents.Client
	redisClient      *redis.Client
	port             int
	lockExpireSecond time.Duration
	gs               *grpc.Server

	apis.UnimplementedFrontendServer
}

func NewFrontend(conf FrontConfig) (*frontend, error) {
	ctx := context.Background()

	producerClient, err := events.NewProducerEvent(conf.EventConfig)
	if err != nil {
		return nil, err
	}
	log.Debug().Msg("event consumer client initializing success")

	redisClient := redis.NewClient(&conf.RedisConfig)
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	gs := grpc.NewServer()
	fs := new(frontend)
	fs.producerClient = producerClient
	fs.redisClient = redisClient
	fs.port = conf.GRPCPort
	fs.lockExpireSecond = conf.LockExpireSecond
	fs.gs = gs
	apis.RegisterFrontendServer(gs, fs)
	return fs, nil
}

func (f *frontend) Start() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", f.port))
	if err != nil {
		return err
	}

	return f.gs.Serve(l)
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

	return res, nil
}

func (f *frontend) Cancel(ctx context.Context, req *apis.CancelRequest) (*apis.CancelResponse, error) {
	log.Debug().Interface("req", req).Msg("cancel order accepted")

	result := f.redisClient.SetNX(ctx, req.RequestId, 1, f.lockExpireSecond)
	success, err := result.Result()
	if err != nil {
		log.Error().
			Err(err).
			Str("user_id", req.UserId).
			Str("request_id", req.RequestId).
			Msg("failed to f.redisClient.SetNX()")
		return nil, err
	}
	if !success {
		msg := "deal already has been executed"
		log.Error().
			Err(err).
			Str("user_id", req.UserId).
			Str("request_id", req.RequestId).
			Msg("msg")
		return nil, status.Errorf(codes.AlreadyExists, msg)
	}

	rid := uuid.New()

	e := cloudevents.NewEvent()
	e.SetID(rid.String())
	e.SetType(events.CancelType)
	e.SetTime(time.Now())
	e.SetSource(events.FrontendSource)
	_ = e.SetData(cloudevents.ApplicationJSON, req)

	if result := f.producerClient.Send(ctx, e); cloudevents.IsUndelivered(result) {
		err := fmt.Errorf("cloud event message send failed")
		log.Error().
			Err(err).
			Str("user_id", req.UserId).
			Str("target", req.RequestId).
			Msg("failed to f.producerClient.Send()")
		return nil, err
	}

	res := new(apis.CancelResponse)
	res.RequestId = rid.String()

	return res, nil
}

func (f *frontend) UpdateBuy(ctx context.Context, req *apis.UpdateRequest) (*apis.UpdateResponse, error) {
	log.Debug().Interface("req", req).Msg("update buy order accepted")

	result := f.redisClient.SetNX(ctx, req.RequestId, 1, f.lockExpireSecond)
	success, err := result.Result()
	if err != nil {
		log.Error().
			Err(err).
			Str("user_id", req.UserId).
			Str("request_id", req.RequestId).
			Str("target", req.Target).
			Int64("amount", req.Amount).
			Int64("price", req.Price).
			Msg("failed to f.redisClient.SetNX()")
		return nil, err
	}
	if !success {
		msg := "deal already has been executed"
		log.Error().
			Err(err).
			Str("user_id", req.UserId).
			Str("request_id", req.RequestId).
			Str("target", req.Target).
			Int64("amount", req.Amount).
			Int64("price", req.Price).
			Msg("msg")
		return nil, status.Errorf(codes.AlreadyExists, msg)
	}

	rid := uuid.New()

	e := cloudevents.NewEvent()
	e.SetID(rid.String())
	e.SetType(events.UpdateBuyType)
	e.SetTime(time.Now())
	e.SetSource(events.FrontendSource)
	_ = e.SetData(cloudevents.ApplicationJSON, req)

	if result := f.producerClient.Send(ctx, e); cloudevents.IsUndelivered(result) {
		err := fmt.Errorf("cloud event message send failed")
		log.Error().
			Err(err).
			Str("user_id", req.UserId).
			Str("target", req.RequestId).
			Str("target", req.Target).
			Int64("amount", req.Amount).
			Int64("price", req.Price).
			Msg("failed to f.producerClient.Send()")
		return nil, err
	}

	res := new(apis.UpdateResponse)
	res.RequestId = rid.String()

	return res, nil
}

func (f *frontend) UpdateSell(ctx context.Context, req *apis.UpdateRequest) (*apis.UpdateResponse, error) {
	log.Debug().Interface("req", req).Msg("update sell order accepted")

	result := f.redisClient.SetNX(ctx, req.RequestId, 1, f.lockExpireSecond)
	success, err := result.Result()
	if err != nil {
		log.Error().
			Err(err).
			Str("user_id", req.UserId).
			Str("request_id", req.RequestId).
			Str("target", req.Target).
			Int64("amount", req.Amount).
			Int64("price", req.Price).
			Msg("failed to f.redisClient.SetNX()")
		return nil, err
	}
	if !success {
		msg := "deal already has been executed"
		log.Error().
			Err(err).
			Str("user_id", req.UserId).
			Str("request_id", req.RequestId).
			Str("target", req.Target).
			Int64("amount", req.Amount).
			Int64("price", req.Price).
			Msg("msg")
		return nil, status.Errorf(codes.AlreadyExists, msg)
	}

	rid := uuid.New()

	e := cloudevents.NewEvent()
	e.SetID(rid.String())
	e.SetType(events.UpdateSellType)
	e.SetTime(time.Now())
	e.SetSource(events.FrontendSource)
	_ = e.SetData(cloudevents.ApplicationJSON, req)

	if result := f.producerClient.Send(ctx, e); cloudevents.IsUndelivered(result) {
		err := fmt.Errorf("cloud event message send failed")
		log.Error().
			Err(err).
			Str("user_id", req.UserId).
			Str("target", req.RequestId).
			Str("target", req.Target).
			Int64("amount", req.Amount).
			Int64("price", req.Price).
			Msg("failed to f.producerClient.Send()")
		return nil, err
	}

	res := new(apis.UpdateResponse)
	res.RequestId = rid.String()

	return res, nil
}
