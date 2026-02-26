package bot

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/JHK/hearts/internal/protocol"
	"github.com/nats-io/nats.go"
)

const botIDPrefix = "bot-"

type Runtime struct {
	mu sync.Mutex

	nc      *nats.Conn
	tableID string
	rng     *rand.Rand

	eventsSub       *nats.Subscription
	playerEventsSub *nats.Subscription

	bots map[string]*botState
}

type botState struct {
	hand    []string
	playing bool
}

func NewRuntime(nc *nats.Conn, tableID string, rng *rand.Rand) *Runtime {
	if rng == nil {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	return &Runtime{
		nc:      nc,
		tableID: tableID,
		rng:     rng,
		bots:    make(map[string]*botState),
	}
}

func (r *Runtime) Start() error {
	var err error
	r.eventsSub, err = r.nc.Subscribe(protocol.EventsSubject(r.tableID), r.handleTableEvent)
	if err != nil {
		return err
	}

	r.playerEventsSub, err = r.nc.Subscribe(protocol.PlayerEventsWildcardSubject(r.tableID), r.handlePlayerEvent)
	if err != nil {
		r.eventsSub.Unsubscribe()
		r.eventsSub = nil
		return err
	}

	return r.nc.Flush()
}

func (r *Runtime) Stop() {
	if r.eventsSub != nil {
		r.eventsSub.Unsubscribe()
		r.eventsSub = nil
	}
	if r.playerEventsSub != nil {
		r.playerEventsSub.Unsubscribe()
		r.playerEventsSub = nil
	}
}

func (r *Runtime) handleTableEvent(msg *nats.Msg) {
	var event protocol.Event
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		return
	}

	switch event.Type {
	case protocol.EventPlayerJoined:
		var data protocol.PlayerJoinedData
		if json.Unmarshal(event.Data, &data) != nil {
			return
		}
		r.registerBot(data.Player.PlayerID)
	}
}

func (r *Runtime) handlePlayerEvent(msg *nats.Msg) {
	playerID, ok := playerIDFromSubject(msg.Subject)
	if !ok || !isBotID(playerID) {
		return
	}

	var event protocol.Event
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		return
	}

	if event.Type != protocol.EventHandUpdated {
		if event.Type != protocol.EventYourTurn {
			return
		}
		r.trigger(playerID)
		return
	}

	var data protocol.HandUpdatedData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return
	}

	r.updateHand(playerID, data.Cards)
}

func (r *Runtime) registerBot(playerID string) {
	if !isBotID(playerID) {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.bots[playerID]; exists {
		return
	}

	r.bots[playerID] = &botState{}
}

func (r *Runtime) updateHand(playerID string, cards []string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	bot, ok := r.bots[playerID]
	if !ok {
		bot = &botState{}
		r.bots[playerID] = bot
	}

	bot.hand = append(bot.hand[:0], cards...)
}

func (r *Runtime) trigger(playerID string) {
	r.mu.Lock()
	bot, ok := r.bots[playerID]
	if !ok || bot.playing || len(bot.hand) == 0 {
		r.mu.Unlock()
		return
	}
	bot.playing = true
	r.mu.Unlock()

	go r.playLoop(playerID)
}

func (r *Runtime) playLoop(playerID string) {
	defer r.clearPlaying(playerID)

	for {
		hand, ok := r.snapshotTurnHand(playerID)
		if !ok {
			return
		}

		card := hand[r.rng.Intn(len(hand))]
		accepted, reason, err := r.playCard(playerID, card)
		if err != nil {
			slog.Warn("bot play request failed", "player_id", playerID, "error", err)
			time.Sleep(50 * time.Millisecond)
			continue
		}

		if accepted {
			return
		}
		if reason == "not your turn" || reason == "round has not started" {
			return
		}

		slog.Debug("bot suggested illegal card, retrying", "player_id", playerID, "card", card)
	}
}

func (r *Runtime) snapshotTurnHand(playerID string) ([]string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	bot, ok := r.bots[playerID]
	if !ok || len(bot.hand) == 0 {
		return nil, false
	}

	hand := make([]string, len(bot.hand))
	copy(hand, bot.hand)
	return hand, true
}

func (r *Runtime) clearPlaying(playerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if bot, ok := r.bots[playerID]; ok {
		bot.playing = false
	}
}

func (r *Runtime) playCard(playerID, card string) (bool, string, error) {
	req := protocol.PlayCardRequest{PlayerID: playerID, Card: card}
	payload, _ := json.Marshal(req)

	msg, err := r.nc.Request(protocol.PlaySubject(r.tableID), payload, 2*time.Second)
	if err != nil {
		return false, "", err
	}

	var resp protocol.CommandResponse
	if err := json.Unmarshal(msg.Data, &resp); err != nil {
		return false, "", fmt.Errorf("decode play response: %w", err)
	}

	return resp.Accepted, resp.Reason, nil
}

func isBotID(playerID string) bool {
	return strings.HasPrefix(playerID, botIDPrefix)
}

func playerIDFromSubject(subject string) (string, bool) {
	parts := strings.Split(subject, ".")
	if len(parts) != 6 {
		return "", false
	}
	if parts[0] != "hearts" || parts[1] != "table" || parts[3] != "player" || parts[5] != "events" {
		return "", false
	}

	return parts[4], true
}
