package app

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/JHK/hearts/internal/game"
	"github.com/JHK/hearts/internal/protocol"
	natswire "github.com/JHK/hearts/internal/transport/nats"
	"github.com/alecthomas/kong"
)

type commandLine struct {
	CLI  cliCommand  `cmd:"" help:"Run interactive terminal CLI."`
	Host hostCommand `cmd:"" help:"Start a local host for a table."`
	Web  webCommand  `cmd:"" help:"Start a web UI to play a table."`
}

type cliCommand struct {
	Name   string `help:"Display name." default:"Player"`
	Server string `help:"Default server for discover/join." default:"nats://127.0.0.1:4222"`
}

type hostCommand struct {
	TableID string `name:"table" help:"Table ID to host." default:"default"`
	Host    string `help:"Host interface for embedded NATS." default:"127.0.0.1"`
	Port    int    `help:"Port for embedded NATS." default:"4222"`
}

type webCommand struct {
	Name        string `help:"Display name." default:"Player"`
	Server      string `help:"Server for reconnect/join." default:"nats://127.0.0.1:4222"`
	TableID     string `name:"table" help:"Table ID to join." default:"default"`
	Addr        string `help:"Web listen address." default:"127.0.0.1:8080"`
	OpenBrowser bool   `help:"Open browser after server starts." default:"true"`
}

func Run() {
	args := commandLine{}
	parser, err := kong.New(
		&args,
		kong.Name("hearts"),
		kong.Description("Play Hearts via terminal, host mode, or simple web UI."),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse setup failed: %v\n", err)
		os.Exit(1)
	}

	argv := os.Args[1:]
	if len(argv) == 0 {
		argv = append([]string{"cli"}, argv...)
	} else if strings.HasPrefix(argv[0], "-") && argv[0] != "--help" && argv[0] != "-h" {
		argv = append([]string{"cli"}, argv...)
	}

	ctx, err := parser.Parse(argv)
	parser.FatalIfErrorf(err)

	switch ctx.Command() {
	case "cli":
		runInteractiveCLI(args.CLI.Name, args.CLI.Server)
	case "host":
		runHostMode(args.Host)
	case "web":
		runWebMode(args.Web)
	}
}

func runInteractiveCLI(name, defaultServer string) {
	if strings.TrimSpace(name) == "" {
		name = "Player"
	}

	app := &cliApp{
		session: NewSession(name, defaultServer),
		name:    name,
	}
	defer app.session.Shutdown()

	fmt.Printf("hearts cli - player %s\n", app.name)
	if err := app.session.Connect(defaultServer); err != nil {
		fmt.Printf("No game bus at %s yet. Use 'open' to host or 'connect <server>' to join another bus.\n", defaultServer)
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

func runHostMode(cfg hostCommand) {
	session := NewSession("host", "")
	defer session.Shutdown()

	if err := session.HostTable(cfg.TableID, cfg.Host, cfg.Port); err != nil {
		fmt.Printf("host failed: %v\n", err)
		return
	}

	status := session.Status()
	fmt.Printf("Hosting table %s on %s\n", cfg.TableID, status.LocalBusURL)
	fmt.Println("Press Ctrl+C to stop.")
	waitForInterrupt()
}

func waitForInterrupt() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
	signal.Stop(sigCh)
}

func (a *cliApp) runCommand(parts []string) bool {
	switch strings.ToLower(parts[0]) {
	case "help":
		printHelp()
	case "connect":
		if len(parts) != 2 {
			fmt.Println("usage: connect <server>")
			return true
		}
		if err := a.session.Connect(parts[1]); err != nil {
			fmt.Printf("connect failed: %v\n", err)
			return true
		}
		status := a.session.Status()
		fmt.Printf("Connected to %s\n", status.Server)
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
		if len(parts) > 2 {
			fmt.Println("usage: join [table-id]")
			return true
		}
		tableID := "default"
		if len(parts) == 2 {
			tableID = parts[1]
		}
		joinResp, err := a.session.JoinTable(tableID, a.eventHandlers())
		if err != nil {
			fmt.Printf("join failed: %v\n", err)
			return true
		}
		fmt.Printf("Joined table %s as %s (seat %d)\n", joinResp.TableID, a.name, joinResp.Seat)
	case "addbot":
		if len(parts) > 2 {
			fmt.Println("usage: addbot [strategy]")
			return true
		}

		strategyName := ""
		if len(parts) >= 2 {
			strategyName = parts[1]
		}

		added, err := a.session.AddBot(strategyName)
		if err != nil {
			fmt.Printf("addbot failed: %v\n", err)
			return true
		}
		fmt.Printf("Added bot %s (%s, strategy=%s) at seat %d\n", added.JoinResponse.PlayerID, added.Name, added.Strategy, added.JoinResponse.Seat)
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
	if status.Server != "" {
		fmt.Printf(" (%s)", status.Server)
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
	fmt.Println("  join [table-id]          join table (default: default)")
	fmt.Println("  connect <server>         switch to another game bus")
	fmt.Println("  addbot [strategy]        add one bot seat (default: random)")
	fmt.Println("  start                    start round (requires 4 players)")
	fmt.Println("  play <card>              play a card, e.g. play QS")
	fmt.Println("  hand                     show your hand")
	fmt.Println("  stats                    show current table stats")
	fmt.Println("  status                   show current connection and table")
	fmt.Println("  help                     show this help")
	fmt.Println("  quit                     exit")
}
