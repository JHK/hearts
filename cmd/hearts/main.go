package main

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/JHK/hearts/internal/protocol"
	"github.com/JHK/hearts/internal/table"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

func main() {
	var (
		defaultURL = flag.String("url", "nats://127.0.0.1:4222", "default NATS URL for discover/join")
		playerID   = flag.String("player-id", "", "player id")
		name       = flag.String("name", "", "display name")
	)
	flag.Parse()

	if *playerID == "" {
		*playerID = "p-" + randomHex(4)
	}
	if *name == "" {
		*name = *playerID
	}

	app := &cliApp{
		playerID:   *playerID,
		name:       *name,
		defaultURL: *defaultURL,
		hand:       &handState{},
	}
	defer app.shutdown()

	fmt.Printf("hearts cli - player %s (%s)\n", app.name, app.playerID)
	if err := app.connect(app.defaultURL); err != nil {
		fmt.Printf("No game bus at %s yet. Use 'open' to host or 'connect <url>' to join another bus.\n", app.defaultURL)
	}
	printHelp()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if !app.runCommand(strings.Fields(line)) {
			return
		}
	}
}

type cliApp struct {
	playerID   string
	name       string
	defaultURL string

	clientNC *nats.Conn

	host *localHost

	currentTable string
	tableSub     *nats.Subscription
	playerSub    *nats.Subscription
	hand         *handState
}

type localHost struct {
	server *server.Server
	nc     *nats.Conn
	url    string
	host   string
	port   int
	tables map[string]*table.Service
}

func (a *cliApp) runCommand(parts []string) bool {
	switch strings.ToLower(parts[0]) {
	case "help":
		printHelp()
	case "connect":
		if len(parts) != 2 {
			fmt.Println("usage: connect <nats-url>")
			return true
		}
		if err := a.connect(parts[1]); err != nil {
			fmt.Printf("connect failed: %v\n", err)
		}
	case "open":
		tableID := "default"
		port := 4222
		if len(parts) >= 2 {
			tableID = parts[1]
		}
		if len(parts) >= 3 {
			parsedPort, err := strconv.Atoi(parts[2])
			if err != nil || parsedPort < 1 || parsedPort > 65535 {
				fmt.Println("usage: open [table-id] [port]")
				return true
			}
			port = parsedPort
		}

		if err := a.openTable(tableID, "127.0.0.1", port); err != nil {
			fmt.Printf("open failed: %v\n", err)
		}
	case "discover":
		tables, err := a.discoverTables()
		if err != nil {
			fmt.Printf("discover failed: %v\n", err)
			return true
		}
		if len(tables) == 0 {
			fmt.Println("No open tables discovered.")
			return true
		}
		fmt.Println("Discovered tables:")
		for _, info := range tables {
			state := "waiting"
			if info.Started {
				state = "in_round"
			}
			fmt.Printf("  - %s (%d/%d, %s)\n", info.TableID, info.Players, info.MaxPlayers, state)
		}
	case "join":
		if len(parts) != 2 {
			fmt.Println("usage: join <table-id>")
			return true
		}
		if err := a.joinTable(parts[1]); err != nil {
			fmt.Printf("join failed: %v\n", err)
		}
	case "start":
		if err := a.startRound(); err != nil {
			fmt.Printf("start rejected: %v\n", err)
		}
	case "play":
		if len(parts) != 2 {
			fmt.Println("usage: play <card>, example: play QS")
			return true
		}
		if err := a.playCard(strings.ToUpper(parts[1])); err != nil {
			fmt.Printf("play rejected: %v\n", err)
		}
	case "hand":
		cards := a.hand.snapshot()
		if len(cards) == 0 {
			fmt.Println("hand: (empty)")
			return true
		}
		fmt.Printf("hand: %s\n", strings.Join(cards, " "))
	case "status":
		a.printStatus()
	case "quit", "exit":
		return false
	default:
		fmt.Println("unknown command, try: help")
	}

	return true
}

func (a *cliApp) connect(url string) error {
	if a.clientNC != nil {
		a.unsubscribeTable()
		a.clientNC.Close()
		a.clientNC = nil
		a.currentTable = ""
	}

	nc, err := nats.Connect(url)
	if err != nil {
		return err
	}

	a.clientNC = nc
	fmt.Printf("Connected to %s\n", nc.ConnectedUrl())
	return nil
}

