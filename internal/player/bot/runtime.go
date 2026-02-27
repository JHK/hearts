package bot

import (
	"sync"

	"github.com/JHK/hearts/internal/game"
	"github.com/JHK/hearts/internal/protocol"
	natswire "github.com/JHK/hearts/internal/transport/nats"
	"github.com/nats-io/nats.go"
)

type Runtime struct {
	playerID    game.PlayerID
	participant *natswire.ParticipantClient
	strategy    Strategy

	events chan any
	stop   chan struct{}
	done   chan struct{}

	stopOnce sync.Once
	doneOnce sync.Once
}

type runtimeState struct {
	started      bool
	heartsBroken bool
	trickNumber  int
	trickCards   []game.Card
	playInFlight bool
	hand         []game.Card
}

type gameStartedEvent struct{}

type turnChangedEvent struct {
	payload protocol.TurnChangedData
}

type cardPlayedEvent struct {
	payload protocol.CardPlayedData
}

type trickCompletedEvent struct{}

type roundCompletedEvent struct{}

type handUpdatedEvent struct {
	payload protocol.HandUpdatedData
}

type yourTurnEvent struct{}

type playFinishedEvent struct{}

func NewRuntime(nc *nats.Conn, tableID string, playerID game.PlayerID, strategy Strategy) *Runtime {
	if strategy == nil {
		strategy = NewRandomBot(nil)
	}

	return &Runtime{
		playerID:    playerID,
		participant: natswire.NewParticipantClient(nc, tableID, playerID),
		strategy:    strategy,
		events:      make(chan any, 128),
		stop:        make(chan struct{}),
		done:        make(chan struct{}),
	}
}

func (r *Runtime) Start() error {
	if err := r.participant.Start(natswire.ParticipantEventHandlers{
		OnGameStarted: func() {
			r.enqueue(gameStartedEvent{})
		},
		OnTurnChanged: r.onTurnChanged,
		OnCardPlayed:  r.onCardPlayed,
		OnTrickCompleted: func(protocol.TrickCompletedData) {
			r.enqueue(trickCompletedEvent{})
		},
		OnRoundCompleted: func(protocol.RoundCompletedData) {
			r.enqueue(roundCompletedEvent{})
		},
		OnHandUpdated: r.onHandUpdated,
		OnYourTurn:    r.onYourTurn,
	}); err != nil {
		r.markDone()
		return err
	}

	go r.run()
	return nil
}

func (r *Runtime) Stop() {
	r.stopOnce.Do(func() {
		r.participant.Stop()
		close(r.stop)
	})
	<-r.done
}

func (r *Runtime) onTurnChanged(payload protocol.TurnChangedData) {
	r.enqueue(turnChangedEvent{payload: payload})
}

func (r *Runtime) onCardPlayed(payload protocol.CardPlayedData) {
	r.enqueue(cardPlayedEvent{payload: payload})
}

func (r *Runtime) onHandUpdated(_ game.PlayerID, payload protocol.HandUpdatedData) {
	r.enqueue(handUpdatedEvent{payload: payload})
}

func (r *Runtime) onYourTurn(_ game.PlayerID, _ protocol.YourTurnData) {
	r.enqueue(yourTurnEvent{})
}

func (r *Runtime) enqueue(event any) {
	select {
	case <-r.stop:
		return
	case r.events <- event:
	}
}

func (r *Runtime) run() {
	defer r.markDone()

	state := runtimeState{}

	for {
		select {
		case <-r.stop:
			return
		case event := <-r.events:
			switch typed := event.(type) {
			case gameStartedEvent:
				state.started = true
				state.heartsBroken = false
				state.trickNumber = 0
				state.trickCards = nil
				state.playInFlight = false
			case turnChangedEvent:
				state.trickNumber = typed.payload.TrickNumber
			case cardPlayedEvent:
				card, err := game.ParseCard(typed.payload.Card)
				if err != nil {
					continue
				}
				state.trickCards = append(state.trickCards, card)
				if card.Suit == game.SuitHearts {
					state.heartsBroken = true
				}
			case trickCompletedEvent:
				state.trickCards = nil
			case roundCompletedEvent:
				state.started = false
				state.trickCards = nil
				state.playInFlight = false
				state.hand = nil
			case handUpdatedEvent:
				hand, err := game.ParseCards(typed.payload.Cards)
				if err != nil {
					continue
				}
				state.hand = hand
			case yourTurnEvent:
				r.triggerPlay(&state)
			case playFinishedEvent:
				state.playInFlight = false
			}
		}
	}
}

func (r *Runtime) markDone() {
	r.doneOnce.Do(func() {
		close(r.done)
	})
}

func (r *Runtime) triggerPlay(state *runtimeState) {
	if !state.started || state.playInFlight {
		return
	}

	hand := append([]game.Card(nil), state.hand...)
	if len(hand) == 0 {
		return
	}

	state.playInFlight = true

	input := TurnInput{
		Hand:         hand,
		Trick:        append([]game.Card(nil), state.trickCards...),
		HeartsBroken: state.heartsBroken,
		FirstTrick:   state.trickNumber == 0,
	}

	go r.playTurn(input)
}

func (r *Runtime) playTurn(input TurnInput) {
	defer r.enqueue(playFinishedEvent{})

	card, err := r.strategy.ChoosePlay(input)
	if err != nil {
		return
	}

	_ = r.participant.PlayRequest(protocol.PlayCardRequest{PlayerID: r.playerID, Card: card.String()})
}
