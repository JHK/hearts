package server

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/JHK/hearts/internal/game"
	"github.com/JHK/hearts/internal/protocol"
	natswire "github.com/JHK/hearts/internal/transport/nats"
	"github.com/nats-io/nats.go"
)

type tableRuntime struct {
	tableID   string
	endpoint  *natswire.TableEndpoint
	drawDeck  func() []game.Card
	commands  chan any
	stop      chan struct{}
	stopOnce  sync.Once
	stoppedCh chan struct{}
}

type tableState struct {
	players     []*playerState
	playersByID map[game.PlayerID]*playerState
	totals      map[game.PlayerID]game.Points

	round         *roundState
	nextPlayerSeq int
}

type discoverCommand struct {
	reply chan protocol.DiscoverResponse
}

type joinCommand struct {
	request protocol.JoinRequest
	reply   chan protocol.JoinResponse
}

type startCommand struct {
	request protocol.StartRequest
	reply   chan protocol.CommandResponse
}

type playCommand struct {
	request protocol.PlayCardRequest
	reply   chan protocol.CommandResponse
}

type playerState struct {
	PlayerID game.PlayerID
	Name     string
	Seat     int
	Hand     []game.Card
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

func newTableRuntime(tableID string, nc *nats.Conn) *tableRuntime {
	return newTableRuntimeWithDeck(tableID, nc, nil)
}

func newTableRuntimeWithDeck(tableID string, nc *nats.Conn, drawDeck func() []game.Card) *tableRuntime {
	if drawDeck == nil {
		drawDeck = defaultShuffledDeck
	}

	t := &tableRuntime{
		tableID:   tableID,
		drawDeck:  drawDeck,
		commands:  make(chan any),
		stop:      make(chan struct{}),
		stoppedCh: make(chan struct{}),
	}

	t.endpoint = natswire.NewTableEndpoint(nc, tableID, natswire.TableEndpointHandlers{
		OnDiscover: t.onDiscover,
		OnJoin:     t.onJoin,
		OnStart:    t.onStart,
		OnPlay:     t.onPlay,
	})

	go t.run()

	return t
}

func defaultShuffledDeck() []game.Card {
	deck := game.BuildDeck()
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	game.Shuffle(deck, rng)
	return deck
}

func (t *tableRuntime) register() error {
	return t.endpoint.Register()
}

func (t *tableRuntime) shutdown() {
	t.stopOnce.Do(func() {
		t.endpoint.Stop()
		close(t.stop)
		<-t.stoppedCh
	})
}

func (t *tableRuntime) onDiscover() protocol.DiscoverResponse {
	reply := make(chan protocol.DiscoverResponse, 1)
	if !t.submit(discoverCommand{reply: reply}) {
		return protocol.DiscoverResponse{}
	}
	return <-reply
}

func (t *tableRuntime) onJoin(request protocol.JoinRequest) protocol.JoinResponse {
	reply := make(chan protocol.JoinResponse, 1)
	if !t.submit(joinCommand{request: request, reply: reply}) {
		return protocol.JoinResponse{Accepted: false, Reason: "table is stopping"}
	}
	return <-reply
}

func (t *tableRuntime) onStart(request protocol.StartRequest) protocol.CommandResponse {
	reply := make(chan protocol.CommandResponse, 1)
	if !t.submit(startCommand{request: request, reply: reply}) {
		return protocol.CommandResponse{Accepted: false, Reason: "table is stopping"}
	}
	return <-reply
}

func (t *tableRuntime) onPlay(request protocol.PlayCardRequest) protocol.CommandResponse {
	reply := make(chan protocol.CommandResponse, 1)
	if !t.submit(playCommand{request: request, reply: reply}) {
		return protocol.CommandResponse{Accepted: false, Reason: "table is stopping"}
	}
	return <-reply
}

func (t *tableRuntime) submit(command any) bool {
	select {
	case <-t.stop:
		return false
	case t.commands <- command:
		return true
	}
}

func (t *tableRuntime) run() {
	defer close(t.stoppedCh)

	state := &tableState{
		playersByID: make(map[game.PlayerID]*playerState),
		totals:      make(map[game.PlayerID]game.Points),
	}

	for {
		select {
		case <-t.stop:
			return
		case command := <-t.commands:
			switch cmd := command.(type) {
			case discoverCommand:
				cmd.reply <- t.handleDiscover(state)
			case joinCommand:
				cmd.reply <- t.handleJoin(state, cmd.request)
			case startCommand:
				cmd.reply <- t.handleStart(state, cmd.request)
			case playCommand:
				cmd.reply <- t.handlePlay(state, cmd.request)
			}
		}
	}
}

func (t *tableRuntime) handleDiscover(state *tableState) protocol.DiscoverResponse {
	return protocol.DiscoverResponse{
		Tables: []protocol.TableInfo{{
			TableID:    t.tableID,
			Players:    len(state.players),
			MaxPlayers: game.PlayersPerTable,
			Started:    state.round != nil,
		}},
	}
}

func (t *tableRuntime) handleJoin(state *tableState, request protocol.JoinRequest) protocol.JoinResponse {
	request.Name = strings.TrimSpace(request.Name)
	if request.Name == "" {
		request.Name = "Player"
	}

	if state.round != nil {
		return protocol.JoinResponse{Accepted: false, Reason: "round already in progress"}
	}
	if len(state.players) >= game.PlayersPerTable {
		return protocol.JoinResponse{Accepted: false, Reason: "table is full"}
	}

	playerID := t.nextPlayerID(state)
	joinedPlayer := &playerState{
		PlayerID: playerID,
		Name:     request.Name,
		Seat:     len(state.players),
	}
	state.players = append(state.players, joinedPlayer)
	state.playersByID[joinedPlayer.PlayerID] = joinedPlayer

	t.publishPublic(protocol.EventPlayerJoined, protocol.PlayerJoinedData{Player: protocol.PlayerInfo{
		PlayerID: joinedPlayer.PlayerID,
		Name:     joinedPlayer.Name,
		Seat:     joinedPlayer.Seat,
	}})

	return protocol.JoinResponse{
		Accepted: true,
		TableID:  t.tableID,
		PlayerID: joinedPlayer.PlayerID,
		Seat:     joinedPlayer.Seat,
	}
}

func (t *tableRuntime) handleStart(state *tableState, request protocol.StartRequest) protocol.CommandResponse {
	request.PlayerID = game.PlayerID(strings.TrimSpace(string(request.PlayerID)))

	if reason := t.validateStartPreconditions(state, request); reason != "" {
		return protocol.CommandResponse{Accepted: false, Reason: reason}
	}

	hands, turnPlayerID := t.initializeRound(state)
	t.publishRoundStart(hands, turnPlayerID)

	return protocol.CommandResponse{Accepted: true}
}

func (t *tableRuntime) nextPlayerID(state *tableState) game.PlayerID {
	for {
		state.nextPlayerSeq++
		candidate := game.PlayerID(fmt.Sprintf("p-%d", state.nextPlayerSeq))
		if _, exists := state.playersByID[candidate]; !exists {
			return candidate
		}
	}
}

func (t *tableRuntime) handlePlay(state *tableState, request protocol.PlayCardRequest) protocol.CommandResponse {
	card, err := game.ParseCard(request.Card)
	if err != nil {
		return protocol.CommandResponse{Accepted: false, Reason: err.Error()}
	}

	if state.round == nil {
		return protocol.CommandResponse{Accepted: false, Reason: "round is not running"}
	}

	player := state.playersByID[request.PlayerID]
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

	update, err := t.applyValidatedPlay(state.round, player, card)
	if err != nil {
		return protocol.CommandResponse{Accepted: false, Reason: err.Error()}
	}

	if len(state.round.Trick) < game.PlayersPerTable {
		nextSeat := (state.round.TurnSeat + 1) % game.PlayersPerTable
		nextPlayerID := state.players[nextSeat].PlayerID
		trickNumber := state.round.TrickNumber
		state.round.TurnSeat = nextSeat

		t.publishPlayAndHand(update)
		t.publishTurn(nextPlayerID, trickNumber)
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
		roundCompleted := t.completeRound(state)
		state.round = nil

		t.publishPlayAndHand(update)
		t.publishPublic(protocol.EventTrickCompleted, trickCompleted)
		t.publishPublic(protocol.EventRoundCompleted, roundCompleted)
		return protocol.CommandResponse{Accepted: true}
	}

	nextSeat := state.playersByID[winnerPlayerID].Seat
	nextTrick := trickNumber + 1

	state.round.Trick = nil
	state.round.TrickNumber = nextTrick
	state.round.TurnSeat = nextSeat

	t.publishPlayAndHand(update)
	t.publishPublic(protocol.EventTrickCompleted, trickCompleted)
	t.publishTurn(winnerPlayerID, nextTrick)

	return protocol.CommandResponse{Accepted: true}
}

func (t *tableRuntime) validateStartPreconditions(state *tableState, request protocol.StartRequest) string {
	if state.round != nil {
		return "round already started"
	}
	if len(state.players) == 0 {
		return "at least one player must join before start"
	}
	if state.playersByID[request.PlayerID] == nil {
		return "only seated players can start"
	}
	if len(state.players) != game.PlayersPerTable {
		return fmt.Sprintf("table requires %d players before start", game.PlayersPerTable)
	}
	return ""
}

func (t *tableRuntime) initializeRound(state *tableState) (map[game.PlayerID][]string, game.PlayerID) {
	hands := make(map[game.PlayerID][]string)
	for _, player := range state.players {
		player.Hand = player.Hand[:0]
	}

	deck := t.drawDeck()

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

func (t *tableRuntime) publishRoundStart(hands map[game.PlayerID][]string, turnPlayerID game.PlayerID) {
	t.publishPublic(protocol.EventGameStarted, struct{}{})
	for playerID, cards := range hands {
		t.publishPrivate(playerID, protocol.EventHandUpdated, protocol.HandUpdatedData{Cards: cards})
	}
	t.publishTurn(turnPlayerID, 0)
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

func (t *tableRuntime) applyValidatedPlay(round *roundState, player *playerState, card game.Card) (playUpdate, error) {
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

func (t *tableRuntime) publishPlayAndHand(update playUpdate) {
	t.publishPublic(protocol.EventCardPlayed, update.cardPlayed)
	t.publishPrivate(update.playerID, protocol.EventHandUpdated, update.handUpdated)
}

func (t *tableRuntime) publishTurn(playerID game.PlayerID, trickNumber int) {
	t.publishPublic(protocol.EventTurnChanged, protocol.TurnChangedData{PlayerID: playerID, TrickNumber: trickNumber})
	t.publishPrivate(playerID, protocol.EventYourTurn, protocol.YourTurnData{PlayerID: playerID, TrickNumber: trickNumber})
}

func (t *tableRuntime) completeRound(state *tableState) protocol.RoundCompletedData {
	adjustedRound := game.ApplyShootTheMoon(state.round.RoundPoints)
	for playerID, playerPoints := range adjustedRound {
		state.totals[playerID] += playerPoints
	}

	return protocol.RoundCompletedData{
		RoundPoints: copyPoints(adjustedRound),
		TotalPoints: copyPoints(state.totals),
	}
}

func (t *tableRuntime) publishPublic(eventType protocol.EventType, payload any) {
	_ = t.endpoint.PublishPublic(eventType, payload)
}

func (t *tableRuntime) publishPrivate(playerID game.PlayerID, eventType protocol.EventType, payload any) {
	_ = t.endpoint.PublishPrivate(playerID, eventType, payload)
}

func copyPoints(source map[game.PlayerID]game.Points) map[game.PlayerID]game.Points {
	out := make(map[game.PlayerID]game.Points, len(source))
	for key, value := range source {
		out[key] = value
	}
	return out
}
