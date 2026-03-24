package session

import (
	"fmt"
	"log/slog"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

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
	roundsStarted int
	nextPlayerSeq int
	gameOver      bool
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

	// cumulative scoring (persists across rounds)
	cumulativePoints game.Points
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
		return protocol.JoinResponse{}, fmt.Errorf("table is stopping")
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
		return AddedBot{}, fmt.Errorf("table is stopping")
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
				}
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

func (r *Table) handleLeave(state *tableState, playerID protocol.PlayerID) {
	player := state.playersByID[playerID]
	if player == nil {
		return
	}

	slog.Info("player left table", "event", "player_left", "table_id", r.tableID, "player_id", playerID, "name", player.Name)

	if state.round != nil {
		// Preserve the token so the player can reclaim their seat if they reconnect.
		// Assign a bot; game state stays in the Round untouched.
		if player.bot == nil {
			player.bot = bot.StrategyRandom.NewBot()
		}

		switch state.round.Phase() {
		case game.PhasePassing:
			if !state.round.HasSubmittedPass(player.position) {
				input := state.round.PassInput(player.position)
				cards, err := player.bot.ChoosePass(input)
				if err == nil {
					_ = r.handlePass(state, playerID, game.CardStrings(cards))
				}
			}
		case game.PhasePassReview:
			if !state.round.IsPassReady(player.position) {
				_ = r.handleReadyAfterPass(state, playerID)
			}
		case game.PhasePlaying:
			if player.position == state.round.TurnSeat() {
				r.scheduleBotTurn(state, player)
			}
		}
		return
	}

	delete(state.playersByID, playerID)
	if player.Token != "" {
		delete(state.playersByToken, player.Token)
		state.departedTokens[player.Token] = playerID
	}

	for index, seated := range state.players {
		if seated.id != playerID {
			continue
		}

		state.players = append(state.players[:index], state.players[index+1:]...)
		for seat, updated := range state.players {
			updated.position = seat
		}
		return
	}
}

func (r *Table) handleJoin(state *tableState, name, token string) protocol.JoinResponse {
	token = strings.TrimSpace(token)
	if token == "" {
		return protocol.JoinResponse{Accepted: false, Reason: "player token is required"}
	}

	if existing := state.playersByToken[token]; existing != nil {
		if existing.bot != nil {
			// Player reconnected after being converted to a bot mid-round — reclaim the seat.
			existing.bot = nil
			if n := strings.TrimSpace(name); n != "" {
				existing.Name = n
			}
			slog.Info("player reclaimed seat", "event", "player_reclaimed", "table_id", r.tableID, "player_id", existing.id, "name", existing.Name, "seat", existing.position)
		}
		return protocol.JoinResponse{
			Accepted: true,
			TableID:  r.tableID,
			PlayerID: existing.id,
			Seat:     existing.position,
		}
	}

	if state.round != nil {
		return protocol.JoinResponse{Accepted: false, Reason: "round already in progress"}
	}
	if len(state.players) >= game.PlayersPerTable {
		return protocol.JoinResponse{Accepted: false, Reason: "table is full"}
	}

	name = strings.TrimSpace(name)
	if name == "" {
		name = "Player"
	}

	id, reused := state.departedTokens[token]
	if reused {
		delete(state.departedTokens, token)
	} else {
		id = r.nextPlayerID(state)
	}
	player := r.addPlayer(state, id, name, nil, token)

	slog.Info("player joined table", "event", "player_joined", "table_id", r.tableID, "player_id", player.id, "name", player.Name, "seat", player.position)

	r.publishPublic(protocol.EventPlayerJoined, protocol.PlayerJoinedData{Player: protocol.PlayerInfo{
		PlayerID: player.id,
		Name:     player.Name,
		Seat:     player.position,
	}})

	return protocol.JoinResponse{
		Accepted: true,
		TableID:  r.tableID,
		PlayerID: player.id,
		Seat:     player.position,
	}
}

