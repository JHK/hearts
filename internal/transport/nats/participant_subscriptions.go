package natswire

import (
	"fmt"

	"github.com/JHK/hearts/internal/game"
	"github.com/JHK/hearts/internal/protocol"
	"github.com/nats-io/nats.go"
)

func (p *ParticipantClient) Start(handlers ParticipantEventHandlers) error {
	if p.nc == nil {
		return fmt.Errorf("nats connection is required")
	}
	if p.tableID == "" {
		return fmt.Errorf("table id is required")
	}

	p.mu.Lock()
	if len(p.subs) > 0 {
		p.mu.Unlock()
		return nil
	}
	p.handlers = handlers
	p.mu.Unlock()

	publicSub, err := p.nc.Subscribe(protocol.EventsSubject(p.tableID), func(msg *nats.Msg) {
		p.handlePublicEvent(msg.Data)
	})
	if err != nil {
		return err
	}

	privateSubject := protocol.PlayerEventsWildcardSubject(p.tableID)
	if p.playerID != "" {
		privateSubject = protocol.PlayerEventsSubject(p.tableID, p.playerID)
	}

	privateSub, err := p.nc.Subscribe(privateSubject, func(msg *nats.Msg) {
		p.handlePrivateEvent(msg.Subject, msg.Data)
	})
	if err != nil {
		publicSub.Unsubscribe()
		return err
	}

	p.mu.Lock()
	p.subs = []*nats.Subscription{publicSub, privateSub}
	p.mu.Unlock()

	return p.nc.Flush()
}

func (p *ParticipantClient) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, sub := range p.subs {
		sub.Unsubscribe()
	}
	p.subs = nil
	p.handlers = ParticipantEventHandlers{}
}

func (p *ParticipantClient) playerIDFromPrivateSubject(subject string) (game.PlayerID, bool) {
	if p.playerID != "" {
		expected := protocol.PlayerEventsSubject(p.tableID, p.playerID)
		if subject != expected {
			return game.PlayerID(""), false
		}
		return p.playerID, true
	}

	return protocol.ParsePlayerEventsSubject(p.tableID, subject)
}
