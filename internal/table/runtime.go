package table

import (
	"fmt"
	"log/slog"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/JHK/hearts/internal/game"
	"github.com/JHK/hearts/internal/player/bot"
	"github.com/JHK/hearts/internal/protocol"
)

type StreamEvent struct {
	Type      protocol.EventType
	Data      any
	PrivateTo game.PlayerID
}

type AddedBot struct {
	JoinResponse protocol.JoinResponse `json:"join_response"`
	Name         string                `json:"name"`
	Strategy     string                `json:"strategy"`
}

type PlayerSnapshot struct {
	PlayerID game.PlayerID `json:"player_id"`
	Name     string        `json:"name"`
	Seat     int           `json:"seat"`
	IsBot    bool          `json:"is_bot"`
}

type TrickPlaySnapshot struct {
	PlayerID game.PlayerID `json:"player_id"`
	Name     string        `json:"name"`
	Seat     int           `json:"seat"`
	Card     string        `json:"card"`
}

type Snapshot struct {
	TableID            string                          `json:"table_id"`
	Players            []PlayerSnapshot                `json:"players"`
	Started            bool                            `json:"started"`
	Phase              string                          `json:"phase"`
	TrickNumber        int                             `json:"trick_number"`
	TurnPlayerID       game.PlayerID                   `json:"turn_player_id"`
	HeartsBroken       bool                            `json:"hearts_broken"`
	CurrentTrick       []string                        `json:"current_trick"`
	TrickPlays         []TrickPlaySnapshot             `json:"trick_plays"`
	Hand               []string                        `json:"hand"`
	HandSizes          map[game.PlayerID]int           `json:"hand_sizes"`
	PassDirection      string                          `json:"pass_direction"`
	PassSubmitted      bool                            `json:"pass_submitted"`
	PassSubmittedCount int                             `json:"pass_submitted_count"`
	PassSent           []string                        `json:"pass_sent"`
	PassReceived       []string                        `json:"pass_received"`
	PassReady          bool                            `json:"pass_ready"`
	PassReadyCount     int                             `json:"pass_ready_count"`
	RoundPoints        map[game.PlayerID]game.Points   `json:"round_points"`
	RoundHistory       []map[game.PlayerID]game.Points `json:"round_history"`
	TotalPoints        map[game.PlayerID]game.Points   `json:"total_points"`
	GameOver           bool                            `json:"game_over"`
	Winners            []game.PlayerID                 `json:"winners,omitempty"`
}

type roundPhase string

const (
	roundPhasePassing    roundPhase = "passing"
	roundPhasePassReview roundPhase = "pass_review"
	roundPhasePlaying    roundPhase = "playing"

	passDirectionLeft   = "left"
	passDirectionRight  = "right"
	passDirectionAcross = "across"
	passDirectionHold   = "hold"
)

type Runtime struct {
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
	playersByID    map[game.PlayerID]*playerState
	playersByToken map[string]*playerState
	bots           map[game.PlayerID]bot.Strategy
	totals         map[game.PlayerID]game.Points
	roundHistory   []map[game.PlayerID]game.Points

	round         *roundState
	roundsStarted int
	nextPlayerSeq int
	gameOver      bool
}

type playerState struct {
	PlayerID game.PlayerID
	Name     string
	Seat     int
	Hand     []game.Card
	Token    string
	IsBot    bool
}

type roundState struct {
	Phase           roundPhase
	PassDirection   string
	PassSubmissions map[game.PlayerID][]game.Card
	PassSent        map[game.PlayerID][]game.Card
	PassReceived    map[game.PlayerID][]game.Card
	PassReady       map[game.PlayerID]bool

	TrickNumber  int
	TurnSeat     int
	HeartsBroken bool
	Trick        []trickPlay
	RoundPoints  map[game.PlayerID]game.Points
}

type trickPlay struct {
	Seat     int
	PlayerID game.PlayerID
	Card     game.Card
}

type playUpdate struct {
	playerID    game.PlayerID
	cardPlayed  protocol.CardPlayedData
	handUpdated protocol.HandUpdatedData
}

type joinCommand struct {
	name  string
	token string
	reply chan protocol.JoinResponse
}

type startCommand struct {
	playerID game.PlayerID
	reply    chan protocol.CommandResponse
}

type playCommand struct {
	playerID game.PlayerID
	card     string
	reply    chan protocol.CommandResponse
}

type passCommand struct {
	playerID game.PlayerID
	cards    []string
	reply    chan protocol.CommandResponse
}

type readyAfterPassCommand struct {
	playerID game.PlayerID
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
	forPlayer game.PlayerID
	reply     chan Snapshot
}

