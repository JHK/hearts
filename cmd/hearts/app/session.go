package app

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/JHK/hearts/internal/game"
	"github.com/JHK/hearts/internal/player/bot"
	"github.com/JHK/hearts/internal/protocol"
	"github.com/JHK/hearts/internal/server"
	natswire "github.com/JHK/hearts/internal/transport/nats"
	"github.com/nats-io/nats.go"
)

type Status struct {
	PlayerID    game.PlayerID
	PlayerName  string
	Connected   bool
	ConnectedTo string
	TableID     string
	LocalBusURL string
}

type Session struct {
	name       string
	defaultURL string

	clientNC *nats.Conn
	host     *server.Local

	currentTable    string
	currentPlayerID game.PlayerID
	participant     *natswire.ParticipantClient
	hand            handState
	bots            []*ephemeralBot
}

type AddedBot struct {
	JoinResponse protocol.JoinResponse
	Name         string
	Strategy     string
}

type ephemeralBot struct {
	conn    *nats.Conn
	runtime *bot.Runtime
}

func NewSession(name, defaultURL string) *Session {
	return &Session{name: name, defaultURL: defaultURL}
}

func (s *Session) Connect(url string) error {
	if s.clientNC != nil {
		s.setParticipant(nil, "", "")
		s.clientNC.Close()
		s.clientNC = nil
	}

	nc, err := nats.Connect(url)
	if err != nil {
		return err
	}

	s.clientNC = nc
	return nil
}

func (s *Session) OpenTable(tableID, host string, port int, handlers natswire.ParticipantEventHandlers) (protocol.JoinResponse, error) {
	if tableID == "" {
		return protocol.JoinResponse{}, fmt.Errorf("table id is required")
	}

	if s.host == nil {
		local, err := server.Open(host, port)
		if err != nil {
			return protocol.JoinResponse{}, err
		}
		s.host = local
	} else if s.host.Host() != host || s.host.Port() != port {
		return protocol.JoinResponse{}, fmt.Errorf("local bus already running on %s", s.host.URL())
	}

	if err := s.host.EnsureTable(tableID); err != nil {
		return protocol.JoinResponse{}, err
	}

	if err := s.Connect(s.host.URL()); err != nil {
		return protocol.JoinResponse{}, fmt.Errorf("connect local bus: %w", err)
	}

	return s.JoinTable(tableID, handlers)
}

func (s *Session) DiscoverTables() ([]protocol.TableInfo, error) {
	if err := s.ensureClientConnection(); err != nil {
		return nil, err
	}

	return natswire.DiscoverTables(s.clientNC, 750*time.Millisecond)
}

func (s *Session) JoinTable(tableID string, handlers natswire.ParticipantEventHandlers) (protocol.JoinResponse, error) {
	if err := s.ensureClientConnection(); err != nil {
		return protocol.JoinResponse{}, err
	}

	joinResp, err := joinPlayer(s.clientNC, tableID, s.name)
	if err != nil {
		return protocol.JoinResponse{}, err
	}

	participant := natswire.NewParticipantClient(s.clientNC, tableID, joinResp.PlayerID)
	handlers = s.withHandSync(handlers)

	if err := participant.Start(handlers); err != nil {
		return protocol.JoinResponse{}, err
	}

	s.setParticipant(participant, tableID, joinResp.PlayerID)
	return joinResp, nil
}

func (s *Session) StartRound() error {
	if err := s.ensureJoinedTable(); err != nil {
		return err
	}

	return s.participant.StartRound()
}

