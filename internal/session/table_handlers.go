package session

import (
	"fmt"
	"log/slog"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/JHK/hearts/internal/game"
	"github.com/JHK/hearts/internal/game/bot"
	"github.com/JHK/hearts/internal/protocol"
)

func (r *Table) handleLeave(state *tableState, playerID protocol.PlayerID) {
	player := state.playersByID[playerID]
	if player == nil {
		return
	}

	slog.Info("player left table", "event", "player_left", "table_id", r.tableID, "player_id", playerID, "name", player.Name)

	if state.round != nil {
		if player.bot != nil {
			return
		}

		// Convert the departing human to a bot and pause the game so remaining
		// humans can decide whether to wait for reconnection or continue.
		player.bot = bot.StrategyRandom.NewBot()
		state.paused = true
		state.pausedPlayerID = player.id
		slog.Info("game paused due to disconnection", "event", "game_paused", "table_id", r.tableID, "player_id", playerID, "name", player.Name)
		r.publishPublic(protocol.EventGamePaused, protocol.GamePausedData{
			Player: protocol.PlayerInfo{PlayerID: player.id, Name: player.Name, Seat: player.position},
		})

		// Immediately resolve any pending action for the departing player's
		// new bot so the game state is consistent when it resumes.
		switch state.round.Phase() {
		case game.PhasePassing:
			if !state.round.HasSubmittedPass(player.position) {
				input := state.round.PassInput(player.position, state.game.Scores())
				cards, err := player.bot.ChoosePass(input)
				if err == nil {
					_ = state.round.SubmitPass(player.position, cards)
				}
			}
		case game.PhasePassReview:
			if !state.round.IsPassReady(player.position) {
				_ = state.round.MarkReady(player.position)
			}
		}
		return
	}

	seat := player.position
	delete(state.playersByID, playerID)
	delete(state.rematchVotes, playerID)
	if player.Token != "" {
		delete(state.playersByToken, player.Token)
		state.departedTokens[player.Token] = playerID
	}

	for index, seated := range state.players {
		if seated.id != playerID {
			continue
		}

		state.players = append(state.players[:index], state.players[index+1:]...)
		for i, updated := range state.players {
			updated.position = i
		}
		break
	}

	r.publishPublic(protocol.EventPlayerLeft, protocol.PlayerLeftData{
		Player: protocol.PlayerInfo{PlayerID: playerID, Name: player.Name, Seat: seat},
	})

	if state.gameOver {
		state.gameOver = false
		state.game = game.NewGame()
		state.rematchVotes = nil
		state.roundHistory = nil
	}
}

func (r *Table) handleJoin(state *tableState, name, token string) protocol.JoinResponse {
	token = strings.TrimSpace(token)
	if token == "" {
		return protocol.JoinResponse{Accepted: false, Reason: "player token is required"}
	}

	if existing := state.playersByToken[token]; existing != nil {
		if existing.bot != nil {
			existing.bot = nil
			if n := strings.TrimSpace(name); n != "" {
				existing.Name = n
			}
			slog.Info("player reclaimed seat", "event", "player_reclaimed", "table_id", r.tableID, "player_id", existing.id, "name", existing.Name, "seat", existing.position)
			if state.paused {
				state.paused = false
				state.pausedPlayerID = ""
				r.publishGameResumed()
				r.resumeAfterPause(state)
			}
		}
		return protocol.JoinResponse{
			Accepted: true,
			TableID:  r.tableID,
			PlayerID: existing.id,
			Seat:     existing.position,
		}
	}

	if state.round != nil {
		return protocol.JoinResponse{Accepted: false, Reason: ErrRoundInProgress.Error()}
	}
	if len(state.players) >= game.PlayersPerTable {
		return protocol.JoinResponse{Accepted: false, Reason: ErrTableFull.Error()}
	}

	name = strings.TrimSpace(name)
	if name == "" {
		name = "Player"
	}

	id, reused := state.departedTokens[token]
	if reused {
		delete(state.departedTokens, token)
	} else {
		id = r.nextPlayerID(state)
	}
	player := r.addPlayer(state, id, name, nil, token)

	slog.Info("player joined table", "event", "player_joined", "table_id", r.tableID, "player_id", player.id, "name", player.Name, "seat", player.position)

	r.publishPublic(protocol.EventPlayerJoined, protocol.PlayerJoinedData{Player: protocol.PlayerInfo{
		PlayerID: player.id,
		Name:     player.Name,
		Seat:     player.position,
	}})

	return protocol.JoinResponse{
		Accepted: true,
		TableID:  r.tableID,
		PlayerID: player.id,
		Seat:     player.position,
	}
}

