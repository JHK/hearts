package session

import (
	"sync"

	"github.com/JHK/hearts/internal/game"
	"github.com/JHK/hearts/internal/game/bot"
	"github.com/JHK/hearts/internal/protocol"
)

type StreamEvent struct {
	Type      protocol.EventType
	Data      any
	PrivateTo protocol.PlayerID
}

type AddedBot struct {
	JoinResponse protocol.JoinResponse `json:"join_response"`
	Name         string                `json:"name"`
	Strategy     string                `json:"strategy"`
}

type PlayerSnapshot struct {
	PlayerID protocol.PlayerID `json:"player_id"`
	Name     string            `json:"name"`
	Seat     int               `json:"seat"`
	IsBot    bool              `json:"is_bot"`
}

type TrickPlaySnapshot struct {
	PlayerID protocol.PlayerID `json:"player_id"`
	Name     string            `json:"name"`
	Seat     int               `json:"seat"`
	Card     string            `json:"card"`
}

type Snapshot struct {
	TableID            string                              `json:"table_id"`
	Players            []PlayerSnapshot                    `json:"players"`
	Started            bool                                `json:"started"`
	Phase              string                              `json:"phase"`
	TrickNumber        int                                 `json:"trick_number"`
	TurnPlayerID       protocol.PlayerID                   `json:"turn_player_id"`
	HeartsBroken       bool                                `json:"hearts_broken"`
	CurrentTrick       []string                            `json:"current_trick"`
	TrickPlays         []TrickPlaySnapshot                 `json:"trick_plays"`
	Hand               []string                            `json:"hand"`
	HandSizes          map[protocol.PlayerID]int           `json:"hand_sizes"`
	PassDirection      game.PassDirection                  `json:"pass_direction"`
	PassSubmitted      bool                                `json:"pass_submitted"`
	PassSubmittedCount int                                 `json:"pass_submitted_count"`
	PassSent           []string                            `json:"pass_sent"`
	PassReceived       []string                            `json:"pass_received"`
	PassReady          bool                                `json:"pass_ready"`
	PassReadyCount     int                                 `json:"pass_ready_count"`
	RoundPoints        map[protocol.PlayerID]game.Points   `json:"round_points"`
	RoundHistory       []map[protocol.PlayerID]game.Points `json:"round_history"`
	TotalPoints        map[protocol.PlayerID]game.Points   `json:"total_points"`
	GameOver           bool                                `json:"game_over"`
	Winners            []protocol.PlayerID                 `json:"winners,omitempty"`
	RematchVotes       int                                 `json:"rematch_votes,omitempty"`
	RematchTotal       int                                 `json:"rematch_total,omitempty"`
	RematchVoted       bool                                `json:"rematch_voted,omitempty"`
	Paused             bool                                `json:"paused,omitempty"`
	PausedForPlayerID  protocol.PlayerID                   `json:"paused_for_player_id,omitempty"`
}

type Table struct {
	tableID  string
	onChange func() // called when lobby-visible state changes (join, leave, start, etc.)

	commands  chan any
	stop      chan struct{}
	stopOnce  sync.Once
	stoppedCh chan struct{}

	subs      map[int]chan StreamEvent
	nextSubID int
}

type tableState struct {
	players        []*playerState
	playersByID    map[protocol.PlayerID]*playerState
	playersByToken map[string]*playerState
	departedTokens map[string]protocol.PlayerID // token → last player ID for pre-round leavers
	roundHistory   []map[protocol.PlayerID]game.Points

	round         *game.Round
	game          *game.Game
	nextPlayerSeq int
	gameOver      bool

	// paused is set when a human disconnects mid-round. While true, all game
	// commands and bot actions are blocked until a remaining human resumes.
	paused         bool
	pausedPlayerID protocol.PlayerID // player who disconnected

	// rematchVotes tracks which human players have voted for a rematch.
	rematchVotes map[protocol.PlayerID]bool
}

// playerState holds the seat-level data for one player.
//
// Game state lives in tableState.round (the game.Round coordinator).
// playerState only carries identity/transport and cumulative scoring.
//
// Bot detection: player.bot != nil (set for bots, nil for humans)
type playerState struct {
	// web-transport state
	id    protocol.PlayerID
	Token string

	// seated identity
	Name     string
	position int // ordinal seat number (0–3)

	// bot handles autonomous decisions; nil for human players.
	// Human turns are driven by Play/Pass commands from the WebSocket connection.
	bot bot.Bot
}

type joinCommand struct {
	name  string
	token string
	reply chan protocol.JoinResponse
}