type leaveCommand struct {
	playerID game.PlayerID
	reply    chan struct{}
}

type infoCommand struct {
	reply chan protocol.TableInfo
}

type botTurnCommand struct {
	playerID game.PlayerID
}

type botPassCommand struct {
	playerID game.PlayerID
}

func NewRuntime(tableID string) *Runtime {
	r := &Runtime{
		tableID:   tableID,
		commands:  make(chan any),
		stop:      make(chan struct{}),
		stoppedCh: make(chan struct{}),
		subs:      make(map[int]chan StreamEvent),
	}

	go r.run()
	return r
}

func (r *Runtime) ID() string {
	return r.tableID
}

func (r *Runtime) Close() {
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

func (r *Runtime) Subscribe() (<-chan StreamEvent, func()) {
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

func (r *Runtime) Join(name, token string) (protocol.JoinResponse, error) {
	reply := make(chan protocol.JoinResponse, 1)
	if !r.submit(joinCommand{name: name, token: token, reply: reply}) {
		return protocol.JoinResponse{}, fmt.Errorf("table is stopping")
	}
	return <-reply, nil
}

func (r *Runtime) Start(playerID game.PlayerID) protocol.CommandResponse {
	reply := make(chan protocol.CommandResponse, 1)
	if !r.submit(startCommand{playerID: playerID, reply: reply}) {
		return protocol.CommandResponse{Accepted: false, Reason: "table is stopping"}
	}
	return <-reply
}

func (r *Runtime) Play(playerID game.PlayerID, card string) protocol.CommandResponse {
	reply := make(chan protocol.CommandResponse, 1)
	if !r.submit(playCommand{playerID: playerID, card: card, reply: reply}) {
		return protocol.CommandResponse{Accepted: false, Reason: "table is stopping"}
	}
	return <-reply
}

func (r *Runtime) Pass(playerID game.PlayerID, cards []string) protocol.CommandResponse {
	reply := make(chan protocol.CommandResponse, 1)
	if !r.submit(passCommand{playerID: playerID, cards: cards, reply: reply}) {
		return protocol.CommandResponse{Accepted: false, Reason: "table is stopping"}
	}
	return <-reply
}

func (r *Runtime) ReadyAfterPass(playerID game.PlayerID) protocol.CommandResponse {
	reply := make(chan protocol.CommandResponse, 1)
	if !r.submit(readyAfterPassCommand{playerID: playerID, reply: reply}) {
		return protocol.CommandResponse{Accepted: false, Reason: "table is stopping"}
	}
	return <-reply
}

func (r *Runtime) AddBot(strategyName string) (AddedBot, error) {
	reply := make(chan addBotResult, 1)
	if !r.submit(addBotCommand{strategyName: strategyName, reply: reply}) {
		return AddedBot{}, fmt.Errorf("table is stopping")
	}
	result := <-reply
	return result.added, result.err
}

func (r *Runtime) Snapshot(forPlayer game.PlayerID) Snapshot {
	reply := make(chan Snapshot, 1)
	if !r.submit(snapshotCommand{forPlayer: forPlayer, reply: reply}) {
		return Snapshot{TableID: r.tableID}
	}
	return <-reply
}

func (r *Runtime) Leave(playerID game.PlayerID) {
	if playerID == "" {
		return
	}

	reply := make(chan struct{}, 1)
	if !r.submit(leaveCommand{playerID: playerID, reply: reply}) {
		return
	}
	<-reply
}

func (r *Runtime) Info() protocol.TableInfo {
	reply := make(chan protocol.TableInfo, 1)
	if !r.submit(infoCommand{reply: reply}) {
		return protocol.TableInfo{TableID: r.tableID, MaxPlayers: game.PlayersPerTable}
	}
	return <-reply
}

func (r *Runtime) submit(command any) bool {
	select {
	case <-r.stop:
		return false
	case r.commands <- command:
		return true
	}
}

func (r *Runtime) run() {
	defer close(r.stoppedCh)

	state := &tableState{
		playersByID:    make(map[game.PlayerID]*playerState),
		playersByToken: make(map[string]*playerState),
		bots:           make(map[game.PlayerID]bot.Strategy),
		totals:         make(map[game.PlayerID]game.Points),
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
			}
		}
	}
}

