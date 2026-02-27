package natswire

import (
	"sync"

	"github.com/JHK/hearts/internal/protocol"
	"github.com/nats-io/nats.go"
)

type TableEndpointHandlers struct {
	OnDiscover func() protocol.DiscoverResponse
	OnJoin     func(protocol.JoinRequest) protocol.JoinResponse
	OnStart    func(protocol.StartRequest) protocol.CommandResponse
	OnPlay     func(protocol.PlayCardRequest) protocol.CommandResponse
}

type TableEndpoint struct {
	nc       *nats.Conn
	tableID  string
	handlers TableEndpointHandlers

	mu         sync.Mutex
	registered bool
	subs       []*nats.Subscription
}

func NewTableEndpoint(nc *nats.Conn, tableID string, handlers TableEndpointHandlers) *TableEndpoint {
	return &TableEndpoint{nc: nc, tableID: tableID, handlers: handlers}
}
