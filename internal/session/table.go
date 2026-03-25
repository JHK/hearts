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
	Paused             bool                                `json:"paused,omitempty"`
	PausedForPlayerID  protocol.PlayerID                   `json:"paused_for_player_id,omitempty"`
}

type Table struct {
	tableID string

	commands  chan any
	stop      chan struct{}
	stopOnce  sync.Once
	stoppedCh chan struct{}

	subsMu    sync.RWMutex
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

type botHandsCommand struct {
	reply chan []BotHandSnapshot
}

// BotHandSnapshot holds a bot's name, seat, and current hand for debugging.
type BotHandSnapshot struct {
	Name  string   `json:"name"`
	Seat  int      `json:"seat"`
	Cards []string `json:"cards"`
}

func NewTable(tableID string) *Table {
	r := &Table{
		tableID:   tableID,
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

		r.subsMu.Lock()
		for id, ch := range r.subs {
			close(ch)
			delete(r.subs, id)
		}
		r.subsMu.Unlock()
	})
}

func (r *Table) Subscribe() (<-chan StreamEvent, func()) {
	r.subsMu.Lock()
	r.nextSubID++
	id := r.nextSubID
	ch := make(chan StreamEvent, 128)
	r.subs[id] = ch
	r.subsMu.Unlock()

	unsubscribe := func() {
		r.subsMu.Lock()
		sub, ok := r.subs[id]
		if ok {
			delete(r.subs, id)
			close(sub)
		}
		r.subsMu.Unlock()
	}

	return ch, unsubscribe
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

// BotHands returns the name, seat, and current hand of every bot at the table.
// Intended for dev/debug use only; returns nil when the session is stopped.
func (r *Table) BotHands() []BotHandSnapshot {
	reply := make(chan []BotHandSnapshot, 1)
	if !r.submit(botHandsCommand{reply: reply}) {
		return nil
	}
	return <-reply
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
			case botTurnCommand:
				r.handleBotTurn(state, cmd.playerID)
			case botPassCommand:
				r.handleBotPass(state, cmd.playerID)
			case botHandsCommand:
				cmd.reply <- r.buildBotHands(state)
			}
		}
	}
}

func (r *Table) publishPublic(eventType protocol.EventType, payload any) {
	r.emit(StreamEvent{Type: eventType, Data: payload})
}

func (r *Table) publishPrivate(playerID protocol.PlayerID, eventType protocol.EventType, payload any) {
	r.emit(StreamEvent{Type: eventType, Data: payload, PrivateTo: playerID})
}

func (r *Table) emit(event StreamEvent) {
	r.subsMu.RLock()
	defer r.subsMu.RUnlock()

	for _, sub := range r.subs {
		select {
		case sub <- event:
		default:
		}
	}
}
