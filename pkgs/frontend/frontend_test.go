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
	{"buy  something", testBuy},
	{"buy  something", testCancel},
	{"buy  something", testUpdateBuy},
	{"buy  something", testUpdateSell},
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
	callbackChan   chan int
	callback       func(ctx context.Context, e cloudevents.Event)
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
	ts.callbackChan = make(chan int, 1)
	ts.callback = func(ctx context.Context, e cloudevents.Event) {
		if e.Type() == events.SellType {
			result := &apis.SellRequest{}
			require.NoError(t, e.DataAs(result))
		} else if e.Type() == events.BuyType {
			result := &apis.BuyRequest{}
			require.NoError(t, e.DataAs(result))
		} else if e.Type() == events.CancelType {
			result := &apis.CancelRequest{}
			require.NoError(t, e.DataAs(result))
		} else if e.Type() == events.UpdateBuyType {
			result := &apis.UpdateRequest{}
			require.NoError(t, e.DataAs(result))
		} else if e.Type() == events.UpdateSellType {
			result := &apis.UpdateRequest{}
			require.NoError(t, e.DataAs(result))
		}

		ts.callbackChan <- 0
	}

	go func() {
		ts.consumerClient.StartReceiver(context.Background(), ts.callback)
	}()

	// run test
	for idx := range testFrontendScenario {
		t.Run(testFrontendScenario[idx].name, func(t *testing.T) {
			testFrontendScenario[idx].fn(t, ts)
		})
	}
}

func testSell(t *testing.T, ts *testState) {
	t.Helper()

	data := &apis.SellRequest{
		UserId: "user1",
		Target: ts.conf.EventConfig.NATSConfig.Subject,
		Amount: 1,
		Price:  30,
	}
	_, err := ts.c.Sell(context.Background(), data)
	require.NoError(t, err)

	require.Equal(t, 0, <-ts.callbackChan)
}

func testBuy(t *testing.T, ts *testState) {
	t.Helper()

	data := &apis.BuyRequest{
		UserId: "user1",
		Target: ts.conf.EventConfig.NATSConfig.Subject,
		Amount: 1,
		Price:  30,
	}
	_, err := ts.c.Buy(context.Background(), data)
	require.NoError(t, err)

	require.Equal(t, 0, <-ts.callbackChan)
}

func testCancel(t *testing.T, ts *testState) {
	t.Helper()

	rid := "0000"
	defer ts.redisClient.Del(context.Background(), rid)

	data := &apis.CancelRequest{
		UserId:    "user1",
		RequestId: rid,
	}
	_, err := ts.c.Cancel(context.Background(), data)
	require.NoError(t, err)

	require.Equal(t, 0, <-ts.callbackChan)

	time.Sleep(10 * time.Millisecond)
	res := ts.redisClient.Get(context.Background(), rid)
	require.NoError(t, res.Err())
	require.Equal(t, "1", res.Val())
}

func testUpdateSell(t *testing.T, ts *testState) {
	t.Helper()

	rid := "0000"
	defer ts.redisClient.Del(context.Background(), rid)

	data := &apis.UpdateRequest{
		UserId:    "user1",
		RequestId: rid,
	}
	_, err := ts.c.UpdateSell(context.Background(), data)
	require.NoError(t, err)

	require.Equal(t, 0, <-ts.callbackChan)

	time.Sleep(10 * time.Millisecond)
	res := ts.redisClient.Get(context.Background(), rid)
	require.NoError(t, res.Err())
	require.Equal(t, "1", res.Val())
}

func testUpdateBuy(t *testing.T, ts *testState) {
	t.Helper()

	rid := "0000"
	defer ts.redisClient.Del(context.Background(), rid)

	data := &apis.UpdateRequest{
		UserId:    "user1",
		RequestId: rid,
	}
	_, err := ts.c.UpdateBuy(context.Background(), data)
	require.NoError(t, err)

	require.Equal(t, 0, <-ts.callbackChan)

	time.Sleep(10 * time.Millisecond)
	res := ts.redisClient.Get(context.Background(), rid)
	require.NoError(t, res.Err())
	require.Equal(t, "1", res.Val())
}
