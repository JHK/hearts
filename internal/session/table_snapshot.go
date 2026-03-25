package session

import (
	"sort"
	"strconv"
	"strings"

	"github.com/JHK/hearts/internal/game"
	"github.com/JHK/hearts/internal/game/bot"
	"github.com/JHK/hearts/internal/protocol"
)

func (r *Table) buildDebugBotContext(state *tableState) *DebugBotSnapshot {
	snap := &DebugBotSnapshot{
		TableID: r.tableID,
	}

	// Collect bot snapshots (hand + strategy + moon-shot state).
	for _, p := range state.players {
		if p.bot == nil {
			continue
		}
		var hand []game.Card
		if state.round != nil {
			hand = state.round.Hand(p.position)
		}
		bs := BotSnapshot{
			Name:     p.Name,
			Seat:     p.position,
			Strategy: string(p.bot.Kind()),
			Hand:     game.CardStrings(hand),
		}
		if smart, ok := p.bot.(*bot.Smart); ok {
			active := smart.MoonShotActive()
			aborted := smart.MoonShotAborted()
			bs.MoonShotActive = &active
			bs.MoonShotAborted = &aborted
		}
		snap.Bots = append(snap.Bots, bs)
	}

	if state.round == nil {
		snap.Phase = "waiting"
		return snap
	}

	rd := state.round
	snap.Phase = roundPhaseString(rd.Phase())
	snap.TrickNumber = rd.TrickNumber()
	snap.HeartsBroken = rd.HeartsBroken()
	snap.FirstTrick = rd.TrickNumber() == 0
	snap.PassDirection = rd.PassDirection()

	// Current trick plays.
	trick := rd.CurrentTrick()
	snap.CurrentTrick = make([]TrickPlaySnapshot, 0, len(trick))
	for _, played := range trick {
		p := state.players[played.Seat]
		snap.CurrentTrick = append(snap.CurrentTrick, TrickPlaySnapshot{
			PlayerID: p.id,
			Name:     p.Name,
			Seat:     played.Seat,
			Card:     played.Card.String(),
		})
	}
	if len(trick) > 0 {
		snap.LedSuit = trick[0].Card.Suit
	}

	// Previously played cards.
	played := rd.PlayedCards()
	snap.PlayedCards = make([]string, 0, len(played))
	for _, c := range played {
		snap.PlayedCards = append(snap.PlayedCards, c.String())
	}

	// Turn info.
	if rd.Phase() == game.PhasePlaying {
		seat := rd.TurnSeat()
		snap.TurnSeat = &seat
		snap.TurnPlayer = state.players[seat].Name
	}

	// All players (ordered by seat) for the scores table.
	snap.Players = make([]string, 0, len(state.players))
	for _, p := range state.players {
		snap.Players = append(snap.Players, p.Name)
	}

	// Scores keyed by player name for readability.
	snap.RoundPoints = make(map[string]game.Points, len(state.players))
	snap.TotalPoints = make(map[string]game.Points, len(state.players))
	gameScores := state.game.Scores()
	for i, p := range state.players {
		snap.RoundPoints[p.Name] = rd.RoundPoints(i)
		snap.TotalPoints[p.Name] = gameScores[p.position]
	}

	return snap
}

// FormatMarkdown renders the debug snapshot as a markdown text block
// suitable for pasting into a Claude conversation.
func (s *DebugBotSnapshot) FormatMarkdown() string {
	var b strings.Builder
	b.WriteString("# Bot Decision Context\n\n")
	b.WriteString("## Table State\n")
	b.WriteString("- **Table:** " + s.TableID + "\n")
	b.WriteString("- **Phase:** " + s.Phase + "\n")
	if s.PassDirection != "" {
		b.WriteString("- **Pass direction:** " + string(s.PassDirection) + "\n")
	}
	b.WriteString("- **Trick:** " + itoa(s.TrickNumber+1) + " of 13\n")
	b.WriteString("- **Hearts broken:** " + boolStr(s.HeartsBroken) + "\n")
	b.WriteString("- **First trick:** " + boolStr(s.FirstTrick) + "\n")
	if s.LedSuit != "" {
		b.WriteString("- **Led suit:** " + string(s.LedSuit) + "\n")
	}
	if s.TurnSeat != nil {
		b.WriteString("- **Turn:** " + s.TurnPlayer + " (seat " + itoa(*s.TurnSeat) + ")\n")
	}

	if len(s.CurrentTrick) > 0 {
		b.WriteString("\n## Current Trick\n")
		b.WriteString("| Seat | Player | Card |\n|------|--------|------|\n")
		for _, tp := range s.CurrentTrick {
			b.WriteString("| " + itoa(tp.Seat) + " | " + tp.Name + " | " + tp.Card + " |\n")
		}
	}

	if len(s.PlayedCards) > 0 {
		b.WriteString("\n## Previously Played Cards\n")
		b.WriteString(strings.Join(s.PlayedCards, " ") + "\n")
	}

	b.WriteString("\n## Scores\n")
	b.WriteString("| Player | Round | Total |\n|--------|-------|-------|\n")
	for _, name := range s.Players {
		rp := s.RoundPoints[name]
		tp := s.TotalPoints[name]
		b.WriteString("| " + name + " | " + itoa(int(rp)) + " | " + itoa(int(tp)) + " |\n")
	}

	b.WriteString("\n## Bots\n")
	for _, bot := range s.Bots {
		b.WriteString("\n### " + bot.Name + " (seat " + itoa(bot.Seat) + ", " + bot.Strategy + ")\n")
		if len(bot.Hand) > 0 {
			b.WriteString("**Hand:** " + strings.Join(bot.Hand, " ") + "\n")
		} else {
			b.WriteString("**Hand:** (none)\n")
		}
		if bot.MoonShotActive != nil && bot.MoonShotAborted != nil {
			b.WriteString("**Moon shot active:** " + boolStr(*bot.MoonShotActive) + "\n")
			b.WriteString("**Moon shot aborted:** " + boolStr(*bot.MoonShotAborted) + "\n")
		}
	}

	return b.String()
}

func boolStr(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

func itoa(n int) string {
	return strconv.Itoa(n)
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

	totals := r.buildTotals(state)

	snapshot := Snapshot{
		TableID:           r.tableID,
		Players:           playerSnapshots,
		Started:           state.round != nil,
		Phase:             "",
		HandSizes:         map[protocol.PlayerID]int{},
		RoundPoints:       map[protocol.PlayerID]game.Points{},
		RoundHistory:      copyRoundHistory(state.roundHistory),
		TotalPoints:       totals,
		GameOver:          state.gameOver,
		Paused:            state.paused,
		PausedForPlayerID: state.pausedPlayerID,
	}

	if state.gameOver {
		snapshot.Winners = r.seatWinnersToPlayerIDs(state, state.game.Winners())
		humans := 0
		for _, p := range state.players {
			if p.bot == nil {
				humans++
			}
		}
		snapshot.RematchVotes = len(state.rematchVotes)
		snapshot.RematchTotal = humans
		if forPlayer != "" && state.rematchVotes[forPlayer] {
			snapshot.RematchVoted = true
		}
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