func (s *Session) AddBot(strategyName string) (AddedBot, error) {
	if err := s.ensureJoinedTable(); err != nil {
		return AddedBot{}, err
	}

	strategyKind, err := bot.ParseStrategyKind(strategyName)
	if err != nil {
		return AddedBot{}, err
	}
	strategy := strategyKind.New()
	strategyName = string(strategyKind)
	name := randomBotName()

	botConn, err := nats.Connect(s.clientNC.ConnectedUrl())
	if err != nil {
		return AddedBot{}, err
	}

	joinResp, err := joinPlayer(botConn, s.currentTable, name)
	if err != nil {
		botConn.Close()
		return AddedBot{}, err
	}

	runtime := bot.NewRuntime(botConn, s.currentTable, joinResp.PlayerID, strategy)
	if err := runtime.Start(); err != nil {
		botConn.Close()
		return AddedBot{}, fmt.Errorf("start bot runtime: %w", err)
	}

	s.bots = append(s.bots, &ephemeralBot{conn: botConn, runtime: runtime})

	return AddedBot{JoinResponse: joinResp, Name: name, Strategy: strategyName}, nil
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

func (s *Session) TableStats() (protocol.TableInfo, error) {
	if s.currentTable == "" {
		return protocol.TableInfo{}, fmt.Errorf("join a table first")
	}
	tables, err := s.DiscoverTables()
	if err != nil {
		return protocol.TableInfo{}, err
	}
	for _, table := range tables {
		if table.TableID == s.currentTable {
			return table, nil
		}
	}
	return protocol.TableInfo{}, fmt.Errorf("table %s not found", s.currentTable)
}

func (s *Session) PlayCard(card string) error {
	if err := s.ensureJoinedTable(); err != nil {
		return err
	}

	return s.participant.PlayCard(card)
}

func (s *Session) HandSnapshot() []string {
	return s.hand.snapshot()
}

func (s *Session) Shutdown() {
	s.setParticipant(nil, "", "")
	s.stopBots()

	if s.clientNC != nil {
		s.clientNC.Close()
		s.clientNC = nil
	}

	if s.host != nil {
		s.host.Close()
		s.host = nil
	}
}

func (s *Session) stopBots() {
	for _, bot := range s.bots {
		bot.runtime.Stop()
		bot.conn.Close()
	}
	s.bots = nil
}

func (s *Session) Status() Status {
	status := Status{
		PlayerID:   s.currentPlayerID,
		PlayerName: s.name,
		TableID:    s.currentTable,
	}

	if s.clientNC != nil {
		status.Connected = true
		status.ConnectedTo = s.clientNC.ConnectedUrl()
	}

	if s.host != nil {
		status.LocalBusURL = s.host.URL()
	}

	return status
}

func (s *Session) ensureClientConnection() error {
	if s.clientNC != nil {
		return nil
	}

	if s.host != nil {
		return s.Connect(s.host.URL())
	}

	if s.defaultURL == "" {
		return fmt.Errorf("not connected; use open or connect")
	}

	if err := s.Connect(s.defaultURL); err != nil {
		return fmt.Errorf("not connected; use open or connect: %w", err)
	}

	return nil
}

func (s *Session) ensureJoinedTable() error {
	if err := s.ensureClientConnection(); err != nil {
		return err
	}
	if s.currentTable == "" || s.participant == nil {
		return fmt.Errorf("join a table first")
	}
	return nil
}

func (s *Session) setParticipant(participant *natswire.ParticipantClient, tableID string, playerID game.PlayerID) {
	if s.participant != nil {
		s.participant.Stop()
	}
	s.participant = participant
	s.currentTable = tableID
	s.currentPlayerID = playerID
	s.hand.update(nil)
}

type handState struct {
	mu    sync.Mutex
	cards []string
}

func (h *handState) update(cards []string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.cards = append(h.cards[:0], cards...)
	sort.Strings(h.cards)
}

func (h *handState) snapshot() []string {
	h.mu.Lock()
	defer h.mu.Unlock()

	out := make([]string, len(h.cards))
	copy(out, h.cards)
	return out
}

func joinPlayer(conn *nats.Conn, tableID, name string) (protocol.JoinResponse, error) {
	joinClient := natswire.NewParticipantClient(conn, tableID, "")
	joinResp, err := joinClient.Join(name)
	if err != nil {
		return protocol.JoinResponse{}, err
	}
	if !joinResp.Accepted {
		return protocol.JoinResponse{}, fmt.Errorf("%s", joinResp.Reason)
	}
	return joinResp, nil
}

func (s *Session) withHandSync(handlers natswire.ParticipantEventHandlers) natswire.ParticipantEventHandlers {
	originalHandHandler := handlers.OnHandUpdated
	handlers.OnHandUpdated = func(playerID game.PlayerID, data protocol.HandUpdatedData) {
		s.hand.update(data.Cards)
		if originalHandHandler != nil {
			originalHandHandler(playerID, data)
		}
	}
	return handlers
}
