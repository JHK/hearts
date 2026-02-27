package app

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/JHK/hearts/internal/game"
	"github.com/JHK/hearts/internal/protocol"
	natswire "github.com/JHK/hearts/internal/transport/nats"
)

func Run() {
	var (
		defaultURL = flag.String("url", "nats://127.0.0.1:4222", "default NATS URL for discover/join")
		name       = flag.String("name", "", "display name")
	)
	flag.Parse()

	if *name == "" {
		*name = "Player"
	}

	app := &cliApp{
		session: NewSession(*name, *defaultURL),
		name:    *name,
	}
	defer app.session.Shutdown()

	fmt.Printf("hearts cli - player %s\n", app.name)
	if err := app.session.Connect(*defaultURL); err != nil {
		fmt.Printf("No game bus at %s yet. Use 'open' to host or 'connect <url>' to join another bus.\n", *defaultURL)
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
	session *Session
	name    string
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
		if err := a.session.Connect(parts[1]); err != nil {
			fmt.Printf("connect failed: %v\n", err)
			return true
		}
		status := a.session.Status()
		fmt.Printf("Connected to %s\n", status.ConnectedTo)
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

		joinResp, err := a.session.OpenTable(tableID, "127.0.0.1", port, a.eventHandlers())
		if err != nil {
			fmt.Printf("open failed: %v\n", err)
			return true
		}

		status := a.session.Status()
		if status.LocalBusURL != "" {
			fmt.Printf("Opened local game bus on %s\n", status.LocalBusURL)
		}
		fmt.Printf("Joined table %s as %s (seat %d)\n", joinResp.TableID, a.name, joinResp.Seat)
	case "discover":
		tables, err := a.session.DiscoverTables()
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
		joinResp, err := a.session.JoinTable(parts[1], a.eventHandlers())
		if err != nil {
			fmt.Printf("join failed: %v\n", err)
			return true
		}
		fmt.Printf("Joined table %s as %s (seat %d)\n", joinResp.TableID, a.name, joinResp.Seat)
	case "addbot":
		joinResp, err := a.session.AddBot()
		if err != nil {
			fmt.Printf("addbot failed: %v\n", err)
			return true
		}
		fmt.Printf("Added bot %s at seat %d\n", joinResp.PlayerID, joinResp.Seat)
	case "start":
		if err := a.session.StartRound(); err != nil {
			fmt.Printf("start rejected: %v\n", err)
		}
	case "play":
		if len(parts) != 2 {
			fmt.Println("usage: play <card>, example: play QS")
			return true
		}
		if err := a.session.PlayCard(strings.ToUpper(parts[1])); err != nil {
			fmt.Printf("play rejected: %v\n", err)
		}
	case "hand":
		cards := a.session.HandSnapshot()
		if len(cards) == 0 {
			fmt.Println("hand: (empty)")
			return true
		}
		fmt.Printf("hand: %s\n", strings.Join(cards, " "))
	case "stats":
		table, err := a.session.TableStats()
		if err != nil {
			fmt.Printf("stats failed: %v\n", err)
			return true
		}
		state := "waiting"
		if table.Started {
			state = "in_round"
		}
		fmt.Printf("table %s: players=%d/%d state=%s\n", table.TableID, table.Players, table.MaxPlayers, state)
	case "status":
		a.printStatus()
	case "quit", "exit":
		return false
	default:
		fmt.Println("unknown command, try: help")
	}

	return true
}

func (a *cliApp) eventHandlers() natswire.ParticipantEventHandlers {
	return natswire.ParticipantEventHandlers{
		OnPlayerJoined: a.onPlayerJoined,
		OnGameStarted: func() {
			fmt.Println("[event] round started")
		},
		OnTurnChanged: a.onTurnChanged,
		OnCardPlayed:  a.onCardPlayed,
		OnTrickCompleted: func(data protocol.TrickCompletedData) {
			fmt.Printf("[event] trick %d won by %s (+%d points)\n", data.TrickNumber+1, data.WinnerPlayerID, data.Points)
		},
		OnRoundCompleted: func(data protocol.RoundCompletedData) {
			fmt.Printf("[event] round over, round points=%v totals=%v\n", data.RoundPoints, data.TotalPoints)
		},
		OnHandUpdated: func(_ game.PlayerID, _ protocol.HandUpdatedData) {
			fmt.Printf("[event] hand updated: %s\n", strings.Join(a.session.HandSnapshot(), " "))
		},
		OnYourTurn: func(_ game.PlayerID, data protocol.YourTurnData) {
			fmt.Printf("[event] your turn (trick %d)\n", data.TrickNumber+1)
		},
		OnDecodeError: func(err error) {
			fmt.Printf("event decode error: %v\n", err)
		},
	}
}

func (a *cliApp) onPlayerJoined(data protocol.PlayerJoinedData) {
	fmt.Printf("[event] %s joined (seat %d)\n", data.Player.Name, data.Player.Seat)
}

func (a *cliApp) onTurnChanged(data protocol.TurnChangedData) {
	fmt.Printf("[event] turn: %s (trick %d)\n", data.PlayerID, data.TrickNumber+1)
}

func (a *cliApp) onCardPlayed(data protocol.CardPlayedData) {
	fmt.Printf("[event] %s played %s\n", data.PlayerID, data.Card)
}

func (a *cliApp) printStatus() {
	status := a.session.Status()

	if status.PlayerID == "" {
		fmt.Printf("player: %s\n", status.PlayerName)
	} else {
		fmt.Printf("player: %s (%s)\n", status.PlayerName, status.PlayerID)
	}
	fmt.Printf("connected: %s", boolLabel(status.Connected))
	if status.ConnectedTo != "" {
		fmt.Printf(" (%s)", status.ConnectedTo)
	}
	fmt.Println()

	if status.TableID == "" {
		fmt.Println("table: none")
	} else {
		fmt.Printf("table: %s\n", status.TableID)
	}

	if status.LocalBusURL != "" {
		fmt.Printf("local bus: %s\n", status.LocalBusURL)
	}
}

func boolLabel(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

func printHelp() {
	fmt.Println("commands:")
	fmt.Println("  open [table-id] [port]   open local game and join it")
	fmt.Println("  discover                 discover open tables on current bus")
	fmt.Println("  join <table-id>          join discovered table")
	fmt.Println("  connect <nats-url>       switch to another game bus")
	fmt.Println("  addbot                   add one bot seat to current table")
	fmt.Println("  start                    start round (requires 4 players)")
	fmt.Println("  play <card>              play a card, e.g. play QS")
	fmt.Println("  hand                     show your hand")
	fmt.Println("  stats                    show current table stats")
	fmt.Println("  status                   show current connection and table")
	fmt.Println("  help                     show this help")
	fmt.Println("  quit                     exit")
}
