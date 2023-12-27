package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/atgane/opentd/pkgs/engine"
	"github.com/atgane/opentd/pkgs/events"
	"github.com/atgane/opentd/pkgs/logging"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type DealerConfig struct {
	EventConfig      events.EventConfig
	StreamConfig     events.EventConfig
	RedisConfig      redis.Options
	LogLevel         string
	LockExpireSecond time.Duration
}

func main() {
	// TODO: make env loader
	conf := DealerConfig{
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

	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	<-ctx.Done()
}

type Dealer struct {
	consumerClient   cloudevents.Client
	producerClient   cloudevents.Client
	lockExpireSecond time.Duration
	engine           engine.Engine
}

func NewDealer(conf DealerConfig) (*Dealer, error) {
	// TODO: add matching engine
	ctx := context.Background()
	consumerClient, err := events.NewConsumerEvent(conf.EventConfig)
	if err != nil {
		return nil, err
	}
	producerClient, err := events.NewProducerEvent(conf.StreamConfig)
	if err != nil {
		return nil, err
	}

	redisClient := redis.NewClient(&conf.RedisConfig)
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	d := new(Dealer)
	d.consumerClient = consumerClient
	d.producerClient = producerClient
	d.lockExpireSecond = conf.LockExpireSecond
	return d, nil
}

func (d *Dealer) Start() error {
	ctx := context.Background()

	for {
		if err := d.consumerClient.StartReceiver(ctx, d.receive); err != nil {
			return err
		}
	}
}

func (d *Dealer) receive(ctx context.Context, e cloudevents.Event) (err error) {
	// TODO: impl matching engine logic
	log.Debug().Interface("event", e).Msg("get event")

	switch e.Type() {
	case events.BuyType:
		err = d.engine.AddBuy(e)
	case events.SellType:
		err = d.engine.AddSell(e)
	case events.CancelType:
		err = d.engine.AddCancel(e)
	case events.UpdateBuyType:
		err = d.engine.AddUpdateBuy(e)
	case events.UpdateSellType:
		err = d.engine.AddUpdateSell(e)
	}
	if err != nil {
		return err
	}

	return nil
}
