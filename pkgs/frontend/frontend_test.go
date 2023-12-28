package frontend_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/atgane/opentd/apis"
	"github.com/atgane/opentd/pkgs/events"
	"github.com/atgane/opentd/pkgs/frontend"
	"github.com/atgane/opentd/pkgs/logging"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var testFrontendScenario = []frontendScenario{
	{"sell something", testSell},
}

type frontendScenario struct {
	name string
	fn   func(t *testing.T, state *testState)
}

type testState struct {
	conf           frontend.FrontConfig
	f              *frontend.Frontend
	redisClient    *redis.Client
	consumerClient cloudevents.Client
	c              apis.FrontendClient
}

func TestFrontend(t *testing.T) {
	ts := new(testState)
	logging.SetLevel("trace")

	// init test
	ts.conf = frontend.FrontConfig{
		GRPCPort: 17011,
		EventConfig: events.EventConfig{
			EventType: events.NATS,
			NATSConfig: events.NATSConfig{
				NATSServer:  "nats://127.0.0.1:4222",
				Subject:     "some-subject",
				NATSOptions: []nats.Option{},
			},
		},
		RedisConfig: redis.Options{
			Addr: "127.0.0.1:6379",
		},
		LogLevel:         "trace",
		LockExpireSecond: 300 * time.Second,
	}

	f, err := frontend.NewFrontend(ts.conf)
	require.NoError(t, err)
	ts.f = f
	go f.Start()

	ctx := context.Background()

	ts.redisClient = redis.NewClient(&ts.conf.RedisConfig)
	err = ts.redisClient.Ping(ctx).Err()
	require.NoError(t, err)

	ts.consumerClient, err = events.NewConsumerEvent(ts.conf.EventConfig)
	require.NoError(t, err)

	conn, err := grpc.DialContext(
		ctx,
		fmt.Sprintf("localhost:%d", ts.conf.GRPCPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()
	ts.c = apis.NewFrontendClient(conn)

	// run test
	for idx := range testFrontendScenario {
		t.Run(testFrontendScenario[idx].name, func(t *testing.T) {
			testFrontendScenario[idx].fn(t, ts)
		})
	}
}

func testSell(t *testing.T, ts *testState) {
	t.Helper()

	_, err := ts.c.Sell(context.Background(), &apis.SellRequest{
		UserId: "user1",
		Target: "some-subject",
		Amount: 1,
		Price:  30,
	})
	require.NoError(t, err)
}
