package engine

import (
	"github.com/atgane/opentd/apis"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type Engine interface {
	AddBuy(e cloudevents.Event) error
	AddSell(e cloudevents.Event) error
	AddCancel(e cloudevents.Event) error
	AddUpdateBuy(e cloudevents.Event) error
	AddUpdateSell(e cloudevents.Event) error
	Start(Snapshot, stream func(apis.GetDealStream) (cloudevents.Event, error)) error
}