func (r *Table) handleAddBot(state *tableState, strategyName string) (AddedBot, error) {
	if state.gameOver {
		return AddedBot{}, fmt.Errorf("game is over")
	}
	if state.round != nil {
		return AddedBot{}, fmt.Errorf("round already in progress")
	}
	if len(state.players) >= game.PlayersPerTable {
		return AddedBot{}, fmt.Errorf("table is full")
	}

	strategyKind, err := bot.ParseStrategyKind(strategyName)
	if err != nil {
		return AddedBot{}, err
	}

	taken := make(map[string]bool, len(state.players))
	for _, p := range state.players {
		taken[p.Name] = true
	}

	id := r.nextPlayerID(state)
	player := r.addPlayer(state, id, strategyKind.BotName(taken), strategyKind.NewBot(), "")

	slog.Debug("bot added to table", "event", "bot_added", "table_id", r.tableID, "player_id", player.id, "name", player.Name, "strategy", string(strategyKind))

	r.publishPublic(protocol.EventPlayerJoined, protocol.PlayerJoinedData{Player: protocol.PlayerInfo{
		PlayerID: player.id,
		Name:     player.Name,
		Seat:     player.position,
	}})

	return AddedBot{
		JoinResponse: protocol.JoinResponse{
			Accepted: true,
			TableID:  r.tableID,
			PlayerID: player.id,
			Seat:     player.position,
		},
		Name:     player.Name,
		Strategy: string(strategyKind),
	}, nil
}

func (r *Table) addPlayer(state *tableState, id protocol.PlayerID, name string, b bot.Bot, token string) *playerState {
	player := &playerState{
		id:       id,
		Name:     name,
		position: len(state.players),
		Token:    token,
		bot:      b,
	}

	state.players = append(state.players, player)
	state.playersByID[id] = player
	if token != "" {
		state.playersByToken[token] = player
	}

	return player
}

func (r *Table) handleStart(state *tableState, playerID protocol.PlayerID) protocol.CommandResponse {
	if reason := r.validateStartPreconditions(state, playerID); reason != "" {
		return protocol.CommandResponse{Accepted: false, Reason: reason}
	}

	state.round = r.initializeRound(state)
	slog.Info("table started", "event", "table_started", "table_id", r.tableID, "round", state.roundsStarted)

	hands := make(map[protocol.PlayerID][]string, len(state.players))
	for _, player := range state.players {
		hands[player.id] = game.CardStrings(state.round.Hand(player.position))
	}
	r.publishRoundStart(state, hands)

	if state.round.PassDirection() == game.PassDirectionHold {
		_ = state.round.StartPlaying()
		r.publishPlayPhaseStart(state)
		return protocol.CommandResponse{Accepted: true}
	}

	r.schedulePassingBots(state)

	return protocol.CommandResponse{Accepted: true}
}

func (r *Table) nextPlayerID(state *tableState) protocol.PlayerID {
	for {
		state.nextPlayerSeq++
		candidate := protocol.PlayerID(fmt.Sprintf("p-%d", state.nextPlayerSeq))
		if _, exists := state.playersByID[candidate]; !exists {
			return candidate
		}
	}
}

func (r *Table) handlePlay(state *tableState, playerID protocol.PlayerID, cardRaw string) protocol.CommandResponse {
	card, err := game.ParseCard(cardRaw)
	if err != nil {
		return protocol.CommandResponse{Accepted: false, Reason: err.Error()}
	}

	if state.round == nil {
		return protocol.CommandResponse{Accepted: false, Reason: "round is not running"}
	}
	if state.round.Phase() != game.PhasePlaying {
		return protocol.CommandResponse{Accepted: false, Reason: "round is not in play phase"}
	}

	player := state.playersByID[playerID]
	if player == nil {
		return protocol.CommandResponse{Accepted: false, Reason: "player is not seated"}
	}
	if player.position != state.round.TurnSeat() {
		return protocol.CommandResponse{Accepted: false, Reason: "not your turn"}
	}

	heartsBrokenBefore := state.round.HeartsBroken()
	trickResult, err := state.round.Play(player.position, card)
	if err != nil {
		return protocol.CommandResponse{Accepted: false, Reason: err.Error()}
	}

	breaksHearts := card.Suit == game.SuitHearts && !heartsBrokenBefore
	r.publishPublic(protocol.EventCardPlayed, protocol.CardPlayedData{
		PlayerID: playerID, Card: card.String(), BreaksHearts: breaksHearts,
	})
	r.publishPrivate(playerID, protocol.EventHandUpdated, protocol.HandUpdatedData{
		Cards: game.CardStrings(state.round.Hand(player.position)),
	})

	if trickResult == nil {
		// Trick not complete — advance to next player.
		nextSeat := state.round.TurnSeat()
		nextPlayer := state.players[nextSeat]
		r.publishTurn(nextPlayer.id, state.round.TrickNumber())
		r.scheduleBotTurn(state, nextPlayer)
		return protocol.CommandResponse{Accepted: true}
	}

	// Trick completed.
	winnerPlayer := state.players[trickResult.WinnerSeat]
	trickCompleted := protocol.TrickCompletedData{
		TrickNumber:    trickResult.TrickNumber,
		WinnerPlayerID: winnerPlayer.id,
		Points:         trickResult.Points,
	}

	if state.round.Phase() == game.PhaseComplete {
		// Last trick — round is complete.
		roundCompleted := r.completeRound(state)
		r.publishPublic(protocol.EventTrickCompleted, trickCompleted)
		r.publishPublic(protocol.EventRoundCompleted, roundCompleted)
		state.round = nil
		r.maybeEndGame(state)
		return protocol.CommandResponse{Accepted: true}
	}

	// More tricks to play.
	r.publishPublic(protocol.EventTrickCompleted, trickCompleted)
	r.publishTurn(winnerPlayer.id, state.round.TrickNumber())
	r.scheduleBotTurn(state, winnerPlayer)

	return protocol.CommandResponse{Accepted: true}
}