func (a *cliApp) openTable(tableID, host string, port int) error {
	if tableID == "" {
		return fmt.Errorf("table id is required")
	}

	if a.host == nil {
		natsServer, err := server.NewServer(&server.Options{
			ServerName: "hearts-embedded",
			Host:       host,
			Port:       port,
			NoSigs:     true,
		})
		if err != nil {
			return fmt.Errorf("create embedded nats server: %w", err)
		}

		go natsServer.Start()
		if !natsServer.ReadyForConnections(10 * time.Second) {
			natsServer.Shutdown()
			return fmt.Errorf("embedded nats server did not become ready")
		}

		nc, err := nats.Connect(natsServer.ClientURL())
		if err != nil {
			natsServer.Shutdown()
			return fmt.Errorf("connect embedded nats server: %w", err)
		}

		a.host = &localHost{
			server: natsServer,
			nc:     nc,
			url:    natsServer.ClientURL(),
			host:   host,
			port:   port,
			tables: make(map[string]*table.Service),
		}

		fmt.Printf("Opened local game bus on %s\n", a.host.url)
	} else if a.host.host != host || a.host.port != port {
		return fmt.Errorf("local bus already running on %s", a.host.url)
	}

	if _, exists := a.host.tables[tableID]; !exists {
		svc := table.NewService(tableID, a.host.nc)
		if err := svc.Register(); err != nil {
			return fmt.Errorf("register table service: %w", err)
		}
		a.host.tables[tableID] = svc
		fmt.Printf("Opened table %s\n", tableID)
	}

	if err := a.connect(a.host.url); err != nil {
		return fmt.Errorf("connect local bus: %w", err)
	}

	return a.joinTable(tableID)
}

func (a *cliApp) discoverTables() ([]protocol.TableInfo, error) {
	if err := a.ensureClientConnection(); err != nil {
		return nil, err
	}

	inbox := nats.NewInbox()
	sub, err := a.clientNC.SubscribeSync(inbox)
	if err != nil {
		return nil, err
	}
	defer sub.Unsubscribe()

	payload, _ := json.Marshal(protocol.DiscoverRequest{})
	if err := a.clientNC.PublishRequest(protocol.DiscoverSubject(), inbox, payload); err != nil {
		return nil, err
	}
	if err := a.clientNC.Flush(); err != nil {
		return nil, err
	}

	deadline := time.Now().Add(750 * time.Millisecond)
	tables := make([]protocol.TableInfo, 0)
	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			break
		}

		msg, err := sub.NextMsg(remaining)
		if err == nats.ErrTimeout {
			break
		}
		if err != nil {
			return nil, err
		}

		var response protocol.DiscoverResponse
		if err := json.Unmarshal(msg.Data, &response); err != nil {
			continue
		}
		tables = append(tables, response.Tables...)
	}

	sort.Slice(tables, func(i, j int) bool {
		return tables[i].TableID < tables[j].TableID
	})

	return tables, nil
}

func (a *cliApp) joinTable(tableID string) error {
	if err := a.ensureClientConnection(); err != nil {
		return err
	}

	if err := a.subscribeTable(tableID); err != nil {
		return err
	}

	joinReq := protocol.JoinRequest{PlayerID: a.playerID, Name: a.name}
	joinPayload, _ := json.Marshal(joinReq)
	joinMsg, err := a.clientNC.Request(protocol.JoinSubject(tableID), joinPayload, 2*time.Second)
	if err != nil {
		a.unsubscribeTable()
		return err
	}

	var joinResp protocol.JoinResponse
	if err := json.Unmarshal(joinMsg.Data, &joinResp); err != nil {
		a.unsubscribeTable()
		return fmt.Errorf("invalid join response: %w", err)
	}
	if !joinResp.Accepted {
		a.unsubscribeTable()
		return fmt.Errorf("%s", joinResp.Reason)
	}

	a.currentTable = tableID
	fmt.Printf("Joined table %s as %s (seat %d)\n", joinResp.TableID, a.name, joinResp.Seat)
	return nil
}

func (a *cliApp) startRound() error {
	if err := a.ensureJoinedTable(); err != nil {
		return err
	}

	return sendCommand(a.clientNC, protocol.StartSubject(a.currentTable), protocol.StartRequest{PlayerID: a.playerID})
}

func (a *cliApp) playCard(card string) error {
	if err := a.ensureJoinedTable(); err != nil {
		return err
	}

	return sendCommand(a.clientNC, protocol.PlaySubject(a.currentTable), protocol.PlayCardRequest{PlayerID: a.playerID, Card: card})
}

func (a *cliApp) ensureClientConnection() error {
	if a.clientNC != nil {
		return nil
	}

	if a.host != nil {
		return a.connect(a.host.url)
	}

	if a.defaultURL == "" {
		return fmt.Errorf("not connected; use open or connect")
	}

	if err := a.connect(a.defaultURL); err != nil {
		return fmt.Errorf("not connected; use open or connect: %w", err)
	}

	return nil
}

func (a *cliApp) ensureJoinedTable() error {
	if err := a.ensureClientConnection(); err != nil {
		return err
	}
	if a.currentTable == "" {
		return fmt.Errorf("join a table first")
	}
	return nil
}

func (a *cliApp) subscribeTable(tableID string) error {
	a.unsubscribeTable()

	tableSub, err := a.clientNC.Subscribe(protocol.EventsSubject(tableID), func(msg *nats.Msg) {
		a.handleEvent(msg.Data)
	})
	if err != nil {
		return err
	}

	playerSub, err := a.clientNC.Subscribe(protocol.PlayerEventsSubject(tableID, a.playerID), func(msg *nats.Msg) {
		a.handleEvent(msg.Data)
	})
	if err != nil {
		tableSub.Unsubscribe()
		return err
	}

	if err := a.clientNC.Flush(); err != nil {
		tableSub.Unsubscribe()
		playerSub.Unsubscribe()
		return err
	}

	a.tableSub = tableSub
	a.playerSub = playerSub
	a.hand.update(nil)
	return nil
}

