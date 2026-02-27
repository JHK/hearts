package natswire

import (
	"fmt"
	"sort"
	"time"

	"github.com/JHK/hearts/internal/protocol"
	"github.com/nats-io/nats.go"
)

func (p *ParticipantClient) Join(name string) (protocol.JoinResponse, error) {
	return p.JoinRequest(protocol.JoinRequest{Name: name})
}

func (p *ParticipantClient) JoinRequest(request protocol.JoinRequest) (protocol.JoinResponse, error) {
	var response protocol.JoinResponse
	if err := p.request(protocol.JoinSubject(p.tableID), request, &response); err != nil {
		return protocol.JoinResponse{}, err
	}
	return response, nil
}

func DiscoverTables(nc *nats.Conn, timeout time.Duration) ([]protocol.TableInfo, error) {
	if nc == nil {
		return nil, fmt.Errorf("nats connection is required")
	}
	if timeout <= 0 {
		timeout = 750 * time.Millisecond
	}

	inbox := nats.NewInbox()
	sub, err := nc.SubscribeSync(inbox)
	if err != nil {
		return nil, err
	}
	defer sub.Unsubscribe()

	payload, err := marshalJSON(protocol.DiscoverRequest{})
	if err != nil {
		return nil, err
	}

	if err := nc.PublishRequest(protocol.DiscoverSubject(), inbox, payload); err != nil {
		return nil, err
	}
	if err := nc.Flush(); err != nil {
		return nil, err
	}

	deadline := time.Now().Add(timeout)
	tables := make([]protocol.TableInfo, 0)
	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			break
		}

		msg, err := sub.NextMsg(remaining)
		if err == nats.ErrTimeout {
			break
		}
		if err != nil {
			return nil, err
		}

		var response protocol.DiscoverResponse
		if err := unmarshalJSON(msg.Data, &response); err != nil {
			continue
		}
		tables = append(tables, response.Tables...)
	}

	sort.Slice(tables, func(i, j int) bool {
		return tables[i].TableID < tables[j].TableID
	})

	return tables, nil
}