func (r *Table) handlePass(state *tableState, playerID protocol.PlayerID, cardsRaw []string) protocol.CommandResponse {
	if state.round == nil {
		return protocol.CommandResponse{Accepted: false, Reason: "round is not running"}
	}
	if state.round.Phase() != game.PhasePassing {
		return protocol.CommandResponse{Accepted: false, Reason: "round is not in pass phase"}
	}

	player := state.playersByID[playerID]
	if player == nil {
		return protocol.CommandResponse{Accepted: false, Reason: "player is not seated"}
	}
	if state.round.HasSubmittedPass(player.position) {
		return protocol.CommandResponse{Accepted: false, Reason: "pass already submitted"}
	}

	passCards, err := r.parseAndValidatePassCards(state.round.Hand(player.position), cardsRaw)
	if err != nil {
		return protocol.CommandResponse{Accepted: false, Reason: err.Error()}
	}

	if err := state.round.SubmitPass(player.position, passCards); err != nil {
		return protocol.CommandResponse{Accepted: false, Reason: err.Error()}
	}

	submitted := 0
	for i := 0; i < game.PlayersPerTable; i++ {
		if state.round.HasSubmittedPass(i) {
			submitted++
		}
	}

	r.publishPublic(protocol.EventPassSubmitted, protocol.PassStatusData{
		Submitted: submitted,
		Total:     game.PlayersPerTable,
		Direction: state.round.PassDirection(),
	})

	if submitted < game.PlayersPerTable {
		return protocol.CommandResponse{Accepted: true}
	}

	_ = state.round.ApplyPasses()
	r.startPassReview(state)

	return protocol.CommandResponse{Accepted: true}
}

func (r *Table) handleReadyAfterPass(state *tableState, playerID protocol.PlayerID) protocol.CommandResponse {
	if state.round == nil {
		return protocol.CommandResponse{Accepted: false, Reason: "round is not running"}
	}
	if state.round.Phase() != game.PhasePassReview {
		return protocol.CommandResponse{Accepted: false, Reason: "round is not in pass review"}
	}
	if state.playersByID[playerID] == nil {
		return protocol.CommandResponse{Accepted: false, Reason: "player is not seated"}
	}

	_ = state.round.MarkReady(state.playersByID[playerID].position)

	ready := 0
	for i := 0; i < game.PlayersPerTable; i++ {
		if state.round.IsPassReady(i) {
			ready++
		}
	}

	r.publishPublic(protocol.EventPassReadyChanged, protocol.PassReadyData{
		Ready: ready,
		Total: game.PlayersPerTable,
	})

	if ready < game.PlayersPerTable {
		return protocol.CommandResponse{Accepted: true}
	}

	_ = state.round.StartPlaying()
	r.publishPlayPhaseStart(state)
	return protocol.CommandResponse{Accepted: true}
}

