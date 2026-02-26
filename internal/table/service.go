package table

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"github.com/JHK/hearts/internal/game"
	"github.com/JHK/hearts/internal/protocol"
	"github.com/nats-io/nats.go"
)

const (
	seatsPerTable  = 4
	cardsPerPlayer = 13
)

type Service struct {
	mu sync.Mutex

	tableID string
	nc      *nats.Conn
	rng     *rand.Rand

	players    []*playerState
	playerByID map[string]*playerState

	started      bool
	currentTurn  int
	trickNumber  int
	heartsBroken bool
	trick        []playedCard
	roundPoints  []int
	totalPoints  []int
}

type playerState struct {
	ID   string
	Name string
	Seat int
	Hand []game.Card
}

type playedCard struct {
	Seat int
	Card game.Card
}

func NewService(tableID string, nc *nats.Conn) *Service {
	return &Service{
		tableID:    tableID,
		nc:         nc,
		rng:        rand.New(rand.NewSource(time.Now().UnixNano())),
		playerByID: make(map[string]*playerState),
	}
}

func (s *Service) Register() error {
	if _, err := s.nc.Subscribe(protocol.DiscoverSubject(), s.handleDiscover); err != nil {
		return err
	}

	if _, err := s.nc.Subscribe(protocol.JoinSubject(s.tableID), s.handleJoin); err != nil {
		return err
	}

	if _, err := s.nc.Subscribe(protocol.StartSubject(s.tableID), s.handleStart); err != nil {
		return err
	}

	if _, err := s.nc.Subscribe(protocol.PlaySubject(s.tableID), s.handlePlay); err != nil {
		return err
	}

	return nil
}

func (s *Service) handleDiscover(msg *nats.Msg) {
	s.mu.Lock()
	defer s.mu.Unlock()

	response := protocol.DiscoverResponse{
		Tables: []protocol.TableInfo{{
			TableID:    s.tableID,
			Players:    len(s.players),
			MaxPlayers: seatsPerTable,
			Started:    s.started,
		}},
	}

	s.reply(msg, response)
}

func (s *Service) handleJoin(msg *nats.Msg) {
	var req protocol.JoinRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		s.reply(msg, protocol.JoinResponse{Accepted: false, Reason: "invalid join payload"})
		return
	}

	if req.PlayerID == "" || req.Name == "" {
		s.reply(msg, protocol.JoinResponse{Accepted: false, Reason: "player_id and name are required"})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if player, ok := s.playerByID[req.PlayerID]; ok {
		s.reply(msg, protocol.JoinResponse{
			Accepted: true,
			TableID:  s.tableID,
			PlayerID: player.ID,
			Seat:     player.Seat,
		})
		return
	}

	if s.started {
		s.reply(msg, protocol.JoinResponse{Accepted: false, Reason: "round already started"})
		return
	}

	if len(s.players) >= seatsPerTable {
		s.reply(msg, protocol.JoinResponse{Accepted: false, Reason: "table is full"})
		return
	}

	seat := len(s.players)
	player := &playerState{
		ID:   req.PlayerID,
		Name: req.Name,
		Seat: seat,
	}

	s.players = append(s.players, player)
	s.playerByID[player.ID] = player

	s.reply(msg, protocol.JoinResponse{
		Accepted: true,
		TableID:  s.tableID,
		PlayerID: player.ID,
		Seat:     player.Seat,
	})

	s.publishEvent(protocol.EventPlayerJoined, protocol.PlayerJoinedData{
		Player: protocol.PlayerInfo{PlayerID: player.ID, Name: player.Name, Seat: player.Seat},
	})
}

func (s *Service) handleStart(msg *nats.Msg) {
	var req protocol.StartRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		s.reply(msg, protocol.CommandResponse{Accepted: false, Reason: "invalid start payload"})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.playerByID[req.PlayerID]; !ok {
		s.reply(msg, protocol.CommandResponse{Accepted: false, Reason: "player is not seated"})
		return
	}

	if s.started {
		s.reply(msg, protocol.CommandResponse{Accepted: false, Reason: "round already started"})
		return
	}

	if len(s.players) != seatsPerTable {
		s.reply(msg, protocol.CommandResponse{Accepted: false, Reason: "need 4 players to start"})
		return
	}

	if err := s.startRound(); err != nil {
		s.reply(msg, protocol.CommandResponse{Accepted: false, Reason: err.Error()})
		return
	}

	s.reply(msg, protocol.CommandResponse{Accepted: true})
}

