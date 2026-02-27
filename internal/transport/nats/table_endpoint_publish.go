package natswire

import (
	"fmt"

	"github.com/JHK/hearts/internal/game"
	"github.com/JHK/hearts/internal/protocol"
)

func (e *TableEndpoint) PublishPublic(eventType protocol.EventType, payload any) error {
	if e.nc == nil {
		return fmt.Errorf("nats connection is required")
	}

	data, err := protocol.EncodeEvent(eventType, payload)
	if err != nil {
		return err
	}

	return e.nc.Publish(protocol.EventsSubject(e.tableID), data)
}

func (e *TableEndpoint) PublishPrivate(playerID game.PlayerID, eventType protocol.EventType, payload any) error {
	if e.nc == nil {
		return fmt.Errorf("nats connection is required")
	}

	data, err := protocol.EncodeEvent(eventType, payload)
	if err != nil {
		return err
	}

	return e.nc.Publish(protocol.PlayerEventsSubject(e.tableID, playerID), data)
}