type startCommand struct {
	playerID protocol.PlayerID
	reply    chan protocol.CommandResponse
}

type playCommand struct {
	playerID protocol.PlayerID
	card     string
	reply    chan protocol.CommandResponse
}

type passCommand struct {
	playerID protocol.PlayerID
	cards    []string
	reply    chan protocol.CommandResponse
}

type readyAfterPassCommand struct {
	playerID protocol.PlayerID
	reply    chan protocol.CommandResponse
}

type addBotCommand struct {
	strategyName string
	reply        chan addBotResult
}

type addBotResult struct {
	added AddedBot
	err   error
}

type snapshotCommand struct {
	forPlayer protocol.PlayerID
	reply     chan Snapshot
}

type leaveCommand struct {
	playerID protocol.PlayerID
	reply    chan struct{}
}

type infoCommand struct {
	reply chan protocol.TableInfo
}

type botTurnCommand struct {
	playerID protocol.PlayerID
}

type botPassCommand struct {
	playerID protocol.PlayerID
}

type resumeGameCommand struct {
	playerID protocol.PlayerID
	reply    chan protocol.CommandResponse
}

type rematchCommand struct {
	playerID protocol.PlayerID
	reply    chan protocol.CommandResponse
}

type claimSeatCommand struct {
	seat  int
	name  string
	token string
	reply chan protocol.JoinResponse
}

type renameCommand struct {
	playerID protocol.PlayerID
	name     string
	reply    chan protocol.CommandResponse
}

type debugBotCommand struct {
	reply chan *DebugBotSnapshot
}

type subscribeCommand struct {
	reply chan subscribeResult
}

type subscribeResult struct {
	ch <-chan StreamEvent
	id int
}

type unsubscribeCommand struct {
	id int
}

// DebugBotSnapshot holds the full decision context for all bots at the table.
type DebugBotSnapshot struct {
	TableID       string                 `json:"table_id"`
	Phase         string                 `json:"phase"`
	TrickNumber   int                    `json:"trick_number"`
	HeartsBroken  bool                   `json:"hearts_broken"`
	FirstTrick    bool                   `json:"first_trick"`
	LedSuit       game.Suit              `json:"led_suit,omitempty"`
	PassDirection game.PassDirection     `json:"pass_direction,omitempty"`
	TurnSeat      *int                   `json:"turn_seat,omitempty"`
	TurnPlayer    string                 `json:"turn_player,omitempty"`
	CurrentTrick  []TrickPlaySnapshot    `json:"current_trick"`
	PlayedCards   []string               `json:"played_cards"`
	Players       []string               `json:"players"`
	RoundPoints   map[string]game.Points `json:"round_points"`
	TotalPoints   map[string]game.Points `json:"total_points"`
	Bots          []BotSnapshot          `json:"bots"`
}

// BotSnapshot holds a single bot's decision context for debugging.
type BotSnapshot struct {
	Name            string   `json:"name"`
	Seat            int      `json:"seat"`
	Strategy        string   `json:"strategy"`
	Hand            []string `json:"hand"`
	MoonShotActive  *bool    `json:"moon_shot_active,omitempty"`
	MoonShotAborted *bool    `json:"moon_shot_aborted,omitempty"`
}

func NewTable(tableID string, onChange func()) *Table {
	r := &Table{
		tableID:   tableID,
		onChange:  onChange,
		commands:  make(chan any),
		stop:      make(chan struct{}),
		stoppedCh: make(chan struct{}),
		subs:      make(map[int]chan StreamEvent),
	}

	go r.run()
	return r
}

func (r *Table) ID() string {
	return r.tableID
}

func (r *Table) Close() {
	r.stopOnce.Do(func() {
		close(r.stop)
		<-r.stoppedCh

		// Safe without lock: the actor goroutine has exited.
		for id, ch := range r.subs {
			close(ch)
			delete(r.subs, id)
		}
	})
}

// Subscribe returns a channel of events and an unsubscribe function.
// It must not be called from the actor goroutine (it would deadlock).
func (r *Table) Subscribe() (<-chan StreamEvent, func()) {
	reply := make(chan subscribeResult, 1)
	if !r.submit(subscribeCommand{reply: reply}) {
		// Table is stopping; return a closed channel so the caller sees EOF.
		ch := make(chan StreamEvent)
		close(ch)
		return ch, func() {}
	}
	res := <-reply

	unsubscribe := func() {
		// Fire-and-forget; if the table is already stopping the actor
		// will never process this, but Close drains subs anyway.
		r.submit(unsubscribeCommand{id: res.id})
	}

	return res.ch, unsubscribe
}