func (s *Service) handlePlay(msg *nats.Msg) {
	var req protocol.PlayCardRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		s.reply(msg, protocol.CommandResponse{Accepted: false, Reason: "invalid play payload"})
		return
	}

	card, err := game.ParseCard(req.Card)
	if err != nil {
		s.reply(msg, protocol.CommandResponse{Accepted: false, Reason: "invalid card"})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		s.reply(msg, protocol.CommandResponse{Accepted: false, Reason: "round has not started"})
		return
	}

	player, ok := s.playerByID[req.PlayerID]
	if !ok {
		s.reply(msg, protocol.CommandResponse{Accepted: false, Reason: "player is not seated"})
		return
	}

	if player.Seat != s.currentTurn {
		s.reply(msg, protocol.CommandResponse{Accepted: false, Reason: "not your turn"})
		return
	}

	if !game.ContainsCard(player.Hand, card) {
		s.reply(msg, protocol.CommandResponse{Accepted: false, Reason: "card not in hand"})
		return
	}

	if len(s.trick) == 0 {
		if err := game.CanLeadCard(player.Hand, card, s.heartsBroken, s.trickNumber); err != nil {
			s.reply(msg, protocol.CommandResponse{Accepted: false, Reason: err.Error()})
			return
		}
	} else {
		leadSuit := s.trick[0].Card.Suit
		if err := game.CanPlayCard(player.Hand, card, leadSuit, s.trickNumber); err != nil {
			s.reply(msg, protocol.CommandResponse{Accepted: false, Reason: err.Error()})
			return
		}
	}

	player.Hand, _ = game.RemoveCard(player.Hand, card)
	if card.Suit == game.Hearts {
		s.heartsBroken = true
	}

	s.trick = append(s.trick, playedCard{Seat: player.Seat, Card: card})
	s.publishEvent(protocol.EventCardPlayed, protocol.CardPlayedData{PlayerID: player.ID, Card: card.String()})
	s.publishPlayerEvent(player.ID, protocol.EventHandUpdated, protocol.HandUpdatedData{Cards: game.CardsToStrings(player.Hand)})

	if len(s.trick) < seatsPerTable {
		s.currentTurn = (s.currentTurn + 1) % seatsPerTable
		s.publishEvent(protocol.EventTurnChanged, protocol.TurnChangedData{
			PlayerID:    s.players[s.currentTurn].ID,
			TrickNumber: s.trickNumber,
		})
		s.reply(msg, protocol.CommandResponse{Accepted: true})
		return
	}

	trickCards := make([]game.Card, 0, len(s.trick))
	for _, play := range s.trick {
		trickCards = append(trickCards, play.Card)
	}

	winnerIdx := game.TrickWinnerIndex(trickCards[0].Suit, trickCards)
	if winnerIdx < 0 {
		s.reply(msg, protocol.CommandResponse{Accepted: false, Reason: "could not determine trick winner"})
		return
	}

	winnerSeat := s.trick[winnerIdx].Seat
	points := game.TrickPoints(trickCards)
	s.roundPoints[winnerSeat] += points

	s.publishEvent(protocol.EventTrickCompleted, protocol.TrickCompletedData{
		WinnerPlayerID: s.players[winnerSeat].ID,
		Points:         points,
		TrickNumber:    s.trickNumber,
	})

	s.trick = nil
	s.trickNumber++

	if s.trickNumber >= cardsPerPlayer {
		s.finishRound()
		s.reply(msg, protocol.CommandResponse{Accepted: true})
		return
	}

	s.currentTurn = winnerSeat
	s.publishEvent(protocol.EventTurnChanged, protocol.TurnChangedData{
		PlayerID:    s.players[s.currentTurn].ID,
		TrickNumber: s.trickNumber,
	})

	s.reply(msg, protocol.CommandResponse{Accepted: true})
}