func (r *Runtime) handleLeave(state *tableState, playerID game.PlayerID) {
	player := state.playersByID[playerID]
	if player == nil {
		return
	}

	slog.Info("player left table", "event", "player_left", "table_id", r.tableID, "player_id", playerID, "name", player.Name)

	if state.round != nil {
		// Preserve the token so the player can reclaim their seat if they reconnect.
		if !player.IsBot {
			player.IsBot = true
		}
		if state.bots[playerID] == nil {
			state.bots[playerID] = bot.StrategyRandom.New()
		}

		switch state.round.Phase {
		case roundPhasePassing:
			if _, submitted := state.round.PassSubmissions[playerID]; !submitted {
				if cards, err := r.chooseBotPassCardStrings(state, playerID, player.Hand); err == nil {
					_ = r.handlePass(state, playerID, cards)
				}
			}
		case roundPhasePassReview:
			if !state.round.PassReady[playerID] {
				_ = r.handleReadyAfterPass(state, playerID)
			}
		case roundPhasePlaying:
			if player.Seat == state.round.TurnSeat {
				r.scheduleBotTurn(state, playerID)
			}
		}
		return
	}

	delete(state.playersByID, playerID)
	if player.Token != "" {
		delete(state.playersByToken, player.Token)
	}
	delete(state.bots, playerID)
	delete(state.totals, playerID)

	for index, seated := range state.players {
		if seated.PlayerID != playerID {
			continue
		}

		state.players = append(state.players[:index], state.players[index+1:]...)
		for seat, updated := range state.players {
			updated.Seat = seat
		}
		return
	}
}