func (r *Table) Join(name, token string) (protocol.JoinResponse, error) {
	reply := make(chan protocol.JoinResponse, 1)
	if !r.submit(joinCommand{name: name, token: token, reply: reply}) {
		return protocol.JoinResponse{}, ErrTableStopping
	}
	return <-reply, nil
}

func (r *Table) Start(playerID protocol.PlayerID) protocol.CommandResponse {
	reply := make(chan protocol.CommandResponse, 1)
	if !r.submit(startCommand{playerID: playerID, reply: reply}) {
		return protocol.CommandResponse{Accepted: false, Reason: "table is stopping"}
	}
	return <-reply
}

func (r *Table) Play(playerID protocol.PlayerID, card string) protocol.CommandResponse {
	reply := make(chan protocol.CommandResponse, 1)
	if !r.submit(playCommand{playerID: playerID, card: card, reply: reply}) {
		return protocol.CommandResponse{Accepted: false, Reason: "table is stopping"}
	}
	return <-reply
}

func (r *Table) Pass(playerID protocol.PlayerID, cards []string) protocol.CommandResponse {
	reply := make(chan protocol.CommandResponse, 1)
	if !r.submit(passCommand{playerID: playerID, cards: cards, reply: reply}) {
		return protocol.CommandResponse{Accepted: false, Reason: "table is stopping"}
	}
	return <-reply
}

func (r *Table) ReadyAfterPass(playerID protocol.PlayerID) protocol.CommandResponse {
	reply := make(chan protocol.CommandResponse, 1)
	if !r.submit(readyAfterPassCommand{playerID: playerID, reply: reply}) {
		return protocol.CommandResponse{Accepted: false, Reason: "table is stopping"}
	}
	return <-reply
}

func (r *Table) AddBot(strategyName string) (AddedBot, error) {
	reply := make(chan addBotResult, 1)
	if !r.submit(addBotCommand{strategyName: strategyName, reply: reply}) {
		return AddedBot{}, ErrTableStopping
	}
	result := <-reply
	return result.added, result.err
}

func (r *Table) Snapshot(forPlayer protocol.PlayerID) Snapshot {
	reply := make(chan Snapshot, 1)
	if !r.submit(snapshotCommand{forPlayer: forPlayer, reply: reply}) {
		return Snapshot{TableID: r.tableID}
	}
	return <-reply
}

func (r *Table) Leave(playerID protocol.PlayerID) {
	if playerID == "" {
		return
	}

	reply := make(chan struct{}, 1)
	if !r.submit(leaveCommand{playerID: playerID, reply: reply}) {
		return
	}
	<-reply
}

func (r *Table) ResumeGame(playerID protocol.PlayerID) protocol.CommandResponse {
	reply := make(chan protocol.CommandResponse, 1)
	if !r.submit(resumeGameCommand{playerID: playerID, reply: reply}) {
		return protocol.CommandResponse{Accepted: false, Reason: "table is stopping"}
	}
	return <-reply
}

func (r *Table) Info() protocol.TableInfo {
	reply := make(chan protocol.TableInfo, 1)
	if !r.submit(infoCommand{reply: reply}) {
		return protocol.TableInfo{TableID: r.tableID, MaxPlayers: game.PlayersPerTable}
	}
	return <-reply
}

func (r *Table) Rematch(playerID protocol.PlayerID) protocol.CommandResponse {
	reply := make(chan protocol.CommandResponse, 1)
	if !r.submit(rematchCommand{playerID: playerID, reply: reply}) {
		return protocol.CommandResponse{Accepted: false, Reason: "table is stopping"}
	}
	return <-reply
}

func (r *Table) Rename(playerID protocol.PlayerID, name string) protocol.CommandResponse {
	reply := make(chan protocol.CommandResponse, 1)
	if !r.submit(renameCommand{playerID: playerID, name: name, reply: reply}) {
		return protocol.CommandResponse{Accepted: false, Reason: "table is stopping"}
	}
	return <-reply
}

// ClaimSeat lets an observer take over a bot-controlled seat.
func (r *Table) ClaimSeat(seat int, name, token string) (protocol.JoinResponse, error) {
	reply := make(chan protocol.JoinResponse, 1)
	if !r.submit(claimSeatCommand{seat: seat, name: name, token: token, reply: reply}) {
		return protocol.JoinResponse{}, ErrTableStopping
	}
	return <-reply, nil
}