func (r *Table) handleAddBot(state *tableState, strategyName string) (AddedBot, error) {
	if state.gameOver {
		return AddedBot{}, ErrGameOver
	}
	if state.round != nil {
		return AddedBot{}, ErrRoundInProgress
	}
	if len(state.players) >= game.PlayersPerTable {
		return AddedBot{}, ErrTableFull
	}

	strategyKind, err := bot.ParseStrategyKind(strategyName)
	if err != nil {
		return AddedBot{}, err
	}

	taken := make(map[string]bool, len(state.players))
	for _, p := range state.players {
		taken[p.Name] = true
	}

	id := r.nextPlayerID(state)
	player := r.addPlayer(state, id, strategyKind.BotName(taken), strategyKind.NewBot(), "")

	slog.Debug("bot added to table", "event", "bot_added", "table_id", r.tableID, "player_id", player.id, "name", player.Name, "strategy", string(strategyKind))

	r.publishPublic(protocol.EventPlayerJoined, protocol.PlayerJoinedData{Player: protocol.PlayerInfo{
		PlayerID: player.id,
		Name:     player.Name,
		Seat:     player.position,
	}})

	return AddedBot{
		JoinResponse: protocol.JoinResponse{
			Accepted: true,
			TableID:  r.tableID,
			PlayerID: player.id,
			Seat:     player.position,
		},
		Name:     player.Name,
		Strategy: string(strategyKind),
	}, nil
}

func (r *Table) addPlayer(state *tableState, id protocol.PlayerID, name string, b bot.Bot, token string) *playerState {
	player := &playerState{
		id:       id,
		Name:     name,
		position: len(state.players),
		Token:    token,
		bot:      b,
	}

	state.players = append(state.players, player)
	state.playersByID[id] = player
	if token != "" {
		state.playersByToken[token] = player
	}

	return player
}

func (r *Table) handleStart(state *tableState, playerID protocol.PlayerID) protocol.CommandResponse {
	if reason := r.validateStartPreconditions(state, playerID); reason != "" {
		return protocol.CommandResponse{Accepted: false, Reason: reason}
	}

	state.round = r.initializeRound(state)
	slog.Info("table started", "event", "table_started", "table_id", r.tableID, "round", state.game.RoundsPlayed()+1)

	hands := make(map[protocol.PlayerID][]string, len(state.players))
	for _, player := range state.players {
		hands[player.id] = game.CardStrings(state.round.Hand(player.position))
	}
	r.publishRoundStart(state, hands)

	if state.round.PassDirection() == game.PassDirectionHold {
		_ = state.round.StartPlaying()
		r.publishPlayPhaseStart(state)
		return protocol.CommandResponse{Accepted: true}
	}

	r.schedulePassingBots(state)

	return protocol.CommandResponse{Accepted: true}
}

func (r *Table) nextPlayerID(state *tableState) protocol.PlayerID {
	for {
		state.nextPlayerSeq++
		candidate := protocol.PlayerID(fmt.Sprintf("p-%d", state.nextPlayerSeq))
		if _, exists := state.playersByID[candidate]; !exists {
			return candidate
		}
	}
}

