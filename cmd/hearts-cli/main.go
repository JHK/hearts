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
	"strings"
	"sync"
	"time"

	"github.com/JHK/hearts/internal/protocol"
	"github.com/nats-io/nats.go"
)

func main() {
	var (
		natsURL  = flag.String("url", "nats://127.0.0.1:4222", "NATS server URL")
		tableID  = flag.String("table", "default", "table id")
		playerID = flag.String("player-id", "", "player id")
		name     = flag.String("name", "", "display name")
	)
	flag.Parse()

	if *playerID == "" {
		*playerID = "p-" + randomHex(4)
	}
	if *name == "" {
		*name = *playerID
	}

	nc, err := nats.Connect(*natsURL)
	if err != nil {
		fatalf("failed to connect to nats: %v", err)
	}
	defer nc.Close()

	hand := &handState{}

	if _, err := nc.Subscribe(protocol.EventsSubject(*tableID), func(msg *nats.Msg) {
		handleEvent(msg.Data, hand)
	}); err != nil {
		fatalf("failed to subscribe to table events: %v", err)
	}

	if _, err := nc.Subscribe(protocol.PlayerEventsSubject(*tableID, *playerID), func(msg *nats.Msg) {
		handleEvent(msg.Data, hand)
	}); err != nil {
		fatalf("failed to subscribe to player events: %v", err)
	}

	joinReq := protocol.JoinRequest{PlayerID: *playerID, Name: *name}
	joinPayload, _ := json.Marshal(joinReq)
	joinMsg, err := nc.Request(protocol.JoinSubject(*tableID), joinPayload, 2*time.Second)
	if err != nil {
		fatalf("join request failed: %v", err)
	}

	var joinResp protocol.JoinResponse
	if err := json.Unmarshal(joinMsg.Data, &joinResp); err != nil {
		fatalf("invalid join response: %v", err)
	}
	if !joinResp.Accepted {
		fatalf("join rejected: %s", joinResp.Reason)
	}

	fmt.Printf("Joined table %s as %s (seat %d)\n", joinResp.TableID, *name, joinResp.Seat)
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

		parts := strings.Fields(line)
		switch strings.ToLower(parts[0]) {
		case "help":
			printHelp()
		case "start":
			if err := sendCommand(nc, protocol.StartSubject(*tableID), protocol.StartRequest{PlayerID: *playerID}); err != nil {
				fmt.Printf("start rejected: %v\n", err)
			}
		case "play":
			if len(parts) != 2 {
				fmt.Println("usage: play <card>, example: play QS")
				continue
			}
			card := strings.ToUpper(parts[1])
			if err := sendCommand(nc, protocol.PlaySubject(*tableID), protocol.PlayCardRequest{PlayerID: *playerID, Card: card}); err != nil {
				fmt.Printf("play rejected: %v\n", err)
			}
		case "hand":
			fmt.Printf("hand: %s\n", strings.Join(hand.snapshot(), " "))
		case "quit", "exit":
			return
		default:
			fmt.Println("unknown command, try: help")
		}
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

func handleEvent(raw []byte, hand *handState) {
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
			hand.update(data.Cards)
			fmt.Printf("[event] hand updated: %s\n", strings.Join(hand.snapshot(), " "))
		}
	}
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
	fmt.Println("  start         start the round (needs 4 players)")
	fmt.Println("  play <card>   play a card, e.g. play QS")
	fmt.Println("  hand          show your hand")
	fmt.Println("  help          show this help")
	fmt.Println("  quit          leave client")
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
