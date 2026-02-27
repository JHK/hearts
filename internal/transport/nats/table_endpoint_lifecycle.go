package natswire

import (
	"fmt"

	"github.com/JHK/hearts/internal/protocol"
	"github.com/nats-io/nats.go"
)

func (e *TableEndpoint) Register() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.registered {
		return nil
	}
	if e.tableID == "" {
		return fmt.Errorf("table id is required")
	}
	if e.nc == nil {
		return fmt.Errorf("nats connection is required")
	}

	specs := []struct {
		subject string
		handler nats.MsgHandler
	}{
		{subject: protocol.DiscoverSubject(), handler: e.handleDiscover},
		{subject: protocol.JoinSubject(e.tableID), handler: e.handleJoin},
		{subject: protocol.StartSubject(e.tableID), handler: e.handleStart},
		{subject: protocol.PlaySubject(e.tableID), handler: e.handlePlay},
	}

	created := make([]*nats.Subscription, 0, len(specs))
	for _, spec := range specs {
		sub, err := e.nc.Subscribe(spec.subject, spec.handler)
		if err != nil {
			for _, createdSub := range created {
				createdSub.Unsubscribe()
			}
			return err
		}
		created = append(created, sub)
	}

	e.subs = created
	e.registered = true

	return e.nc.Flush()
}

func (e *TableEndpoint) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, sub := range e.subs {
		sub.Unsubscribe()
	}
	e.subs = nil
	e.registered = false
}