func (r *Table) handlePlay(state *tableState, playerID protocol.PlayerID, cardRaw string) protocol.CommandResponse {
	if reason := validateRoundCommandPreconditions(state, playerID); reason != "" {
		return protocol.CommandResponse{Accepted: false, Reason: reason}
	}

	card, err := game.ParseCard(cardRaw)
	if err != nil {
		return protocol.CommandResponse{Accepted: false, Reason: err.Error()}
	}

	player := state.playersByID[playerID]

	heartsBrokenBefore := state.round.HeartsBroken()
	trickResult, err := state.round.Play(player.position, card)
	if err != nil {
		return protocol.CommandResponse{Accepted: false, Reason: err.Error()}
	}

	breaksHearts := card.Suit == game.SuitHearts && !heartsBrokenBefore
	r.publishPublic(protocol.EventCardPlayed, protocol.CardPlayedData{
		PlayerID: playerID, Card: card.String(), BreaksHearts: breaksHearts,
	})
	r.publishPrivate(playerID, protocol.EventHandUpdated, protocol.HandUpdatedData{
		Cards: game.CardStrings(state.round.Hand(player.position)),
	})

	if trickResult == nil {
		// Trick not complete — advance to next player.
		nextSeat := state.round.TurnSeat()
		nextPlayer := state.players[nextSeat]
		r.publishTurn(nextPlayer.id, state.round.TrickNumber())
		r.scheduleBotTurn(state, nextPlayer)
		return protocol.CommandResponse{Accepted: true}
	}

	// Trick completed.
	winnerPlayer := state.players[trickResult.WinnerSeat]
	trickCompleted := protocol.TrickCompletedData{
		TrickNumber:    trickResult.TrickNumber,
		WinnerPlayerID: winnerPlayer.id,
		Points:         trickResult.Points,
	}

	if state.round.Phase() == game.PhaseComplete {
		// Last trick — round is complete.
		roundCompleted := r.completeRound(state)
		r.publishPublic(protocol.EventTrickCompleted, trickCompleted)
		r.publishPublic(protocol.EventRoundCompleted, roundCompleted)
		state.round = nil
		r.maybeEndGame(state)
		return protocol.CommandResponse{Accepted: true}
	}

	// More tricks to play.
	r.publishPublic(protocol.EventTrickCompleted, trickCompleted)
	r.publishTurn(winnerPlayer.id, state.round.TrickNumber())
	r.scheduleBotTurn(state, winnerPlayer)

	return protocol.CommandResponse{Accepted: true}
}

func (r *Table) handlePass(state *tableState, playerID protocol.PlayerID, cardsRaw []string) protocol.CommandResponse {
	if reason := validateRoundCommandPreconditions(state, playerID); reason != "" {
		return protocol.CommandResponse{Accepted: false, Reason: reason}
	}
	player := state.playersByID[playerID]

	passCards, err := r.parseAndValidatePassCards(cardsRaw)
	if err != nil {
		return protocol.CommandResponse{Accepted: false, Reason: err.Error()}
	}

	if err := state.round.SubmitPass(player.position, passCards); err != nil {
		return protocol.CommandResponse{Accepted: false, Reason: err.Error()}
	}

	submitted := 0
	for i := 0; i < game.PlayersPerTable; i++ {
		if state.round.HasSubmittedPass(i) {
			submitted++
		}
	}

	r.publishPublic(protocol.EventPassSubmitted, protocol.PassStatusData{
		Submitted: submitted,
		Total:     game.PlayersPerTable,
		Direction: state.round.PassDirection(),
	})

	if submitted < game.PlayersPerTable {
		return protocol.CommandResponse{Accepted: true}
	}

	_ = state.round.ApplyPasses()
	r.startPassReview(state)

	return protocol.CommandResponse{Accepted: true}
}