func (r *Table) parseAndValidatePassCards(hand []game.Card, cardsRaw []string) ([]game.Card, error) {
	if len(cardsRaw) != 3 {
		return nil, fmt.Errorf("pass requires exactly 3 cards")
	}

	cards := make([]game.Card, 0, len(cardsRaw))
	seen := make(map[game.Card]struct{}, len(cardsRaw))
	for _, raw := range cardsRaw {
		card, err := game.ParseCard(raw)
		if err != nil {
			return nil, err
		}
		if _, duplicate := seen[card]; duplicate {
			return nil, fmt.Errorf("pass contains duplicate card %s", card.String())
		}
		if !game.ContainsCard(hand, card) {
			return nil, fmt.Errorf("card %s is not in hand", card.String())
		}

		seen[card] = struct{}{}
		cards = append(cards, card)
	}

	return cards, nil
}

func (r *Table) startPassReview(state *tableState) {
	// Auto-mark bots ready.
	for _, player := range state.players {
		if player.bot != nil {
			_ = state.round.MarkReady(player.position)
		}
	}

	ready := 0
	for i := 0; i < game.PlayersPerTable; i++ {
		if state.round.IsPassReady(i) {
			ready++
		}
	}

	for _, player := range state.players {
		r.publishPrivate(player.id, protocol.EventHandUpdated, protocol.HandUpdatedData{Cards: game.CardStrings(state.round.Hand(player.position))})
	}

	r.publishPublic(protocol.EventPassReviewStarted, protocol.PassStatusData{
		Submitted: game.PlayersPerTable,
		Total:     game.PlayersPerTable,
		Direction: state.round.PassDirection(),
	})
	r.publishPublic(protocol.EventPassReadyChanged, protocol.PassReadyData{Ready: ready, Total: game.PlayersPerTable})

	if ready == game.PlayersPerTable {
		_ = state.round.StartPlaying()
		r.publishPlayPhaseStart(state)
	}
}

func (r *Table) publishPlayPhaseStart(state *tableState) {
	if state.round == nil {
		return
	}
	turnSeat := state.round.TurnSeat()
	turnPlayer := state.players[turnSeat]
	r.publishTurn(turnPlayer.id, 0)
	r.scheduleBotTurn(state, turnPlayer)
}

func (r *Table) handleBotTurn(state *tableState, playerID protocol.PlayerID) {
	if state.round == nil || state.round.Phase() != game.PhasePlaying {
		return
	}

	player := state.playersByID[playerID]
	if player == nil || player.bot == nil || player.position != state.round.TurnSeat() {
		return
	}

	input := state.round.TurnInput(player.position)
	card, err := player.bot.ChoosePlay(input)
	if err != nil {
		return
	}
	_ = r.handlePlay(state, playerID, card.String())
}

func (r *Table) handleBotPass(state *tableState, playerID protocol.PlayerID) {
	if state.round == nil || state.round.Phase() != game.PhasePassing {
		return
	}

	player := state.playersByID[playerID]
	if player == nil || player.bot == nil || state.round.HasSubmittedPass(player.position) {
		return
	}

	input := state.round.PassInput(player.position)
	cards, err := player.bot.ChoosePass(input)
	if err != nil {
		return
	}

	_ = r.handlePass(state, playerID, game.CardStrings(cards))
}

func (r *Table) scheduleBotTurn(_ *tableState, player *playerState) {
	if player.bot == nil {
		return
	}

	go r.submit(botTurnCommand{playerID: player.id})
}

func (r *Table) schedulePassingBots(state *tableState) {
	if state.round == nil || state.round.Phase() != game.PhasePassing {
		return
	}

	for _, player := range state.players {
		if player.bot != nil {
			r.scheduleBotPass(player)
		}
	}
}

func (r *Table) scheduleBotPass(player *playerState) {
	if player.bot == nil {
		return
	}

	go r.submit(botPassCommand{playerID: player.id})
}

func (r *Table) validateStartPreconditions(state *tableState, playerID protocol.PlayerID) string {
	if state.gameOver {
		return "game is over"
	}
	if state.round != nil {
		return "round already started"
	}
	if len(state.players) == 0 {
		return "at least one player must join before start"
	}
	if state.playersByID[playerID] == nil {
		return "only seated players can start"
	}
	if len(state.players) != game.PlayersPerTable {
		return fmt.Sprintf("table requires %d players before start", game.PlayersPerTable)
	}
	return ""
}