func (s *Service) startRound() error {
	deck := game.NewShuffledDeck(s.rng)
	if len(deck) != seatsPerTable*cardsPerPlayer {
		return fmt.Errorf("unexpected deck size: %d", len(deck))
	}

	for _, player := range s.players {
		player.Hand = player.Hand[:0]
	}

	for i, card := range deck {
		seat := i % seatsPerTable
		s.players[seat].Hand = append(s.players[seat].Hand, card)
	}

	startingSeat := -1
	for _, player := range s.players {
		game.SortCards(player.Hand)
		if game.ContainsCard(player.Hand, game.Card{Suit: game.Clubs, Rank: 2}) {
			startingSeat = player.Seat
		}
	}

	if startingSeat < 0 {
		return fmt.Errorf("could not find 2C holder")
	}

	s.started = true
	s.currentTurn = startingSeat
	s.trickNumber = 0
	s.heartsBroken = false
	s.trick = nil
	s.roundPoints = make([]int, seatsPerTable)
	if len(s.totalPoints) == 0 {
		s.totalPoints = make([]int, seatsPerTable)
	}

	players := make([]protocol.PlayerInfo, 0, len(s.players))
	for _, player := range s.players {
		players = append(players, protocol.PlayerInfo{PlayerID: player.ID, Name: player.Name, Seat: player.Seat})
	}

	s.publishEvent(protocol.EventGameStarted, protocol.GameStartedData{Players: players})
	s.publishEvent(protocol.EventTurnChanged, protocol.TurnChangedData{PlayerID: s.players[s.currentTurn].ID, TrickNumber: 0})

	for _, player := range s.players {
		s.publishPlayerEvent(player.ID, protocol.EventHandUpdated, protocol.HandUpdatedData{Cards: game.CardsToStrings(player.Hand)})
	}

	return nil
}

func (s *Service) finishRound() {
	round := game.ApplyShootMoon(s.roundPoints)
	roundByPlayer := make(map[string]int, len(s.players))
	totalByPlayer := make(map[string]int, len(s.players))

	for _, player := range s.players {
		s.totalPoints[player.Seat] += round[player.Seat]
		roundByPlayer[player.ID] = round[player.Seat]
		totalByPlayer[player.ID] = s.totalPoints[player.Seat]
		player.Hand = nil
	}

	s.publishEvent(protocol.EventRoundCompleted, protocol.RoundCompletedData{
		RoundPoints: roundByPlayer,
		TotalPoints: totalByPlayer,
	})

	s.started = false
	s.trick = nil
	s.trickNumber = 0
	s.heartsBroken = false
	s.roundPoints = nil
}

func (s *Service) publishEvent(eventType string, payload any) {
	encoded, err := protocol.EncodeEvent(s.tableID, eventType, payload)
	if err != nil {
		slog.Error("failed to encode event", "event_type", eventType, "error", err)
		return
	}

	if err := s.nc.Publish(protocol.EventsSubject(s.tableID), encoded); err != nil {
		slog.Error("failed to publish event", "event_type", eventType, "error", err)
	}
}

func (s *Service) publishPlayerEvent(playerID, eventType string, payload any) {
	encoded, err := protocol.EncodeEvent(s.tableID, eventType, payload)
	if err != nil {
		slog.Error("failed to encode private event", "event_type", eventType, "error", err)
		return
	}

	if err := s.nc.Publish(protocol.PlayerEventsSubject(s.tableID, playerID), encoded); err != nil {
		slog.Error("failed to publish private event", "event_type", eventType, "error", err)
	}
}

func (s *Service) reply(msg *nats.Msg, response any) {
	if msg.Reply == "" {
		return
	}

	encoded, err := json.Marshal(response)
	if err != nil {
		slog.Error("failed to encode response", "error", err)
		return
	}

	if err := s.nc.Publish(msg.Reply, encoded); err != nil {
		slog.Error("failed to publish response", "error", err)
	}
}
