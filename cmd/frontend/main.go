package main

import (
	"time"

	"github.com/atgane/opentd/pkgs/events"
	"github.com/atgane/opentd/pkgs/frontend"
	"github.com/atgane/opentd/pkgs/logging"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

func main() {
	// TODO: make env loader
	conf := frontend.FrontConfig{
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

	fs, err := frontend.NewFrontend(conf)
	if err != nil {
		log.Fatal().Err(err).Msg("frontend initialize error")
		return
	}

	if err := fs.Start(); err != nil {
		log.Fatal().Err(err).Msg("frontend runtime error")
	}
}