func (r *Table) initializeRound(state *tableState) *game.Round {
	passDirection := game.PassDirectionForRound(state.roundsStarted)
	state.roundsStarted++

	var hands [game.PlayersPerTable][]game.Card
	deck := defaultShuffledDeck()
	for i, card := range deck {
		seat := i % game.PlayersPerTable
		hands[seat] = append(hands[seat], card)
	}
	for i := range hands {
		game.SortCards(hands[i])
	}

	return game.NewRound(hands, passDirection)
}

func defaultShuffledDeck() []game.Card {
	deck := game.BuildDeck()
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	game.Shuffle(deck, rng)
	return deck
}

func (r *Table) publishRoundStart(state *tableState, hands map[protocol.PlayerID][]string) {
	r.publishPublic(protocol.EventGameStarted, struct{}{})
	for playerID, cards := range hands {
		r.publishPrivate(playerID, protocol.EventHandUpdated, protocol.HandUpdatedData{Cards: cards})
	}

	if state.round != nil && state.round.PassDirection() != game.PassDirectionHold {
		r.publishPublic(protocol.EventPassSubmitted, protocol.PassStatusData{
			Submitted: 0,
			Total:     game.PlayersPerTable,
			Direction: state.round.PassDirection(),
		})
	}
}

func (r *Table) publishTurn(playerID protocol.PlayerID, trickNumber int) {
	r.publishPublic(protocol.EventTurnChanged, protocol.TurnChangedData{PlayerID: playerID, TrickNumber: trickNumber})
	r.publishPrivate(playerID, protocol.EventYourTurn, protocol.YourTurnData{PlayerID: playerID, TrickNumber: trickNumber})
}

func (r *Table) completeRound(state *tableState) protocol.RoundCompletedData {
	scores := state.round.Scores()
	adjustedRound := make(map[protocol.PlayerID]game.Points, len(state.players))
	for i, player := range state.players {
		player.cumulativePoints += scores.Adjusted[i]
		adjustedRound[player.id] = scores.Adjusted[i]
	}
	state.roundHistory = append(state.roundHistory, adjustedRound)

	totals := make(map[protocol.PlayerID]game.Points, len(state.players))
	for _, player := range state.players {
		totals[player.id] = player.cumulativePoints
	}
	return protocol.RoundCompletedData{
		RoundPoints: copyPoints(adjustedRound),
		TotalPoints: totals,
	}
}

const gameOverThreshold = game.Points(100)

func computeWinners(totals map[protocol.PlayerID]game.Points) []protocol.PlayerID {
	var minPts game.Points
	first := true
	for _, pts := range totals {
		if first || pts < minPts {
			minPts = pts
			first = false
		}
	}

	winners := make([]protocol.PlayerID, 0, len(totals))
	for playerID, pts := range totals {
		if pts == minPts {
			winners = append(winners, playerID)
		}
	}
	sort.Slice(winners, func(i, j int) bool { return winners[i] < winners[j] })
	return winners
}

func (r *Table) maybeEndGame(state *tableState) {
	for _, player := range state.players {
		if player.cumulativePoints >= gameOverThreshold {
			state.gameOver = true
			totals := make(map[protocol.PlayerID]game.Points, len(state.players))
			for _, p := range state.players {
				totals[p.id] = p.cumulativePoints
			}
			r.publishPublic(protocol.EventGameOver, protocol.GameOverData{
				FinalScores: copyPoints(totals),
				Winners:     computeWinners(totals),
			})
			return
		}
	}
}

func (r *Table) buildBotHands(state *tableState) []BotHandSnapshot {
	var out []BotHandSnapshot
	for _, p := range state.players {
		if p.bot == nil {
			continue
		}
		var hand []game.Card
		if state.round != nil {
			hand = state.round.Hand(p.position)
		}
		cards := make([]string, 0, len(hand))
		for _, c := range hand {
			cards = append(cards, c.String())
		}
		out = append(out, BotHandSnapshot{Name: p.Name, Seat: p.position, Cards: cards})
	}
	return out
}