func (a *cliApp) unsubscribeTable() {
	if a.tableSub != nil {
		a.tableSub.Unsubscribe()
		a.tableSub = nil
	}
	if a.playerSub != nil {
		a.playerSub.Unsubscribe()
		a.playerSub = nil
	}
	a.currentTable = ""
	a.hand.update(nil)
}

func (a *cliApp) handleEvent(raw []byte) {
	var event protocol.Event
	if err := json.Unmarshal(raw, &event); err != nil {
		fmt.Printf("event decode error: %v\n", err)
		return
	}

	switch event.Type {
	case protocol.EventPlayerJoined:
		var data protocol.PlayerJoinedData
		if decodeEventPayload(event, &data) {
			fmt.Printf("[event] %s joined (seat %d)\n", data.Player.Name, data.Player.Seat)
		}
	case protocol.EventGameStarted:
		fmt.Println("[event] round started")
	case protocol.EventTurnChanged:
		var data protocol.TurnChangedData
		if decodeEventPayload(event, &data) {
			fmt.Printf("[event] turn: %s (trick %d)\n", data.PlayerID, data.TrickNumber+1)
		}
	case protocol.EventCardPlayed:
		var data protocol.CardPlayedData
		if decodeEventPayload(event, &data) {
			fmt.Printf("[event] %s played %s\n", data.PlayerID, data.Card)
		}
	case protocol.EventTrickCompleted:
		var data protocol.TrickCompletedData
		if decodeEventPayload(event, &data) {
			fmt.Printf("[event] trick %d won by %s (+%d points)\n", data.TrickNumber+1, data.WinnerPlayerID, data.Points)
		}
	case protocol.EventRoundCompleted:
		var data protocol.RoundCompletedData
		if decodeEventPayload(event, &data) {
			fmt.Printf("[event] round over, round points=%v totals=%v\n", data.RoundPoints, data.TotalPoints)
		}
	case protocol.EventHandUpdated:
		var data protocol.HandUpdatedData
		if decodeEventPayload(event, &data) {
			a.hand.update(data.Cards)
			fmt.Printf("[event] hand updated: %s\n", strings.Join(a.hand.snapshot(), " "))
		}
	}
}

func (a *cliApp) printStatus() {
	connected := "no"
	url := ""
	if a.clientNC != nil {
		connected = "yes"
		url = a.clientNC.ConnectedUrl()
	}

	fmt.Printf("player: %s (%s)\n", a.name, a.playerID)
	fmt.Printf("connected: %s", connected)
	if url != "" {
		fmt.Printf(" (%s)", url)
	}
	fmt.Println()

	if a.currentTable == "" {
		fmt.Println("table: none")
	} else {
		fmt.Printf("table: %s\n", a.currentTable)
	}

	if a.host != nil {
		fmt.Printf("local bus: %s\n", a.host.url)
	}
}

func (a *cliApp) shutdown() {
	a.unsubscribeTable()

	if a.clientNC != nil {
		a.clientNC.Close()
		a.clientNC = nil
	}

	if a.host != nil {
		a.host.nc.Close()
		a.host.server.Shutdown()
		a.host = nil
	}
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

func sendCommand[T any](nc *nats.Conn, subject string, req T) error {
	payload, _ := json.Marshal(req)
	msg, err := nc.Request(subject, payload, 2*time.Second)
	if err != nil {
		return err
	}

	var resp protocol.CommandResponse
	if err := json.Unmarshal(msg.Data, &resp); err != nil {
		return fmt.Errorf("invalid response: %w", err)
	}
	if !resp.Accepted {
		return fmt.Errorf("%s", resp.Reason)
	}

	return nil
}

func decodeEventPayload[T any](event protocol.Event, out *T) bool {
	if err := json.Unmarshal(event.Data, out); err != nil {
		fmt.Printf("event payload decode error (%s): %v\n", event.Type, err)
		return false
	}
	return true
}

func randomHex(bytes int) string {
	raw := make([]byte, bytes)
	if _, err := rand.Read(raw); err != nil {
		return "0000"
	}
	return hex.EncodeToString(raw)
}

func printHelp() {
	fmt.Println("commands:")
	fmt.Println("  open [table-id] [port]   open local game and join it")
	fmt.Println("  discover                 discover open tables on current bus")
	fmt.Println("  join <table-id>          join discovered table")
	fmt.Println("  connect <nats-url>       switch to another game bus")
	fmt.Println("  start                    start the round (needs 4 players)")
	fmt.Println("  play <card>              play a card, e.g. play QS")
	fmt.Println("  hand                     show your hand")
	fmt.Println("  status                   show current connection and table")
	fmt.Println("  help                     show this help")
	fmt.Println("  quit                     exit")
}
