package table

import (
	"fmt"
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
	TableID      string                        `json:"table_id"`
	Players      []PlayerSnapshot              `json:"players"`
	Started      bool                          `json:"started"`
	TrickNumber  int                           `json:"trick_number"`
	TurnPlayerID game.PlayerID                 `json:"turn_player_id"`
	HeartsBroken bool                          `json:"hearts_broken"`
	CurrentTrick []string                      `json:"current_trick"`
	TrickPlays   []TrickPlaySnapshot           `json:"trick_plays"`
	Hand         []string                      `json:"hand"`
	HandSizes    map[game.PlayerID]int         `json:"hand_sizes"`
	RoundPoints  map[game.PlayerID]game.Points `json:"round_points"`
	TotalPoints  map[game.PlayerID]game.Points `json:"total_points"`
}

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

	round         *roundState
	nextPlayerSeq int
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

type infoCommand struct {
	reply chan protocol.TableInfo
}

type botTurnCommand struct {
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
			case addBotCommand:
				added, err := r.handleAddBot(state, cmd.strategyName)
				cmd.reply <- addBotResult{added: added, err: err}
			case snapshotCommand:
				cmd.reply <- r.buildSnapshot(state, cmd.forPlayer)
			case infoCommand:
				cmd.reply <- protocol.TableInfo{
					TableID:    r.tableID,
					Players:    len(state.players),
					MaxPlayers: game.PlayersPerTable,
					Started:    state.round != nil,
				}
			case botTurnCommand:
				r.handleBotTurn(state, cmd.playerID)
			}
		}
	}
}

func (r *Runtime) handleJoin(state *tableState, name, token string) protocol.JoinResponse {
	token = strings.TrimSpace(token)
	if token == "" {
		return protocol.JoinResponse{Accepted: false, Reason: "player token is required"}
	}

	if existing := state.playersByToken[token]; existing != nil {
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

	hands, turnPlayerID := r.initializeRound(state)
	r.publishRoundStart(hands, turnPlayerID)
	r.scheduleBotTurn(state, turnPlayerID)

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

func (r *Runtime) handleBotTurn(state *tableState, playerID game.PlayerID) {
	if state.round == nil {
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

func (r *Runtime) validateStartPreconditions(state *tableState, playerID game.PlayerID) string {
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

func (r *Runtime) initializeRound(state *tableState) (map[game.PlayerID][]string, game.PlayerID) {
	hands := make(map[game.PlayerID][]string)
	for _, player := range state.players {
		player.Hand = player.Hand[:0]
	}

	deck := defaultShuffledDeck()
	for i, card := range deck {
		seat := i % game.PlayersPerTable
		state.players[seat].Hand = append(state.players[seat].Hand, card)
	}

	startSeat := 0
	twoClubs := game.Card{Suit: game.SuitClubs, Rank: 2}
	for _, player := range state.players {
		game.SortCards(player.Hand)
		hands[player.PlayerID] = game.CardStrings(player.Hand)
		if game.ContainsCard(player.Hand, twoClubs) {
			startSeat = player.Seat
		}
	}

	roundPoints := make(map[game.PlayerID]game.Points, len(state.players))
	for _, player := range state.players {
		roundPoints[player.PlayerID] = game.Points(0)
	}

	state.round = &roundState{
		TrickNumber:  0,
		TurnSeat:     startSeat,
		HeartsBroken: false,
		Trick:        nil,
		RoundPoints:  roundPoints,
	}

	return hands, state.players[startSeat].PlayerID
}

func defaultShuffledDeck() []game.Card {
	deck := game.BuildDeck()
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	game.Shuffle(deck, rng)
	return deck
}

func (r *Runtime) publishRoundStart(hands map[game.PlayerID][]string, turnPlayerID game.PlayerID) {
	r.publishPublic(protocol.EventGameStarted, struct{}{})
	for playerID, cards := range hands {
		r.publishPrivate(playerID, protocol.EventHandUpdated, protocol.HandUpdatedData{Cards: cards})
	}
	r.publishTurn(turnPlayerID, 0)
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

	if card.Suit == game.SuitHearts {
		round.HeartsBroken = true
	}

	round.Trick = append(round.Trick, trickPlay{Seat: player.Seat, PlayerID: player.PlayerID, Card: card})

	return playUpdate{
		playerID:    player.PlayerID,
		cardPlayed:  protocol.CardPlayedData{PlayerID: player.PlayerID, Card: card.String()},
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

	return protocol.RoundCompletedData{
		RoundPoints: copyPoints(adjustedRound),
		TotalPoints: copyPoints(state.totals),
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
		TableID:     r.tableID,
		Players:     players,
		Started:     state.round != nil,
		HandSizes:   map[game.PlayerID]int{},
		RoundPoints: map[game.PlayerID]game.Points{},
		TotalPoints: copyPoints(state.totals),
	}

	for _, player := range state.players {
		snapshot.HandSizes[player.PlayerID] = len(player.Hand)
	}

	if state.round != nil {
		snapshot.TrickNumber = state.round.TrickNumber
		snapshot.TurnPlayerID = state.players[state.round.TurnSeat].PlayerID
		snapshot.HeartsBroken = state.round.HeartsBroken
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