func (r *Table) buildSnapshot(state *tableState, forPlayer protocol.PlayerID) Snapshot {
	playerSnapshots := make([]PlayerSnapshot, 0, len(state.players))
	for _, player := range state.players {
		playerSnapshots = append(playerSnapshots, PlayerSnapshot{
			PlayerID: player.id,
			Name:     player.Name,
			Seat:     player.position,
			IsBot:    player.bot != nil,
		})
	}
	sort.Slice(playerSnapshots, func(i, j int) bool {
		return playerSnapshots[i].Seat < playerSnapshots[j].Seat
	})

	totals := make(map[protocol.PlayerID]game.Points, len(state.players))
	for _, player := range state.players {
		totals[player.id] = player.cumulativePoints
	}

	snapshot := Snapshot{
		TableID:      r.tableID,
		Players:      playerSnapshots,
		Started:      state.round != nil,
		Phase:        "",
		HandSizes:    map[protocol.PlayerID]int{},
		RoundPoints:  map[protocol.PlayerID]game.Points{},
		RoundHistory: copyRoundHistory(state.roundHistory),
		TotalPoints:  totals,
		GameOver:     state.gameOver,
	}

	if state.gameOver {
		snapshot.Winners = computeWinners(totals)
	}

	if state.round != nil {
		for _, player := range state.players {
			snapshot.HandSizes[player.id] = len(state.round.Hand(player.position))
		}

		snapshot.Phase = roundPhaseString(state.round.Phase())
		snapshot.TrickNumber = state.round.TrickNumber()
		if state.round.Phase() == game.PhasePlaying {
			snapshot.TurnPlayerID = state.players[state.round.TurnSeat()].id
		}
		snapshot.HeartsBroken = state.round.HeartsBroken()
		snapshot.PassDirection = state.round.PassDirection()

		submitted := 0
		readyCount := 0
		for i := 0; i < game.PlayersPerTable; i++ {
			if state.round.HasSubmittedPass(i) {
				submitted++
			}
			if state.round.IsPassReady(i) {
				readyCount++
			}
		}
		snapshot.PassSubmittedCount = submitted
		snapshot.PassReadyCount = readyCount

		trick := state.round.CurrentTrick()
		snapshot.CurrentTrick = make([]string, 0, len(trick))
		snapshot.TrickPlays = make([]TrickPlaySnapshot, 0, len(trick))
		for _, played := range trick {
			snapshot.CurrentTrick = append(snapshot.CurrentTrick, played.Card.String())
			p := state.players[played.Seat]
			snapshot.TrickPlays = append(snapshot.TrickPlays, TrickPlaySnapshot{
				PlayerID: p.id,
				Name:     p.Name,
				Seat:     played.Seat,
				Card:     played.Card.String(),
			})
		}

		roundPoints := make(map[protocol.PlayerID]game.Points, len(state.players))
		for i, p := range state.players {
			roundPoints[p.id] = state.round.RoundPoints(i)
		}
		snapshot.RoundPoints = roundPoints
	}

	if forPlayer != "" {
		if player := state.playersByID[forPlayer]; player != nil && state.round != nil {
			snapshot.Hand = game.CardStrings(state.round.Hand(player.position))
			snapshot.PassSubmitted = state.round.HasSubmittedPass(player.position)
			snapshot.PassReady = state.round.IsPassReady(player.position)
			snapshot.PassSent = game.CardStrings(state.round.PassSent(player.position))
			snapshot.PassReceived = game.CardStrings(state.round.PassReceived(player.position))
		}
	}

	return snapshot
}

func roundPhaseString(phase game.RoundPhase) string {
	switch phase {
	case game.PhasePassing:
		return "passing"
	case game.PhasePassReview:
		return "pass_review"
	case game.PhasePlaying:
		return "playing"
	case game.PhaseComplete:
		return "complete"
	default:
		return ""
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

func copyPoints(source map[protocol.PlayerID]game.Points) map[protocol.PlayerID]game.Points {
	out := make(map[protocol.PlayerID]game.Points, len(source))
	for key, value := range source {
		out[key] = value
	}
	return out
}

func copyRoundHistory(source []map[protocol.PlayerID]game.Points) []map[protocol.PlayerID]game.Points {
	out := make([]map[protocol.PlayerID]game.Points, 0, len(source))
	for _, entry := range source {
		out = append(out, copyPoints(entry))
	}

	return out
}
