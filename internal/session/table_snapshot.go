package session

import (
	"sort"

	"github.com/JHK/hearts/internal/game"
	"github.com/JHK/hearts/internal/protocol"
)

func (r *Table) buildBotHands(state *tableState) []BotHandSnapshot {
	var out []BotHandSnapshot
	for _, p := range state.players {
		if p.bot == nil {
			continue
		}
		var hand []game.Card
		if state.round != nil {
			hand = state.round.Hand(p.position)
		}
		cards := make([]string, 0, len(hand))
		for _, c := range hand {
			cards = append(cards, c.String())
		}
		out = append(out, BotHandSnapshot{Name: p.Name, Seat: p.position, Cards: cards})
	}
	return out
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
