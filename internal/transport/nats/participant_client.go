package natswire

import (
	"fmt"
	"sync"
	"time"

	"github.com/JHK/hearts/internal/game"
	"github.com/JHK/hearts/internal/protocol"
	"github.com/nats-io/nats.go"
)

const defaultRequestTimeout = 2 * time.Second

type ParticipantEventHandlers struct {
	OnPlayerJoined   func(protocol.PlayerJoinedData)
	OnGameStarted    func()
	OnTurnChanged    func(protocol.TurnChangedData)
	OnCardPlayed     func(protocol.CardPlayedData)
	OnTrickCompleted func(protocol.TrickCompletedData)
	OnRoundCompleted func(protocol.RoundCompletedData)
	OnHandUpdated    func(game.PlayerID, protocol.HandUpdatedData)
	OnYourTurn       func(game.PlayerID, protocol.YourTurnData)
	OnDecodeError    func(error)
}

type ParticipantClient struct {
	nc       *nats.Conn
	tableID  string
	playerID game.PlayerID

	mu       sync.Mutex
	handlers ParticipantEventHandlers
	subs     []*nats.Subscription
}

func NewParticipantClient(nc *nats.Conn, tableID string, playerID game.PlayerID) *ParticipantClient {
	return &ParticipantClient{nc: nc, tableID: tableID, playerID: playerID}
}

func (p *ParticipantClient) request(subject string, request any, out any) error {
	if p.nc == nil {
		return fmt.Errorf("nats connection is required")
	}
	if p.tableID == "" {
		return fmt.Errorf("table id is required")
	}

	payload, err := marshalJSON(request)
	if err != nil {
		return err
	}

	msg, err := p.nc.Request(subject, payload, defaultRequestTimeout)
	if err != nil {
		return err
	}

	if err := unmarshalJSON(msg.Data, out); err != nil {
		return err
	}

	return nil
}

func (p *ParticipantClient) emitDecodeError(err error) {
	handlers := p.snapshotHandlers()
	if handlers == nil || handlers.OnDecodeError == nil {
		return
	}
	handlers.OnDecodeError(err)
}

func (p *ParticipantClient) snapshotHandlers() *ParticipantEventHandlers {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.subs) == 0 {
		return nil
	}

	copy := p.handlers
	return &copy
}