func (r *Runtime) handleJoin(state *tableState, name, token string) protocol.JoinResponse {
	token = strings.TrimSpace(token)
	if token == "" {
		return protocol.JoinResponse{Accepted: false, Reason: "player token is required"}
	}

	if existing := state.playersByToken[token]; existing != nil {
		if existing.IsBot {
			// Player reconnected after being converted to a bot mid-round — reclaim the seat.
			existing.IsBot = false
			if n := strings.TrimSpace(name); n != "" {
				existing.Name = n
			}
			delete(state.bots, existing.PlayerID)
			slog.Info("player reclaimed seat", "event", "player_reclaimed", "table_id", r.tableID, "player_id", existing.PlayerID, "name", existing.Name, "seat", existing.Seat)
		}
		return protocol.JoinResponse{
			Accepted: true,
			TableID:  r.tableID,
			PlayerID: existing.PlayerID,
			Seat:     existing.Seat,
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

	player := r.addPlayer(state, name, false, token)

	slog.Info("player joined table", "event", "player_joined", "table_id", r.tableID, "player_id", player.PlayerID, "name", player.Name, "seat", player.Seat)

	r.publishPublic(protocol.EventPlayerJoined, protocol.PlayerJoinedData{Player: protocol.PlayerInfo{
		PlayerID: player.PlayerID,
		Name:     player.Name,
		Seat:     player.Seat,
	}})

	return protocol.JoinResponse{
		Accepted: true,
		TableID:  r.tableID,
		PlayerID: player.PlayerID,
		Seat:     player.Seat,
	}
}

func (r *Runtime) handleAddBot(state *tableState, strategyName string) (AddedBot, error) {
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

	name := randomBotName()
	player := r.addPlayer(state, name, true, "")
	state.bots[player.PlayerID] = strategyKind.New()

	slog.Debug("bot added to table", "event", "bot_added", "table_id", r.tableID, "player_id", player.PlayerID, "name", player.Name, "strategy", string(strategyKind))

	r.publishPublic(protocol.EventPlayerJoined, protocol.PlayerJoinedData{Player: protocol.PlayerInfo{
		PlayerID: player.PlayerID,
		Name:     player.Name,
		Seat:     player.Seat,
	}})

	return AddedBot{
		JoinResponse: protocol.JoinResponse{
			Accepted: true,
			TableID:  r.tableID,
			PlayerID: player.PlayerID,
			Seat:     player.Seat,
		},
		Name:     player.Name,
		Strategy: string(strategyKind),
	}, nil
}

func (r *Runtime) addPlayer(state *tableState, name string, isBot bool, token string) *playerState {
	player := &playerState{
		PlayerID: r.nextPlayerID(state),
		Name:     name,
		Seat:     len(state.players),
		Token:    token,
		IsBot:    isBot,
	}

	state.players = append(state.players, player)
	state.playersByID[player.PlayerID] = player
	if token != "" {
		state.playersByToken[token] = player
	}

	if _, ok := state.totals[player.PlayerID]; !ok {
		state.totals[player.PlayerID] = 0
	}

	return player
}

func (r *Runtime) handleStart(state *tableState, playerID game.PlayerID) protocol.CommandResponse {
	if reason := r.validateStartPreconditions(state, playerID); reason != "" {
		return protocol.CommandResponse{Accepted: false, Reason: reason}
	}

	hands := r.initializeRound(state)
	slog.Info("table started", "event", "table_started", "table_id", r.tableID, "round", state.roundsStarted)
	r.publishRoundStart(state, hands)
	if state.round.PassDirection == passDirectionHold {
		r.startPlayPhase(state)
		return protocol.CommandResponse{Accepted: true}
	}

	r.schedulePassingBots(state)

	return protocol.CommandResponse{Accepted: true}
}

func (r *Runtime) nextPlayerID(state *tableState) game.PlayerID {
	for {
		state.nextPlayerSeq++
		candidate := game.PlayerID(fmt.Sprintf("p-%d", state.nextPlayerSeq))
		if _, exists := state.playersByID[candidate]; !exists {
			return candidate
		}
	}
}

func (r *Runtime) handlePlay(state *tableState, playerID game.PlayerID, cardRaw string) protocol.CommandResponse {
	card, err := game.ParseCard(cardRaw)
	if err != nil {
		return protocol.CommandResponse{Accepted: false, Reason: err.Error()}
	}

	if state.round == nil {
		return protocol.CommandResponse{Accepted: false, Reason: "round is not running"}
	}
	if state.round.Phase != roundPhasePlaying {
		return protocol.CommandResponse{Accepted: false, Reason: "round is not in play phase"}
	}

	player := state.playersByID[playerID]
	if player == nil {
		return protocol.CommandResponse{Accepted: false, Reason: "player is not seated"}
	}
	if player.Seat != state.round.TurnSeat {
		return protocol.CommandResponse{Accepted: false, Reason: "not your turn"}
	}

	err = game.ValidatePlay(game.ValidatePlayInput{
		Hand:         player.Hand,
		Card:         card,
		Trick:        currentTrickCards(state.round),
		HeartsBroken: state.round.HeartsBroken,
		FirstTrick:   state.round.TrickNumber == 0,
	})
	if err != nil {
		return protocol.CommandResponse{Accepted: false, Reason: err.Error()}
	}

	update, err := r.applyValidatedPlay(state.round, player, card)
	if err != nil {
		return protocol.CommandResponse{Accepted: false, Reason: err.Error()}
	}

	if len(state.round.Trick) < game.PlayersPerTable {
		nextSeat := (state.round.TurnSeat + 1) % game.PlayersPerTable
		nextPlayerID := state.players[nextSeat].PlayerID
		trickNumber := state.round.TrickNumber
		state.round.TurnSeat = nextSeat

		r.publishPlayAndHand(update)
		r.publishTurn(nextPlayerID, trickNumber)
		r.scheduleBotTurn(state, nextPlayerID)
		return protocol.CommandResponse{Accepted: true}
	}

	winnerPlayerID, points, err := game.TrickWinner(currentTrickPlays(state.round))
	if err != nil {
		return protocol.CommandResponse{Accepted: false, Reason: err.Error()}
	}

	state.round.RoundPoints[winnerPlayerID] += points
	trickNumber := state.round.TrickNumber
	trickCompleted := protocol.TrickCompletedData{TrickNumber: trickNumber, WinnerPlayerID: winnerPlayerID, Points: points}

	if trickNumber == 12 {
		roundCompleted := r.completeRound(state)
		state.round = nil

		r.publishPlayAndHand(update)
		r.publishPublic(protocol.EventTrickCompleted, trickCompleted)
		r.publishPublic(protocol.EventRoundCompleted, roundCompleted)
		r.maybeEndGame(state)
		return protocol.CommandResponse{Accepted: true}
	}

	nextSeat := state.playersByID[winnerPlayerID].Seat
	nextTrick := trickNumber + 1

	state.round.Trick = nil
	state.round.TrickNumber = nextTrick
	state.round.TurnSeat = nextSeat

	r.publishPlayAndHand(update)
	r.publishPublic(protocol.EventTrickCompleted, trickCompleted)
	r.publishTurn(winnerPlayerID, nextTrick)
	r.scheduleBotTurn(state, winnerPlayerID)

	return protocol.CommandResponse{Accepted: true}
}

func (r *Runtime) handlePass(state *tableState, playerID game.PlayerID, cardsRaw []string) protocol.CommandResponse {
	if state.round == nil {
		return protocol.CommandResponse{Accepted: false, Reason: "round is not running"}
	}
	if state.round.Phase != roundPhasePassing {
		return protocol.CommandResponse{Accepted: false, Reason: "round is not in pass phase"}
	}

	player := state.playersByID[playerID]
	if player == nil {
		return protocol.CommandResponse{Accepted: false, Reason: "player is not seated"}
	}
	if _, submitted := state.round.PassSubmissions[playerID]; submitted {
		return protocol.CommandResponse{Accepted: false, Reason: "pass already submitted"}
	}

	passCards, err := r.parseAndValidatePassCards(player.Hand, cardsRaw)
	if err != nil {
		return protocol.CommandResponse{Accepted: false, Reason: err.Error()}
	}

	state.round.PassSubmissions[playerID] = passCards
	state.round.PassSent[playerID] = append([]game.Card(nil), passCards...)

	r.publishPublic(protocol.EventPassSubmitted, protocol.PassStatusData{
		Submitted: len(state.round.PassSubmissions),
		Total:     game.PlayersPerTable,
		Direction: state.round.PassDirection,
	})

	if len(state.round.PassSubmissions) < game.PlayersPerTable {
		return protocol.CommandResponse{Accepted: true}
	}

	r.applyPassSubmissions(state)
	r.startPassReview(state)

	return protocol.CommandResponse{Accepted: true}
}

func (r *Runtime) handleReadyAfterPass(state *tableState, playerID game.PlayerID) protocol.CommandResponse {
	if state.round == nil {
		return protocol.CommandResponse{Accepted: false, Reason: "round is not running"}
	}
	if state.round.Phase != roundPhasePassReview {
		return protocol.CommandResponse{Accepted: false, Reason: "round is not in pass review"}
	}
	if state.playersByID[playerID] == nil {
		return protocol.CommandResponse{Accepted: false, Reason: "player is not seated"}
	}

	state.round.PassReady[playerID] = true

	ready := 0
	for _, isReady := range state.round.PassReady {
		if isReady {
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

	r.startPlayPhase(state)
	return protocol.CommandResponse{Accepted: true}
}

func (r *Runtime) parseAndValidatePassCards(hand []game.Card, cardsRaw []string) ([]game.Card, error) {
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

func (r *Runtime) applyPassSubmissions(state *tableState) {
	offset := r.passDirectionOffset(state.round.PassDirection)

	for _, sender := range state.players {
		cards := state.round.PassSubmissions[sender.PlayerID]
		receiverSeat := (sender.Seat + offset) % game.PlayersPerTable
		receiver := state.players[receiverSeat]

		for _, card := range cards {
			updatedHand, removed := game.RemoveCard(sender.Hand, card)
			if !removed {
				continue
			}
			sender.Hand = updatedHand
			receiver.Hand = append(receiver.Hand, card)
			state.round.PassReceived[receiver.PlayerID] = append(state.round.PassReceived[receiver.PlayerID], card)
		}
	}

	for _, player := range state.players {
		game.SortCards(player.Hand)
	}
}

func (r *Runtime) startPassReview(state *tableState) {
	state.round.Phase = roundPhasePassReview
	for _, player := range state.players {
		state.round.PassReady[player.PlayerID] = player.IsBot
		r.publishPrivate(player.PlayerID, protocol.EventHandUpdated, protocol.HandUpdatedData{Cards: game.CardStrings(player.Hand)})
	}

	ready := 0
	for _, isReady := range state.round.PassReady {
		if isReady {
			ready++
		}
	}

	r.publishPublic(protocol.EventPassReviewStarted, protocol.PassStatusData{
		Submitted: game.PlayersPerTable,
		Total:     game.PlayersPerTable,
		Direction: state.round.PassDirection,
	})
	r.publishPublic(protocol.EventPassReadyChanged, protocol.PassReadyData{Ready: ready, Total: game.PlayersPerTable})

	if ready == game.PlayersPerTable {
		r.startPlayPhase(state)
	}
}

func (r *Runtime) startPlayPhase(state *tableState) {
	if state.round == nil {
		return
	}

	state.round.Phase = roundPhasePlaying
	state.round.Trick = nil
	state.round.TrickNumber = 0
	state.round.HeartsBroken = false

	startSeat := r.findTwoClubsSeat(state)
	state.round.TurnSeat = startSeat
	turnPlayerID := state.players[startSeat].PlayerID

	r.publishTurn(turnPlayerID, 0)
	r.scheduleBotTurn(state, turnPlayerID)
}

func (r *Runtime) findTwoClubsSeat(state *tableState) int {
	twoClubs := game.Card{Suit: game.SuitClubs, Rank: 2}
	for _, player := range state.players {
		if game.ContainsCard(player.Hand, twoClubs) {
			return player.Seat
		}
	}

	return 0
}

func (r *Runtime) passDirectionOffset(direction string) int {
	switch direction {
	case passDirectionLeft:
		return 1
	case passDirectionRight:
		return game.PlayersPerTable - 1
	case passDirectionAcross:
		return 2
	case passDirectionHold:
		return 0
	default:
		return 1
	}
}

func passDirectionForRound(roundIndex int) string {
	switch roundIndex % 4 {
	case 0:
		return passDirectionLeft
	case 1:
		return passDirectionRight
	case 2:
		return passDirectionAcross
	default:
		return passDirectionHold
	}
}

func (r *Runtime) handleBotTurn(state *tableState, playerID game.PlayerID) {
	if state.round == nil || state.round.Phase != roundPhasePlaying {
		return
	}

	player := state.playersByID[playerID]
	if player == nil || player.Seat != state.round.TurnSeat {
		return
	}

	strategy := state.bots[playerID]
	if strategy == nil {
		return
	}

	card, err := strategy.ChoosePlay(bot.TurnInput{
		Hand:         player.Hand,
		Trick:        currentTrickCards(state.round),
		HeartsBroken: state.round.HeartsBroken,
		FirstTrick:   state.round.TrickNumber == 0,
	})
	if err != nil {
		return
	}

	_ = r.handlePlay(state, playerID, card.String())
}

func (r *Runtime) handleBotPass(state *tableState, playerID game.PlayerID) {
	if state.round == nil || state.round.Phase != roundPhasePassing {
		return
	}

	player := state.playersByID[playerID]
	if player == nil {
		return
	}
	if _, submitted := state.round.PassSubmissions[playerID]; submitted {
		return
	}

	cards, err := r.chooseBotPassCardStrings(state, playerID, player.Hand)
	if err != nil {
		return
	}

	_ = r.handlePass(state, playerID, cards)
}

func (r *Runtime) scheduleBotTurn(state *tableState, playerID game.PlayerID) {
	if state.bots[playerID] == nil {
		return
	}

	go func() {
		timer := time.NewTimer(350 * time.Millisecond)
		defer timer.Stop()

		select {
		case <-r.stop:
			return
		case <-timer.C:
		}

		r.submit(botTurnCommand{playerID: playerID})
	}()
}

func (r *Runtime) schedulePassingBots(state *tableState) {
	if state.round == nil || state.round.Phase != roundPhasePassing {
		return
	}

	for _, player := range state.players {
		if !player.IsBot {
			continue
		}
		r.scheduleBotPass(state, player.PlayerID)
	}
}

func (r *Runtime) scheduleBotPass(state *tableState, playerID game.PlayerID) {
	if state.bots[playerID] == nil {
		return
	}

	go func() {
		timer := time.NewTimer(300 * time.Millisecond)
		defer timer.Stop()

		select {
		case <-r.stop:
			return
		case <-timer.C:
		}

		r.submit(botPassCommand{playerID: playerID})
	}()
}

func (r *Runtime) chooseBotPassCardStrings(state *tableState, playerID game.PlayerID, hand []game.Card) ([]string, error) {
	strategy := state.bots[playerID]
	if strategy == nil {
		return nil, fmt.Errorf("bot strategy is not available")
	}

	cards, err := strategy.ChoosePass(bot.PassInput{
		Hand:      append([]game.Card(nil), hand...),
		Direction: state.round.PassDirection,
	})
	if err != nil {
		return nil, err
	}

	return game.CardStrings(cards), nil
}

func (r *Runtime) validateStartPreconditions(state *tableState, playerID game.PlayerID) string {
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

func (r *Runtime) initializeRound(state *tableState) map[game.PlayerID][]string {
	hands := make(map[game.PlayerID][]string)
	passDirection := passDirectionForRound(state.roundsStarted)
	state.roundsStarted++

	for _, player := range state.players {
		player.Hand = player.Hand[:0]
	}

	deck := defaultShuffledDeck()
	for i, card := range deck {
		seat := i % game.PlayersPerTable
		state.players[seat].Hand = append(state.players[seat].Hand, card)
	}

	for _, player := range state.players {
		game.SortCards(player.Hand)
		hands[player.PlayerID] = game.CardStrings(player.Hand)
	}

	roundPoints := make(map[game.PlayerID]game.Points, len(state.players))
	for _, player := range state.players {
		roundPoints[player.PlayerID] = game.Points(0)
	}

	state.round = &roundState{
		Phase:           roundPhasePassing,
		PassDirection:   passDirection,
		PassSubmissions: make(map[game.PlayerID][]game.Card, len(state.players)),
		PassSent:        make(map[game.PlayerID][]game.Card, len(state.players)),
		PassReceived:    make(map[game.PlayerID][]game.Card, len(state.players)),
		PassReady:       make(map[game.PlayerID]bool, len(state.players)),
		TrickNumber:     0,
		TurnSeat:        0,
		HeartsBroken:    false,
		Trick:           nil,
		RoundPoints:     roundPoints,
	}

	return hands
}

func defaultShuffledDeck() []game.Card {
	deck := game.BuildDeck()
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	game.Shuffle(deck, rng)
	return deck
}

func (r *Runtime) publishRoundStart(state *tableState, hands map[game.PlayerID][]string) {
	r.publishPublic(protocol.EventGameStarted, struct{}{})
	for playerID, cards := range hands {
		r.publishPrivate(playerID, protocol.EventHandUpdated, protocol.HandUpdatedData{Cards: cards})
	}

	if state.round != nil && state.round.PassDirection != passDirectionHold {
		r.publishPublic(protocol.EventPassSubmitted, protocol.PassStatusData{
			Submitted: 0,
			Total:     game.PlayersPerTable,
			Direction: state.round.PassDirection,
		})
	}
}

func currentTrickCards(round *roundState) []game.Card {
	cards := make([]game.Card, len(round.Trick))
	for i, played := range round.Trick {
		cards[i] = played.Card
	}
	return cards
}

func currentTrickPlays(round *roundState) []game.Play {
	plays := make([]game.Play, len(round.Trick))
	for i, played := range round.Trick {
		plays[i] = game.Play{PlayerID: played.PlayerID, Card: played.Card}
	}
	return plays
}

func (r *Runtime) applyValidatedPlay(round *roundState, player *playerState, card game.Card) (playUpdate, error) {
	updatedHand, removed := game.RemoveCard(player.Hand, card)
	if !removed {
		return playUpdate{}, fmt.Errorf("card not found in hand")
	}
	player.Hand = updatedHand

	breaksHearts := card.Suit == game.SuitHearts && !round.HeartsBroken
	if card.Suit == game.SuitHearts {
		round.HeartsBroken = true
	}

	round.Trick = append(round.Trick, trickPlay{Seat: player.Seat, PlayerID: player.PlayerID, Card: card})

	return playUpdate{
		playerID:    player.PlayerID,
		cardPlayed:  protocol.CardPlayedData{PlayerID: player.PlayerID, Card: card.String(), BreaksHearts: breaksHearts},
		handUpdated: protocol.HandUpdatedData{Cards: game.CardStrings(player.Hand)},
	}, nil
}

func (r *Runtime) publishPlayAndHand(update playUpdate) {
	r.publishPublic(protocol.EventCardPlayed, update.cardPlayed)
	r.publishPrivate(update.playerID, protocol.EventHandUpdated, update.handUpdated)
}

func (r *Runtime) publishTurn(playerID game.PlayerID, trickNumber int) {
	r.publishPublic(protocol.EventTurnChanged, protocol.TurnChangedData{PlayerID: playerID, TrickNumber: trickNumber})
	r.publishPrivate(playerID, protocol.EventYourTurn, protocol.YourTurnData{PlayerID: playerID, TrickNumber: trickNumber})
}

func (r *Runtime) completeRound(state *tableState) protocol.RoundCompletedData {
	adjustedRound := game.ApplyShootTheMoon(state.round.RoundPoints)
	for playerID, playerPoints := range adjustedRound {
		state.totals[playerID] += playerPoints
	}
	state.roundHistory = append(state.roundHistory, copyPoints(adjustedRound))

	return protocol.RoundCompletedData{
		RoundPoints: copyPoints(adjustedRound),
		TotalPoints: copyPoints(state.totals),
	}
}

const gameOverThreshold = game.Points(100)

func computeWinners(totals map[game.PlayerID]game.Points) []game.PlayerID {
	var minPts game.Points
	first := true
	for _, pts := range totals {
		if first || pts < minPts {
			minPts = pts
			first = false
		}
	}

	winners := make([]game.PlayerID, 0, len(totals))
	for playerID, pts := range totals {
		if pts == minPts {
			winners = append(winners, playerID)
		}
	}
	sort.Slice(winners, func(i, j int) bool { return winners[i] < winners[j] })
	return winners
}

func (r *Runtime) maybeEndGame(state *tableState) {
	for _, pts := range state.totals {
		if pts >= gameOverThreshold {
			state.gameOver = true
			r.publishPublic(protocol.EventGameOver, protocol.GameOverData{
				FinalScores: copyPoints(state.totals),
				Winners:     computeWinners(state.totals),
			})
			return
		}
	}
}

func (r *Runtime) buildSnapshot(state *tableState, forPlayer game.PlayerID) Snapshot {
	players := make([]PlayerSnapshot, 0, len(state.players))
	for _, player := range state.players {
		players = append(players, PlayerSnapshot{
			PlayerID: player.PlayerID,
			Name:     player.Name,
			Seat:     player.Seat,
			IsBot:    player.IsBot,
		})
	}
	sort.Slice(players, func(i, j int) bool {
		return players[i].Seat < players[j].Seat
	})

	snapshot := Snapshot{
		TableID:      r.tableID,
		Players:      players,
		Started:      state.round != nil,
		Phase:        "",
		HandSizes:    map[game.PlayerID]int{},
		RoundPoints:  map[game.PlayerID]game.Points{},
		RoundHistory: copyRoundHistory(state.roundHistory),
		TotalPoints:  copyPoints(state.totals),
		GameOver:     state.gameOver,
	}

	if state.gameOver {
		snapshot.Winners = computeWinners(state.totals)
	}

	for _, player := range state.players {
		snapshot.HandSizes[player.PlayerID] = len(player.Hand)
	}

	if state.round != nil {
		snapshot.Phase = string(state.round.Phase)
		snapshot.TrickNumber = state.round.TrickNumber
		if state.round.Phase == roundPhasePlaying {
			snapshot.TurnPlayerID = state.players[state.round.TurnSeat].PlayerID
		}
		snapshot.HeartsBroken = state.round.HeartsBroken
		snapshot.PassDirection = state.round.PassDirection
		snapshot.PassSubmittedCount = len(state.round.PassSubmissions)

		readyCount := 0
		for _, ready := range state.round.PassReady {
			if ready {
				readyCount++
			}
		}
		snapshot.PassReadyCount = readyCount

		snapshot.CurrentTrick = make([]string, 0, len(state.round.Trick))
		snapshot.TrickPlays = make([]TrickPlaySnapshot, 0, len(state.round.Trick))
		for _, played := range state.round.Trick {
			snapshot.CurrentTrick = append(snapshot.CurrentTrick, played.Card.String())
			player := state.playersByID[played.PlayerID]
			name := played.PlayerID.String()
			if player != nil {
				name = player.Name
			}
			snapshot.TrickPlays = append(snapshot.TrickPlays, TrickPlaySnapshot{
				PlayerID: played.PlayerID,
				Name:     name,
				Seat:     played.Seat,
				Card:     played.Card.String(),
			})
		}
		snapshot.RoundPoints = copyPoints(state.round.RoundPoints)
	}

	if forPlayer != "" {
		if player := state.playersByID[forPlayer]; player != nil {
			snapshot.Hand = game.CardStrings(player.Hand)
			if state.round != nil {
				if _, submitted := state.round.PassSubmissions[forPlayer]; submitted {
					snapshot.PassSubmitted = true
				}
				snapshot.PassReady = state.round.PassReady[forPlayer]
				snapshot.PassSent = game.CardStrings(state.round.PassSent[forPlayer])
				snapshot.PassReceived = game.CardStrings(state.round.PassReceived[forPlayer])
			}
		}
	}

	return snapshot
}

func (r *Runtime) publishPublic(eventType protocol.EventType, payload any) {
	r.emit(StreamEvent{Type: eventType, Data: payload})
}

func (r *Runtime) publishPrivate(playerID game.PlayerID, eventType protocol.EventType, payload any) {
	r.emit(StreamEvent{Type: eventType, Data: payload, PrivateTo: playerID})
}

func (r *Runtime) emit(event StreamEvent) {
	r.subsMu.RLock()
	defer r.subsMu.RUnlock()

	for _, sub := range r.subs {
		select {
		case sub <- event:
		default:
		}
	}
}

func copyPoints(source map[game.PlayerID]game.Points) map[game.PlayerID]game.Points {
	out := make(map[game.PlayerID]game.Points, len(source))
	for key, value := range source {
		out[key] = value
	}
	return out
}

func copyRoundHistory(source []map[game.PlayerID]game.Points) []map[game.PlayerID]game.Points {
	out := make([]map[game.PlayerID]game.Points, 0, len(source))
	for _, entry := range source {
		out = append(out, copyPoints(entry))
	}

	return out
}

var defaultBotNames = []string{
	"Ada",
	"Linus",
	"Grace",
	"Ken",
	"Dennis",
	"Margaret",
	"Alan",
	"Radia",
	"Barbara",
	"Edsger",
	"Anita",
	"Claude",
}

func randomBotName() string {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	return defaultBotNames[rng.Intn(len(defaultBotNames))]
}