// DebugBotContext returns the full decision context for all bots at the table.
// Intended for dev/debug use only; returns nil when the session is stopped.
func (r *Table) DebugBotContext() *DebugBotSnapshot {
	reply := make(chan *DebugBotSnapshot, 1)
	if !r.submit(debugBotCommand{reply: reply}) {
		return nil
	}
	return <-reply
}

// Drain is a test helper that gives pending fire-and-forget commands
// (e.g. scheduled bot actions) a chance to be processed. It performs 10
// synchronous round-trips through the actor's unbuffered channel. This is
// a best-effort heuristic, not a formal guarantee — cascading actions that
// spawn more than 10 async commands may not fully settle.
func (r *Table) Drain() {
	for range 10 {
		r.Snapshot("")
	}
}

func (r *Table) submit(command any) bool {
	select {
	case <-r.stop:
		return false
	case r.commands <- command:
		return true
	}
}

func (r *Table) run() {
	defer close(r.stoppedCh)

	state := &tableState{
		playersByID:    make(map[protocol.PlayerID]*playerState),
		playersByToken: make(map[string]*playerState),
		departedTokens: make(map[string]protocol.PlayerID),
		game:           game.NewGame(),
	}

	for {
		select {
		case <-r.stop:
			return
		case command := <-r.commands:
			switch cmd := command.(type) {
			case joinCommand:
				cmd.reply <- r.handleJoin(state, cmd.name, cmd.token)
			case startCommand:
				cmd.reply <- r.handleStart(state, cmd.playerID)
			case playCommand:
				cmd.reply <- r.handlePlay(state, cmd.playerID, cmd.card)
			case passCommand:
				cmd.reply <- r.handlePass(state, cmd.playerID, cmd.cards)
			case readyAfterPassCommand:
				cmd.reply <- r.handleReadyAfterPass(state, cmd.playerID)
			case addBotCommand:
				added, err := r.handleAddBot(state, cmd.strategyName)
				cmd.reply <- addBotResult{added: added, err: err}
			case snapshotCommand:
				cmd.reply <- r.buildSnapshot(state, cmd.forPlayer)
			case leaveCommand:
				r.handleLeave(state, cmd.playerID)
				cmd.reply <- struct{}{}
			case infoCommand:
				cmd.reply <- protocol.TableInfo{
					TableID:    r.tableID,
					Players:    len(state.players),
					MaxPlayers: game.PlayersPerTable,
					Started:    state.round != nil,
					GameOver:   state.gameOver,
					Paused:     state.paused,
				}
			case resumeGameCommand:
				cmd.reply <- r.handleResumeGame(state, cmd.playerID)
			case rematchCommand:
				cmd.reply <- r.handleRematch(state, cmd.playerID)
			case botTurnCommand:
				r.handleBotTurn(state, cmd.playerID)
			case botPassCommand:
				r.handleBotPass(state, cmd.playerID)
			case renameCommand:
				cmd.reply <- r.handleRename(state, cmd.playerID, cmd.name)
			case claimSeatCommand:
				cmd.reply <- r.handleClaimSeat(state, cmd.seat, cmd.name, cmd.token)
			case debugBotCommand:
				cmd.reply <- r.buildDebugBotContext(state)
			case subscribeCommand:
				r.nextSubID++
				id := r.nextSubID
				ch := make(chan StreamEvent, 128)
				r.subs[id] = ch
				cmd.reply <- subscribeResult{ch: ch, id: id}
			case unsubscribeCommand:
				if sub, ok := r.subs[cmd.id]; ok {
					close(sub)
					delete(r.subs, cmd.id)
				}
			}
		}
	}
}

// lobbyRelevantEvents are event types that change how a table appears in the lobby list.
var lobbyRelevantEvents = map[protocol.EventType]bool{
	protocol.EventPlayerJoined: true,
	protocol.EventPlayerLeft:   true,
	protocol.EventGameStarted:  true,
	protocol.EventGameOver:     true,
	protocol.EventGamePaused:   true,
	protocol.EventGameResumed:  true,
}

func (r *Table) publishPublic(eventType protocol.EventType, payload any) {
	r.emit(StreamEvent{Type: eventType, Data: payload})
	if r.onChange != nil && lobbyRelevantEvents[eventType] {
		r.onChange()
	}
}

func (r *Table) publishPrivate(playerID protocol.PlayerID, eventType protocol.EventType, payload any) {
	r.emit(StreamEvent{Type: eventType, Data: payload, PrivateTo: playerID})
}

func (r *Table) emit(event StreamEvent) {
	for _, sub := range r.subs {
		select {
		case sub <- event:
		default:
		}
	}
}