func (r *Table) handleReadyAfterPass(state *tableState, playerID protocol.PlayerID) protocol.CommandResponse {
	if reason := validateRoundCommandPreconditions(state, playerID); reason != "" {
		return protocol.CommandResponse{Accepted: false, Reason: reason}
	}

	if err := state.round.MarkReady(state.playersByID[playerID].position); err != nil {
		return protocol.CommandResponse{Accepted: false, Reason: err.Error()}
	}

	ready := 0
	for i := 0; i < game.PlayersPerTable; i++ {
		if state.round.IsPassReady(i) {
			ready++
		}
	}

	r.publishPublic(protocol.EventPassReadyChanged, protocol.PassReadyData{
		Ready: ready,
		Total: game.PlayersPerTable,
	})

	if ready < game.PlayersPerTable {
		return protocol.CommandResponse{Accepted: true}
	}

	_ = state.round.StartPlaying()
	r.publishPlayPhaseStart(state)
	return protocol.CommandResponse{Accepted: true}
}

func (r *Table) parseAndValidatePassCards(cardsRaw []string) ([]game.Card, error) {
	if len(cardsRaw) != 3 {
		return nil, fmt.Errorf("pass requires exactly 3 cards")
	}

	cards := make([]game.Card, 0, len(cardsRaw))
	seen := make(map[game.Card]struct{}, len(cardsRaw))
	for _, raw := range cardsRaw {
		card, err := game.ParseCard(raw)
		if err != nil {
			return nil, err
		}
		if _, duplicate := seen[card]; duplicate {
			return nil, fmt.Errorf("pass contains duplicate card %s", card.String())
		}
		seen[card] = struct{}{}
		cards = append(cards, card)
	}

	return cards, nil
}

func (r *Table) startPassReview(state *tableState) {
	// Auto-mark bots ready.
	for _, player := range state.players {
		if player.bot != nil {
			_ = state.round.MarkReady(player.position)
		}
	}

	ready := 0
	for i := 0; i < game.PlayersPerTable; i++ {
		if state.round.IsPassReady(i) {
			ready++
		}
	}

	for _, player := range state.players {
		r.publishPrivate(player.id, protocol.EventHandUpdated, protocol.HandUpdatedData{Cards: game.CardStrings(state.round.Hand(player.position))})
	}

	r.publishPublic(protocol.EventPassReviewStarted, protocol.PassStatusData{
		Submitted: game.PlayersPerTable,
		Total:     game.PlayersPerTable,
		Direction: state.round.PassDirection(),
	})
	r.publishPublic(protocol.EventPassReadyChanged, protocol.PassReadyData{Ready: ready, Total: game.PlayersPerTable})

	if ready == game.PlayersPerTable {
		_ = state.round.StartPlaying()
		r.publishPlayPhaseStart(state)
	}
}

func (r *Table) publishPlayPhaseStart(state *tableState) {
	if state.round == nil {
		return
	}
	turnSeat := state.round.TurnSeat()
	turnPlayer := state.players[turnSeat]
	r.publishTurn(turnPlayer.id, 0)
	r.scheduleBotTurn(state, turnPlayer)
}

// botForPhase returns the bot player if the round is active in the given phase,
// the table is not paused, and the player is a valid bot. Returns nil otherwise.
func botForPhase(state *tableState, playerID protocol.PlayerID, phase game.RoundPhase) *playerState {
	if state.paused || state.round == nil || state.round.Phase() != phase {
		return nil
	}
	player := state.playersByID[playerID]
	if player == nil || player.bot == nil {
		return nil
	}
	return player
}

func (r *Table) handleBotTurn(state *tableState, playerID protocol.PlayerID) {
	player := botForPhase(state, playerID, game.PhasePlaying)
	if player == nil || player.position != state.round.TurnSeat() {
		return
	}

	input := state.round.TurnInput(player.position, state.game.Scores())
	card, err := player.bot.ChoosePlay(input)
	if err != nil {
		return
	}
	_ = r.handlePlay(state, playerID, card.String())
}

func (r *Table) handleBotPass(state *tableState, playerID protocol.PlayerID) {
	player := botForPhase(state, playerID, game.PhasePassing)
	if player == nil || state.round.HasSubmittedPass(player.position) {
		return
	}

	input := state.round.PassInput(player.position, state.game.Scores())
	cards, err := player.bot.ChoosePass(input)
	if err != nil {
		return
	}
	_ = r.handlePass(state, playerID, game.CardStrings(cards))
}

func (r *Table) scheduleBotTurn(_ *tableState, player *playerState) {
	if player.bot == nil {
		return
	}

	go r.submit(botTurnCommand{playerID: player.id})
}

func (r *Table) schedulePassingBots(state *tableState) {
	if state.round == nil || state.round.Phase() != game.PhasePassing {
		return
	}

	for _, player := range state.players {
		if player.bot != nil {
			r.scheduleBotPass(player)
		}
	}
}

func (r *Table) scheduleBotPass(player *playerState) {
	if player.bot == nil {
		return
	}

	go r.submit(botPassCommand{playerID: player.id})
}

func (r *Table) validateStartPreconditions(state *tableState, playerID protocol.PlayerID) string {
	if state.gameOver {
		return ErrGameOver.Error()
	}
	if state.round != nil {
		return ErrRoundInProgress.Error()
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

func validateRoundCommandPreconditions(state *tableState, playerID protocol.PlayerID) string {
	if state.paused {
		return "game is paused"
	}
	if state.round == nil {
		return "round is not running"
	}
	if state.playersByID[playerID] == nil {
		return "player is not seated"
	}
	return ""
}

func (r *Table) initializeRound(state *tableState) *game.Round {
	passDirection := state.game.NextPassDirection()

	var hands [game.PlayersPerTable][]game.Card
	deck := defaultShuffledDeck()
	for i, card := range deck {
		seat := i % game.PlayersPerTable
		hands[seat] = append(hands[seat], card)
	}
	for i := range hands {
		game.SortCards(hands[i])
	}

	return game.NewRound(hands, passDirection)
}

func defaultShuffledDeck() []game.Card {
	deck := game.BuildDeck()
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	game.Shuffle(deck, rng)
	return deck
}

func (r *Table) publishRoundStart(state *tableState, hands map[protocol.PlayerID][]string) {
	r.publishPublic(protocol.EventGameStarted, struct{}{})
	for playerID, cards := range hands {
		r.publishPrivate(playerID, protocol.EventHandUpdated, protocol.HandUpdatedData{Cards: cards})
	}

	if state.round != nil && state.round.PassDirection() != game.PassDirectionHold {
		r.publishPublic(protocol.EventPassSubmitted, protocol.PassStatusData{
			Submitted: 0,
			Total:     game.PlayersPerTable,
			Direction: state.round.PassDirection(),
		})
	}
}

func (r *Table) publishTurn(playerID protocol.PlayerID, trickNumber int) {
	r.publishPublic(protocol.EventTurnChanged, protocol.TurnChangedData{PlayerID: playerID, TrickNumber: trickNumber})
	r.publishPrivate(playerID, protocol.EventYourTurn, protocol.YourTurnData{PlayerID: playerID, TrickNumber: trickNumber})
}

func (r *Table) completeRound(state *tableState) protocol.RoundCompletedData {
	scores := state.round.Scores()
	adjustedRound := make(map[protocol.PlayerID]game.Points, len(state.players))
	for i, player := range state.players {
		adjustedRound[player.id] = scores.Adjusted[i]
	}
	state.roundHistory = append(state.roundHistory, adjustedRound)
	state.game.AddRoundScores(scores.Adjusted)

	return protocol.RoundCompletedData{
		RoundPoints: copyPoints(adjustedRound),
		TotalPoints: r.buildTotals(state),
	}
}

func (r *Table) buildTotals(state *tableState) map[protocol.PlayerID]game.Points {
	totals := make(map[protocol.PlayerID]game.Points, len(state.players))
	scores := state.game.Scores()
	for _, player := range state.players {
		totals[player.id] = scores[player.position]
	}
	return totals
}

func (r *Table) seatWinnersToPlayerIDs(state *tableState, seats []int) []protocol.PlayerID {
	winners := make([]protocol.PlayerID, 0, len(seats))
	for _, seat := range seats {
		winners = append(winners, state.players[seat].id)
	}
	sort.Slice(winners, func(i, j int) bool { return winners[i] < winners[j] })
	return winners
}

func (r *Table) maybeEndGame(state *tableState) {
	if !state.game.IsOver() {
		return
	}
	state.gameOver = true
	totals := r.buildTotals(state)
	r.publishPublic(protocol.EventGameOver, protocol.GameOverData{
		FinalScores: copyPoints(totals),
		Winners:     r.seatWinnersToPlayerIDs(state, state.game.Winners()),
	})
}

func (r *Table) publishGameResumed() {
	slog.Info("game resumed", "event", "game_resumed", "table_id", r.tableID)
	r.publishPublic(protocol.EventGameResumed, protocol.GameResumedData{})
}

func (r *Table) resumeAfterPause(state *tableState) {
	if state.round == nil {
		return
	}
	switch state.round.Phase() {
	case game.PhasePassing:
		r.schedulePassingBots(state)
	case game.PhasePassReview:
		for _, player := range state.players {
			if player.bot != nil && !state.round.IsPassReady(player.position) {
				_ = state.round.MarkReady(player.position)
			}
		}
		ready := 0
		for i := 0; i < game.PlayersPerTable; i++ {
			if state.round.IsPassReady(i) {
				ready++
			}
		}
		r.publishPublic(protocol.EventPassReadyChanged, protocol.PassReadyData{Ready: ready, Total: game.PlayersPerTable})
		if ready == game.PlayersPerTable {
			_ = state.round.StartPlaying()
			r.publishPlayPhaseStart(state)
		}
	case game.PhasePlaying:
		turnSeat := state.round.TurnSeat()
		turnPlayer := state.players[turnSeat]
		r.scheduleBotTurn(state, turnPlayer)
	}
}

func (r *Table) handleRematch(state *tableState, playerID protocol.PlayerID) protocol.CommandResponse {
	if !state.gameOver {
		return protocol.CommandResponse{Accepted: false, Reason: "game is not over"}
	}
	player := state.playersByID[playerID]
	if player == nil || player.bot != nil {
		return protocol.CommandResponse{Accepted: false, Reason: "only seated human players can vote for rematch"}
	}

	if state.rematchVotes == nil {
		state.rematchVotes = make(map[protocol.PlayerID]bool)
	}
	state.rematchVotes[playerID] = true

	humans := 0
	for _, p := range state.players {
		if p.bot == nil {
			humans++
		}
	}
	votes := len(state.rematchVotes)

	slog.Info("rematch vote", "event", "rematch_vote", "table_id", r.tableID, "player_id", playerID, "votes", votes, "total", humans)
	r.publishPublic(protocol.EventRematchVote, protocol.RematchVoteData{
		PlayerID: playerID,
		Votes:    votes,
		Total:    humans,
	})

	if votes >= humans {
		r.startRematch(state)
	}

	return protocol.CommandResponse{Accepted: true}
}

func (r *Table) startRematch(state *tableState) {
	slog.Info("rematch starting", "event", "rematch_starting", "table_id", r.tableID)
	r.publishPublic(protocol.EventRematchStarting, protocol.RematchStartingData{})

	// Remove all bot players.
	humanPlayers := make([]*playerState, 0, len(state.players))
	for _, p := range state.players {
		if p.bot == nil {
			humanPlayers = append(humanPlayers, p)
		} else {
			delete(state.playersByID, p.id)
		}
	}
	state.players = humanPlayers
	for i, p := range state.players {
		p.position = i
	}

	// Reset game state.
	state.game = game.NewGame()
	state.gameOver = false
	state.round = nil
	state.roundHistory = nil
	state.rematchVotes = nil
	state.paused = false
	state.pausedPlayerID = ""

	// Fill empty seats with bots.
	for len(state.players) < game.PlayersPerTable {
		taken := make(map[string]bool, len(state.players))
		for _, p := range state.players {
			taken[p.Name] = true
		}
		id := r.nextPlayerID(state)
		botPlayer := r.addPlayer(state, id, bot.StrategyHard.BotName(taken), bot.StrategyHard.NewBot(), "")
		r.publishPublic(protocol.EventPlayerJoined, protocol.PlayerJoinedData{Player: protocol.PlayerInfo{
			PlayerID: botPlayer.id,
			Name:     botPlayer.Name,
			Seat:     botPlayer.position,
		}})
	}
}

func (r *Table) handleClaimSeat(state *tableState, seat int, name, token string) protocol.JoinResponse {
	token = strings.TrimSpace(token)
	if token == "" {
		return protocol.JoinResponse{Accepted: false, Reason: "player token is required"}
	}

	// Reject if this token is already mapped to a seated player.
	if existing := state.playersByToken[token]; existing != nil {
		return protocol.JoinResponse{Accepted: false, Reason: "already seated"}
	}

	if seat < 0 || seat >= len(state.players) {
		return protocol.JoinResponse{Accepted: false, Reason: "invalid seat"}
	}

	target := state.players[seat]
	if target.bot == nil {
		return protocol.JoinResponse{Accepted: false, Reason: "seat is not bot-controlled"}
	}

	oldName := target.Name
	target.bot = nil
	target.Token = token
	state.playersByToken[token] = target
	delete(state.departedTokens, token)

	name = strings.TrimSpace(name)
	if name != "" {
		target.Name = name
	}

	slog.Info("observer claimed bot seat", "event", "seat_claimed", "table_id", r.tableID, "player_id", target.id, "name", target.Name, "seat", seat, "old_name", oldName)

	r.publishPublic(protocol.EventSeatClaimed, protocol.SeatClaimedData{
		Player:  protocol.PlayerInfo{PlayerID: target.id, Name: target.Name, Seat: seat},
		OldName: oldName,
	})

	// Send the new player their hand if a round is active.
	if state.round != nil {
		r.publishPrivate(target.id, protocol.EventHandUpdated, protocol.HandUpdatedData{
			Cards: game.CardStrings(state.round.Hand(target.position)),
		})
	}

	// If the game was paused for a disconnected player at this seat, resume.
	if state.paused && state.pausedPlayerID == target.id {
		state.paused = false
		state.pausedPlayerID = ""
		r.publishGameResumed()
		r.resumeAfterPause(state)
	}

	return protocol.JoinResponse{
		Accepted: true,
		TableID:  r.tableID,
		PlayerID: target.id,
		Seat:     target.position,
	}
}

func (r *Table) handleResumeGame(state *tableState, playerID protocol.PlayerID) protocol.CommandResponse {
	if !state.paused {
		return protocol.CommandResponse{Accepted: false, Reason: "game is not paused"}
	}
	requester := state.playersByID[playerID]
	if requester == nil || requester.bot != nil {
		return protocol.CommandResponse{Accepted: false, Reason: "only seated human players can resume"}
	}

	state.paused = false
	state.pausedPlayerID = ""
	slog.Info("game resumed by player", "event", "game_resumed", "table_id", r.tableID, "resumed_by", playerID)
	r.publishGameResumed()
	r.resumeAfterPause(state)
	return protocol.CommandResponse{Accepted: true}
}
