package table

import (
	"fmt"
	"log/slog"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/JHK/hearts/internal/game"
	"github.com/JHK/hearts/internal/game/bot"
	"github.com/JHK/hearts/internal/protocol"
)

type StreamEvent struct {
	Type      protocol.EventType
	Data      any
	PrivateTo protocol.PlayerID
}

type AddedBot struct {
	JoinResponse protocol.JoinResponse `json:"join_response"`
	Name         string                `json:"name"`
	Strategy     string                `json:"strategy"`
}

type PlayerSnapshot struct {
	PlayerID protocol.PlayerID `json:"player_id"`
	Name     string        `json:"name"`
	Seat     int           `json:"seat"`
	IsBot    bool          `json:"is_bot"`
}

type TrickPlaySnapshot struct {
	PlayerID protocol.PlayerID `json:"player_id"`
	Name     string        `json:"name"`
	Seat     int           `json:"seat"`
	Card     string        `json:"card"`
}

type Snapshot struct {
	TableID            string                          `json:"table_id"`
	Players            []PlayerSnapshot                `json:"players"`
	Started            bool                            `json:"started"`
	Phase              string                          `json:"phase"`
	TrickNumber        int                             `json:"trick_number"`
	TurnPlayerID       protocol.PlayerID                   `json:"turn_player_id"`
	HeartsBroken       bool                            `json:"hearts_broken"`
	CurrentTrick       []string                        `json:"current_trick"`
	TrickPlays         []TrickPlaySnapshot             `json:"trick_plays"`
	Hand               []string                        `json:"hand"`
	HandSizes          map[protocol.PlayerID]int           `json:"hand_sizes"`
	PassDirection      game.PassDirection              `json:"pass_direction"`
	PassSubmitted      bool                            `json:"pass_submitted"`
	PassSubmittedCount int                             `json:"pass_submitted_count"`
	PassSent           []string                        `json:"pass_sent"`
	PassReceived       []string                        `json:"pass_received"`
	PassReady          bool                            `json:"pass_ready"`
	PassReadyCount     int                             `json:"pass_ready_count"`
	RoundPoints        map[protocol.PlayerID]game.Points   `json:"round_points"`
	RoundHistory       []map[protocol.PlayerID]game.Points `json:"round_history"`
	TotalPoints        map[protocol.PlayerID]game.Points   `json:"total_points"`
	GameOver           bool                            `json:"game_over"`
	Winners            []protocol.PlayerID                 `json:"winners,omitempty"`
}

type roundPhase string

const (
	roundPhasePassing    roundPhase = "passing"
	roundPhasePassReview roundPhase = "pass_review"
	roundPhasePlaying    roundPhase = "playing"
)

type Runtime struct {
	tableID string

	commands  chan any
	stop      chan struct{}
	stopOnce  sync.Once
	stoppedCh chan struct{}

	subsMu    sync.RWMutex
	subs      map[int]chan StreamEvent
	nextSubID int
}

type tableState struct {
	players        []*playerState
	playersByID    map[protocol.PlayerID]*playerState
	playersByToken map[string]*playerState
	roundHistory   []map[protocol.PlayerID]game.Points

	round         *roundState
	roundsStarted int
	nextPlayerSeq int
	gameOver      bool
}

// playerState holds the seat-level data for one player.
// The embedded game.Participant is *game.Player for human players and a bot.Bot for bots.
// Bot detection: _, isBot := player.Participant.(bot.Bot)
type playerState struct {
	id   protocol.PlayerID
	game.Participant
	Name  string
	Seat  int
	Token string
}

type roundState struct {
	Phase         roundPhase
	PassDirection game.PassDirection

	TrickNumber  int
	TurnSeat     int
	HeartsBroken bool
	Trick        []trickPlay
	PlayedCards  []game.Card // cards played in completed tricks
}

type trickPlay struct {
	Seat   int
	Player game.Participant
	Card   game.Card
}

type playUpdate struct {
	playerID    protocol.PlayerID
	cardPlayed  protocol.CardPlayedData
	handUpdated protocol.HandUpdatedData
}

type joinCommand struct {
	name  string
	token string
	reply chan protocol.JoinResponse
}

type startCommand struct {
	playerID protocol.PlayerID
	reply    chan protocol.CommandResponse
}

type playCommand struct {
	playerID protocol.PlayerID
	card     string
	reply    chan protocol.CommandResponse
}

type passCommand struct {
	playerID protocol.PlayerID
	cards    []string
	reply    chan protocol.CommandResponse
}

type readyAfterPassCommand struct {
	playerID protocol.PlayerID
	reply    chan protocol.CommandResponse
}

type addBotCommand struct {
	strategyName string
	reply        chan addBotResult
}

type addBotResult struct {
	added AddedBot
	err   error
}

type snapshotCommand struct {
	forPlayer protocol.PlayerID
	reply     chan Snapshot
}

type leaveCommand struct {
	playerID protocol.PlayerID
	reply    chan struct{}
}

type infoCommand struct {
	reply chan protocol.TableInfo
}

type botTurnCommand struct {
	playerID protocol.PlayerID
}

type botPassCommand struct {
	playerID protocol.PlayerID
}

func NewRuntime(tableID string) *Runtime {
	r := &Runtime{
		tableID:   tableID,
		commands:  make(chan any),
		stop:      make(chan struct{}),
		stoppedCh: make(chan struct{}),
		subs:      make(map[int]chan StreamEvent),
	}

	go r.run()
	return r
}

func (r *Runtime) ID() string {
	return r.tableID
}

func (r *Runtime) Close() {
	r.stopOnce.Do(func() {
		close(r.stop)
		<-r.stoppedCh

		r.subsMu.Lock()
		for id, ch := range r.subs {
			close(ch)
			delete(r.subs, id)
		}
		r.subsMu.Unlock()
	})
}

func (r *Runtime) Subscribe() (<-chan StreamEvent, func()) {
	r.subsMu.Lock()
	r.nextSubID++
	id := r.nextSubID
	ch := make(chan StreamEvent, 128)
	r.subs[id] = ch
	r.subsMu.Unlock()

	unsubscribe := func() {
		r.subsMu.Lock()
		sub, ok := r.subs[id]
		if ok {
			delete(r.subs, id)
			close(sub)
		}
		r.subsMu.Unlock()
	}

	return ch, unsubscribe
}

func (r *Runtime) Join(name, token string) (protocol.JoinResponse, error) {
	reply := make(chan protocol.JoinResponse, 1)
	if !r.submit(joinCommand{name: name, token: token, reply: reply}) {
		return protocol.JoinResponse{}, fmt.Errorf("table is stopping")
	}
	return <-reply, nil
}

func (r *Runtime) Start(playerID protocol.PlayerID) protocol.CommandResponse {
	reply := make(chan protocol.CommandResponse, 1)
	if !r.submit(startCommand{playerID: playerID, reply: reply}) {
		return protocol.CommandResponse{Accepted: false, Reason: "table is stopping"}
	}
	return <-reply
}

func (r *Runtime) Play(playerID protocol.PlayerID, card string) protocol.CommandResponse {
	reply := make(chan protocol.CommandResponse, 1)
	if !r.submit(playCommand{playerID: playerID, card: card, reply: reply}) {
		return protocol.CommandResponse{Accepted: false, Reason: "table is stopping"}
	}
	return <-reply
}

func (r *Runtime) Pass(playerID protocol.PlayerID, cards []string) protocol.CommandResponse {
	reply := make(chan protocol.CommandResponse, 1)
	if !r.submit(passCommand{playerID: playerID, cards: cards, reply: reply}) {
		return protocol.CommandResponse{Accepted: false, Reason: "table is stopping"}
	}
	return <-reply
}

func (r *Runtime) ReadyAfterPass(playerID protocol.PlayerID) protocol.CommandResponse {
	reply := make(chan protocol.CommandResponse, 1)
	if !r.submit(readyAfterPassCommand{playerID: playerID, reply: reply}) {
		return protocol.CommandResponse{Accepted: false, Reason: "table is stopping"}
	}
	return <-reply
}

func (r *Runtime) AddBot(strategyName string) (AddedBot, error) {
	reply := make(chan addBotResult, 1)
	if !r.submit(addBotCommand{strategyName: strategyName, reply: reply}) {
		return AddedBot{}, fmt.Errorf("table is stopping")
	}
	result := <-reply
	return result.added, result.err
}

func (r *Runtime) Snapshot(forPlayer protocol.PlayerID) Snapshot {
	reply := make(chan Snapshot, 1)
	if !r.submit(snapshotCommand{forPlayer: forPlayer, reply: reply}) {
		return Snapshot{TableID: r.tableID}
	}
	return <-reply
}

func (r *Runtime) Leave(playerID protocol.PlayerID) {
	if playerID == "" {
		return
	}

	reply := make(chan struct{}, 1)
	if !r.submit(leaveCommand{playerID: playerID, reply: reply}) {
		return
	}
	<-reply
}

func (r *Runtime) Info() protocol.TableInfo {
	reply := make(chan protocol.TableInfo, 1)
	if !r.submit(infoCommand{reply: reply}) {
		return protocol.TableInfo{TableID: r.tableID, MaxPlayers: game.PlayersPerTable}
	}
	return <-reply
}

func (r *Runtime) submit(command any) bool {
	select {
	case <-r.stop:
		return false
	case r.commands <- command:
		return true
	}
}

func (r *Runtime) run() {
	defer close(r.stoppedCh)

	state := &tableState{
		playersByID:    make(map[protocol.PlayerID]*playerState),
		playersByToken: make(map[string]*playerState),
	}

	for {
		select {
		case <-r.stop:
			return
		case command := <-r.commands:
			switch cmd := command.(type) {
			case joinCommand:
				cmd.reply <- r.handleJoin(state, cmd.name, cmd.token)
			case startCommand:
				cmd.reply <- r.handleStart(state, cmd.playerID)
			case playCommand:
				cmd.reply <- r.handlePlay(state, cmd.playerID, cmd.card)
			case passCommand:
				cmd.reply <- r.handlePass(state, cmd.playerID, cmd.cards)
			case readyAfterPassCommand:
				cmd.reply <- r.handleReadyAfterPass(state, cmd.playerID)
			case addBotCommand:
				added, err := r.handleAddBot(state, cmd.strategyName)
				cmd.reply <- addBotResult{added: added, err: err}
			case snapshotCommand:
				cmd.reply <- r.buildSnapshot(state, cmd.forPlayer)
			case leaveCommand:
				r.handleLeave(state, cmd.playerID)
				cmd.reply <- struct{}{}
			case infoCommand:
				cmd.reply <- protocol.TableInfo{
					TableID:    r.tableID,
					Players:    len(state.players),
					MaxPlayers: game.PlayersPerTable,
					Started:    state.round != nil,
					GameOver:   state.gameOver,
				}
			case botTurnCommand:
				r.handleBotTurn(state, cmd.playerID)
			case botPassCommand:
				r.handleBotPass(state, cmd.playerID)
			}
		}
	}
}

func (r *Runtime) handleLeave(state *tableState, playerID protocol.PlayerID) {
	player := state.playersByID[playerID]
	if player == nil {
		return
	}

	slog.Info("player left table", "event", "player_left", "table_id", r.tableID, "player_id", playerID, "name", player.Name)

	if state.round != nil {
		// Preserve the token so the player can reclaim their seat if they reconnect.
		// Convert human to bot; if already a bot, leave as is.
		if _, isBot := player.Participant.(bot.Bot); !isBot {
			human := player.Participant.(*game.Player)
			player.Participant = bot.StrategyRandom.WrapPlayer(human)
		}

		switch state.round.Phase {
		case roundPhasePassing:
			if !player.HasSubmittedPass() {
				if cards, err := r.chooseBotPassCards(player, state.round.PassDirection); err == nil {
					_ = r.handlePass(state, playerID, cards)
				}
			}
		case roundPhasePassReview:
			if !player.PassReady() {
				_ = r.handleReadyAfterPass(state, playerID)
			}
		case roundPhasePlaying:
			if player.Seat == state.round.TurnSeat {
				r.scheduleBotTurn(state, player)
			}
		}
		return
	}

	delete(state.playersByID, playerID)
	if player.Token != "" {
		delete(state.playersByToken, player.Token)
	}

	for index, seated := range state.players {
		if seated.id != playerID {
			continue
		}

		state.players = append(state.players[:index], state.players[index+1:]...)
		for seat, updated := range state.players {
			updated.Seat = seat
		}
		return
	}
}

func (r *Runtime) handleJoin(state *tableState, name, token string) protocol.JoinResponse {
	token = strings.TrimSpace(token)
	if token == "" {
		return protocol.JoinResponse{Accepted: false, Reason: "player token is required"}
	}

	if existing := state.playersByToken[token]; existing != nil {
		if b, isBot := existing.Participant.(bot.Bot); isBot {
			// Player reconnected after being converted to a bot mid-round — reclaim the seat.
			existing.Participant = b.Unwrap()
			if n := strings.TrimSpace(name); n != "" {
				existing.Name = n
			}
			slog.Info("player reclaimed seat", "event", "player_reclaimed", "table_id", r.tableID, "player_id", existing.id, "name", existing.Name, "seat", existing.Seat)
		}
		return protocol.JoinResponse{
			Accepted: true,
			TableID:  r.tableID,
			PlayerID: existing.id,
			Seat:     existing.Seat,
		}
	}

	if state.round != nil {
		return protocol.JoinResponse{Accepted: false, Reason: "round already in progress"}
	}
	if len(state.players) >= game.PlayersPerTable {
		return protocol.JoinResponse{Accepted: false, Reason: "table is full"}
	}

	name = strings.TrimSpace(name)
	if name == "" {
		name = "Player"
	}

	id := r.nextPlayerID(state)
	player := r.addPlayer(state, id, name, game.NewPlayer(), token)

	slog.Info("player joined table", "event", "player_joined", "table_id", r.tableID, "player_id", player.id, "name", player.Name, "seat", player.Seat)

	r.publishPublic(protocol.EventPlayerJoined, protocol.PlayerJoinedData{Player: protocol.PlayerInfo{
		PlayerID: player.id,
		Name:     player.Name,
		Seat:     player.Seat,
	}})

	return protocol.JoinResponse{
		Accepted: true,
		TableID:  r.tableID,
		PlayerID: player.id,
		Seat:     player.Seat,
	}
}

func (r *Runtime) handleAddBot(state *tableState, strategyName string) (AddedBot, error) {
	if state.gameOver {
		return AddedBot{}, fmt.Errorf("game is over")
	}
	if state.round != nil {
		return AddedBot{}, fmt.Errorf("round already in progress")
	}
	if len(state.players) >= game.PlayersPerTable {
		return AddedBot{}, fmt.Errorf("table is full")
	}

	strategyKind, err := bot.ParseStrategyKind(strategyName)
	if err != nil {
		return AddedBot{}, err
	}

	id := r.nextPlayerID(state)
	b := strategyKind.NewBot()
	player := r.addPlayer(state, id, b.BotName(), b, "")

	slog.Debug("bot added to table", "event", "bot_added", "table_id", r.tableID, "player_id", player.id, "name", player.Name, "strategy", string(strategyKind))

	r.publishPublic(protocol.EventPlayerJoined, protocol.PlayerJoinedData{Player: protocol.PlayerInfo{
		PlayerID: player.id,
		Name:     player.Name,
		Seat:     player.Seat,
	}})

	return AddedBot{
		JoinResponse: protocol.JoinResponse{
			Accepted: true,
			TableID:  r.tableID,
			PlayerID: player.id,
			Seat:     player.Seat,
		},
		Name:     player.Name,
		Strategy: string(strategyKind),
	}, nil
}

func (r *Runtime) addPlayer(state *tableState, id protocol.PlayerID, name string, p game.Participant, token string) *playerState {
	player := &playerState{
		id:          id,
		Participant: p,
		Name:        name,
		Seat:        len(state.players),
		Token:       token,
	}

	state.players = append(state.players, player)
	state.playersByID[id] = player
	if token != "" {
		state.playersByToken[token] = player
	}

	return player
}

func (r *Runtime) handleStart(state *tableState, playerID protocol.PlayerID) protocol.CommandResponse {
	if reason := r.validateStartPreconditions(state, playerID); reason != "" {
		return protocol.CommandResponse{Accepted: false, Reason: reason}
	}

	hands := r.initializeRound(state)
	slog.Info("table started", "event", "table_started", "table_id", r.tableID, "round", state.roundsStarted)
	r.publishRoundStart(state, hands)
	if state.round.PassDirection == game.PassDirectionHold {
		r.startPlayPhase(state)
		return protocol.CommandResponse{Accepted: true}
	}

	r.schedulePassingBots(state)

	return protocol.CommandResponse{Accepted: true}
}

func (r *Runtime) nextPlayerID(state *tableState) protocol.PlayerID {
	for {
		state.nextPlayerSeq++
		candidate := protocol.PlayerID(fmt.Sprintf("p-%d", state.nextPlayerSeq))
		if _, exists := state.playersByID[candidate]; !exists {
			return candidate
		}
	}
}

func (r *Runtime) handlePlay(state *tableState, playerID protocol.PlayerID, cardRaw string) protocol.CommandResponse {
	card, err := game.ParseCard(cardRaw)
	if err != nil {
		return protocol.CommandResponse{Accepted: false, Reason: err.Error()}
	}

	if state.round == nil {
		return protocol.CommandResponse{Accepted: false, Reason: "round is not running"}
	}
	if state.round.Phase != roundPhasePlaying {
		return protocol.CommandResponse{Accepted: false, Reason: "round is not in play phase"}
	}

	player := state.playersByID[playerID]
	if player == nil {
		return protocol.CommandResponse{Accepted: false, Reason: "player is not seated"}
	}
	if player.Seat != state.round.TurnSeat {
		return protocol.CommandResponse{Accepted: false, Reason: "not your turn"}
	}

	err = game.ValidatePlay(game.ValidatePlayInput{
		Hand:         player.Hand(),
		Card:         card,
		Trick:        currentTrickCards(state.round),
		HeartsBroken: state.round.HeartsBroken,
		FirstTrick:   state.round.TrickNumber == 0,
	})
	if err != nil {
		return protocol.CommandResponse{Accepted: false, Reason: err.Error()}
	}

	update, err := r.applyValidatedPlay(state.round, player, card)
	if err != nil {
		return protocol.CommandResponse{Accepted: false, Reason: err.Error()}
	}

	if len(state.round.Trick) < game.PlayersPerTable {
		nextSeat := (state.round.TurnSeat + 1) % game.PlayersPerTable
		nextPlayer := state.players[nextSeat]
		trickNumber := state.round.TrickNumber
		state.round.TurnSeat = nextSeat

		r.publishPlayAndHand(update)
		r.publishTurn(nextPlayer.id, trickNumber)
		r.scheduleBotTurn(state, nextPlayer)
		return protocol.CommandResponse{Accepted: true}
	}

	winner, points, err := game.TrickWinner(currentTrickPlays(state.round))
	if err != nil {
		return protocol.CommandResponse{Accepted: false, Reason: err.Error()}
	}

	winner.AddTrickPoints(points)
	var nextSeat int
	for _, tp := range state.round.Trick {
		if tp.Player == winner {
			nextSeat = tp.Seat
			break
		}
	}
	winnerPlayerID := state.players[nextSeat].id
	trickNumber := state.round.TrickNumber
	trickCompleted := protocol.TrickCompletedData{TrickNumber: trickNumber, WinnerPlayerID: winnerPlayerID, Points: points}

	if trickNumber == 12 {
		roundCompleted := r.completeRound(state)
		state.round = nil

		r.publishPlayAndHand(update)
		r.publishPublic(protocol.EventTrickCompleted, trickCompleted)
		r.publishPublic(protocol.EventRoundCompleted, roundCompleted)
		r.maybeEndGame(state)
		return protocol.CommandResponse{Accepted: true}
	}
	nextTrick := trickNumber + 1

	for _, tp := range state.round.Trick {
		state.round.PlayedCards = append(state.round.PlayedCards, tp.Card)
	}
	state.round.Trick = nil
	state.round.TrickNumber = nextTrick
	state.round.TurnSeat = nextSeat

	r.publishPlayAndHand(update)
	r.publishPublic(protocol.EventTrickCompleted, trickCompleted)
	r.publishTurn(winnerPlayerID, nextTrick)
	r.scheduleBotTurn(state, state.players[nextSeat])

	return protocol.CommandResponse{Accepted: true}
}

func (r *Runtime) handlePass(state *tableState, playerID protocol.PlayerID, cardsRaw []string) protocol.CommandResponse {
	if state.round == nil {
		return protocol.CommandResponse{Accepted: false, Reason: "round is not running"}
	}
	if state.round.Phase != roundPhasePassing {
		return protocol.CommandResponse{Accepted: false, Reason: "round is not in pass phase"}
	}

	player := state.playersByID[playerID]
	if player == nil {
		return protocol.CommandResponse{Accepted: false, Reason: "player is not seated"}
	}
	if player.HasSubmittedPass() {
		return protocol.CommandResponse{Accepted: false, Reason: "pass already submitted"}
	}

	passCards, err := r.parseAndValidatePassCards(player.Hand(), cardsRaw)
	if err != nil {
		return protocol.CommandResponse{Accepted: false, Reason: err.Error()}
	}

	player.SubmitPass(passCards)

	submitted := 0
	for _, p := range state.players {
		if p.HasSubmittedPass() {
			submitted++
		}
	}

	r.publishPublic(protocol.EventPassSubmitted, protocol.PassStatusData{
		Submitted: submitted,
		Total:     game.PlayersPerTable,
		Direction: state.round.PassDirection,
	})

	if submitted < game.PlayersPerTable {
		return protocol.CommandResponse{Accepted: true}
	}

	r.applyPassSubmissions(state)
	r.startPassReview(state)

	return protocol.CommandResponse{Accepted: true}
}

func (r *Runtime) handleReadyAfterPass(state *tableState, playerID protocol.PlayerID) protocol.CommandResponse {
	if state.round == nil {
		return protocol.CommandResponse{Accepted: false, Reason: "round is not running"}
	}
	if state.round.Phase != roundPhasePassReview {
		return protocol.CommandResponse{Accepted: false, Reason: "round is not in pass review"}
	}
	if state.playersByID[playerID] == nil {
		return protocol.CommandResponse{Accepted: false, Reason: "player is not seated"}
	}

	state.playersByID[playerID].MarkPassReady()

	ready := 0
	for _, p := range state.players {
		if p.PassReady() {
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

	r.startPlayPhase(state)
	return protocol.CommandResponse{Accepted: true}
}

func (r *Runtime) parseAndValidatePassCards(hand []game.Card, cardsRaw []string) ([]game.Card, error) {
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
		if !game.ContainsCard(hand, card) {
			return nil, fmt.Errorf("card %s is not in hand", card.String())
		}

		seen[card] = struct{}{}
		cards = append(cards, card)
	}

	return cards, nil
}

func (r *Runtime) applyPassSubmissions(state *tableState) {
	var passes [game.PlayersPerTable][]game.Card
	for seat, player := range state.players {
		passes[seat] = player.PassSent()
	}

	received := game.ExchangePasses(passes, state.round.PassDirection)

	for seat, player := range state.players {
		player.SendPassCards()
		player.ReceivePassCards(received[seat])
	}
}

func (r *Runtime) startPassReview(state *tableState) {
	state.round.Phase = roundPhasePassReview
	ready := 0
	for _, player := range state.players {
		if _, isBot := player.Participant.(bot.Bot); isBot {
			player.MarkPassReady()
		}
		if player.PassReady() {
			ready++
		}
		r.publishPrivate(player.id, protocol.EventHandUpdated, protocol.HandUpdatedData{Cards: game.CardStrings(player.Hand())})
	}

	r.publishPublic(protocol.EventPassReviewStarted, protocol.PassStatusData{
		Submitted: game.PlayersPerTable,
		Total:     game.PlayersPerTable,
		Direction: state.round.PassDirection,
	})
	r.publishPublic(protocol.EventPassReadyChanged, protocol.PassReadyData{Ready: ready, Total: game.PlayersPerTable})

	if ready == game.PlayersPerTable {
		r.startPlayPhase(state)
	}
}

func (r *Runtime) startPlayPhase(state *tableState) {
	if state.round == nil {
		return
	}

	state.round.Phase = roundPhasePlaying
	state.round.Trick = nil
	state.round.TrickNumber = 0
	state.round.HeartsBroken = false

	startSeat := r.findTwoClubsSeat(state)
	state.round.TurnSeat = startSeat
	turnPlayer := state.players[startSeat]

	r.publishTurn(turnPlayer.id, 0)
	r.scheduleBotTurn(state, turnPlayer)
}

func (r *Runtime) findTwoClubsSeat(state *tableState) int {
	twoClubs := game.Card{Suit: game.SuitClubs, Rank: 2}
	for _, player := range state.players {
		if game.ContainsCard(player.Hand(), twoClubs) {
			return player.Seat
		}
	}

	return 0
}

func (r *Runtime) handleBotTurn(state *tableState, playerID protocol.PlayerID) {
	if state.round == nil || state.round.Phase != roundPhasePlaying {
		return
	}

	player := state.playersByID[playerID]
	if player == nil || player.Seat != state.round.TurnSeat {
		return
	}

	b, ok := player.Participant.(bot.Bot)
	if !ok {
		return
	}

	card, err := b.ChoosePlay(bot.TurnInput{
		Hand:         player.Hand(),
		Trick:        currentTrickCards(state.round),
		HeartsBroken: state.round.HeartsBroken,
		FirstTrick:   state.round.TrickNumber == 0,
		PlayedCards:  state.round.PlayedCards,
	})
	if err != nil {
		return
	}

	_ = r.handlePlay(state, playerID, card.String())
}

func (r *Runtime) handleBotPass(state *tableState, playerID protocol.PlayerID) {
	if state.round == nil || state.round.Phase != roundPhasePassing {
		return
	}

	player := state.playersByID[playerID]
	if player == nil || player.HasSubmittedPass() {
		return
	}

	cards, err := r.chooseBotPassCards(player, state.round.PassDirection)
	if err != nil {
		return
	}

	_ = r.handlePass(state, playerID, cards)
}

func (r *Runtime) scheduleBotTurn(state *tableState, player *playerState) {
	if _, ok := player.Participant.(bot.Bot); !ok {
		return
	}

	playerID := player.id
	go func() {
		timer := time.NewTimer(350 * time.Millisecond)
		defer timer.Stop()

		select {
		case <-r.stop:
			return
		case <-timer.C:
		}

		r.submit(botTurnCommand{playerID: playerID})
	}()
}

func (r *Runtime) schedulePassingBots(state *tableState) {
	if state.round == nil || state.round.Phase != roundPhasePassing {
		return
	}

	for _, player := range state.players {
		if _, isBot := player.Participant.(bot.Bot); !isBot {
			continue
		}
		r.scheduleBotPass(player)
	}
}

func (r *Runtime) scheduleBotPass(player *playerState) {
	if _, ok := player.Participant.(bot.Bot); !ok {
		return
	}

	playerID := player.id
	go func() {
		timer := time.NewTimer(300 * time.Millisecond)
		defer timer.Stop()

		select {
		case <-r.stop:
			return
		case <-timer.C:
		}

		r.submit(botPassCommand{playerID: playerID})
	}()
}

func (r *Runtime) chooseBotPassCards(player *playerState, dir game.PassDirection) ([]string, error) {
	b, ok := player.Participant.(bot.Bot)
	if !ok {
		return nil, fmt.Errorf("not a bot")
	}

	cards, err := b.ChoosePass(bot.PassInput{
		Hand:      append([]game.Card(nil), player.Hand()...),
		Direction: dir,
	})
	if err != nil {
		return nil, err
	}

	return game.CardStrings(cards), nil
}

func (r *Runtime) validateStartPreconditions(state *tableState, playerID protocol.PlayerID) string {
	if state.gameOver {
		return "game is over"
	}
	if state.round != nil {
		return "round already started"
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

func (r *Runtime) initializeRound(state *tableState) map[protocol.PlayerID][]string {
	passDirection := game.PassDirectionForRound(state.roundsStarted)
	state.roundsStarted++

	var handSlices [game.PlayersPerTable][]game.Card
	deck := defaultShuffledDeck()
	for i, card := range deck {
		seat := i % game.PlayersPerTable
		handSlices[seat] = append(handSlices[seat], card)
	}

	hands := make(map[protocol.PlayerID][]string, len(state.players))
	for _, player := range state.players {
		game.SortCards(handSlices[player.Seat])
		player.DealCards(handSlices[player.Seat])
		hands[player.id] = game.CardStrings(player.Hand())
	}

	state.round = &roundState{
		Phase:         roundPhasePassing,
		PassDirection: passDirection,
	}

	return hands
}

func defaultShuffledDeck() []game.Card {
	deck := game.BuildDeck()
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	game.Shuffle(deck, rng)
	return deck
}

func (r *Runtime) publishRoundStart(state *tableState, hands map[protocol.PlayerID][]string) {
	r.publishPublic(protocol.EventGameStarted, struct{}{})
	for playerID, cards := range hands {
		r.publishPrivate(playerID, protocol.EventHandUpdated, protocol.HandUpdatedData{Cards: cards})
	}

	if state.round != nil && state.round.PassDirection != game.PassDirectionHold {
		r.publishPublic(protocol.EventPassSubmitted, protocol.PassStatusData{
			Submitted: 0,
			Total:     game.PlayersPerTable,
			Direction: state.round.PassDirection,
		})
	}
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
		plays[i] = game.Play{Player: played.Player, Card: played.Card}
	}
	return plays
}

func (r *Runtime) applyValidatedPlay(round *roundState, player *playerState, card game.Card) (playUpdate, error) {
	if !player.PlayCard(card) {
		return playUpdate{}, fmt.Errorf("card not found in hand")
	}

	breaksHearts := card.Suit == game.SuitHearts && !round.HeartsBroken
	if card.Suit == game.SuitHearts {
		round.HeartsBroken = true
	}

	round.Trick = append(round.Trick, trickPlay{Seat: player.Seat, Player: player.Participant, Card: card})

	return playUpdate{
		playerID:    player.id,
		cardPlayed:  protocol.CardPlayedData{PlayerID: player.id, Card: card.String(), BreaksHearts: breaksHearts},
		handUpdated: protocol.HandUpdatedData{Cards: game.CardStrings(player.Hand())},
	}, nil
}

func (r *Runtime) publishPlayAndHand(update playUpdate) {
	r.publishPublic(protocol.EventCardPlayed, update.cardPlayed)
	r.publishPrivate(update.playerID, protocol.EventHandUpdated, update.handUpdated)
}

func (r *Runtime) publishTurn(playerID protocol.PlayerID, trickNumber int) {
	r.publishPublic(protocol.EventTurnChanged, protocol.TurnChangedData{PlayerID: playerID, TrickNumber: trickNumber})
	r.publishPrivate(playerID, protocol.EventYourTurn, protocol.YourTurnData{PlayerID: playerID, TrickNumber: trickNumber})
}

func (r *Runtime) completeRound(state *tableState) protocol.RoundCompletedData {
	var rawPoints [game.PlayersPerTable]game.Points
	for i, player := range state.players {
		rawPoints[i] = player.RoundPoints()
	}
	adjusted := game.ApplyShootTheMoon(rawPoints)
	adjustedRound := make(map[protocol.PlayerID]game.Points, len(state.players))
	for i, player := range state.players {
		player.FinalizeRound(adjusted[i])
		adjustedRound[player.id] = adjusted[i]
	}
	state.roundHistory = append(state.roundHistory, adjustedRound)

	totals := make(map[protocol.PlayerID]game.Points, len(state.players))
	for _, player := range state.players {
		totals[player.id] = player.CumulativePoints()
	}
	return protocol.RoundCompletedData{
		RoundPoints: copyPoints(adjustedRound),
		TotalPoints: totals,
	}
}

const gameOverThreshold = game.Points(100)

func computeWinners(totals map[protocol.PlayerID]game.Points) []protocol.PlayerID {
	var minPts game.Points
	first := true
	for _, pts := range totals {
		if first || pts < minPts {
			minPts = pts
			first = false
		}
	}

	winners := make([]protocol.PlayerID, 0, len(totals))
	for playerID, pts := range totals {
		if pts == minPts {
			winners = append(winners, playerID)
		}
	}
	sort.Slice(winners, func(i, j int) bool { return winners[i] < winners[j] })
	return winners
}

func (r *Runtime) maybeEndGame(state *tableState) {
	for _, player := range state.players {
		if player.CumulativePoints() >= gameOverThreshold {
			state.gameOver = true
			totals := make(map[protocol.PlayerID]game.Points, len(state.players))
			for _, p := range state.players {
				totals[p.id] = p.CumulativePoints()
			}
			r.publishPublic(protocol.EventGameOver, protocol.GameOverData{
				FinalScores: copyPoints(totals),
				Winners:     computeWinners(totals),
			})
			return
		}
	}
}

func (r *Runtime) buildSnapshot(state *tableState, forPlayer protocol.PlayerID) Snapshot {
	playerSnapshots := make([]PlayerSnapshot, 0, len(state.players))
	for _, player := range state.players {
		_, isBot := player.Participant.(bot.Bot)
		playerSnapshots = append(playerSnapshots, PlayerSnapshot{
			PlayerID: player.id,
			Name:     player.Name,
			Seat:     player.Seat,
			IsBot:    isBot,
		})
	}
	sort.Slice(playerSnapshots, func(i, j int) bool {
		return playerSnapshots[i].Seat < playerSnapshots[j].Seat
	})

	totals := make(map[protocol.PlayerID]game.Points, len(state.players))
	for _, player := range state.players {
		totals[player.id] = player.CumulativePoints()
	}

	snapshot := Snapshot{
		TableID:      r.tableID,
		Players:      playerSnapshots,
		Started:      state.round != nil,
		Phase:        "",
		HandSizes:    map[protocol.PlayerID]int{},
		RoundPoints:  map[protocol.PlayerID]game.Points{},
		RoundHistory: copyRoundHistory(state.roundHistory),
		TotalPoints:  totals,
		GameOver:     state.gameOver,
	}

	if state.gameOver {
		snapshot.Winners = computeWinners(totals)
	}

	for _, player := range state.players {
		snapshot.HandSizes[player.id] = len(player.Hand())
	}

	if state.round != nil {
		snapshot.Phase = string(state.round.Phase)
		snapshot.TrickNumber = state.round.TrickNumber
		if state.round.Phase == roundPhasePlaying {
			snapshot.TurnPlayerID = state.players[state.round.TurnSeat].id
		}
		snapshot.HeartsBroken = state.round.HeartsBroken
		snapshot.PassDirection = state.round.PassDirection

		submitted := 0
		readyCount := 0
		for _, p := range state.players {
			if p.HasSubmittedPass() {
				submitted++
			}
			if p.PassReady() {
				readyCount++
			}
		}
		snapshot.PassSubmittedCount = submitted
		snapshot.PassReadyCount = readyCount

		snapshot.CurrentTrick = make([]string, 0, len(state.round.Trick))
		snapshot.TrickPlays = make([]TrickPlaySnapshot, 0, len(state.round.Trick))
		for _, played := range state.round.Trick {
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
		for _, p := range state.players {
			roundPoints[p.id] = p.RoundPoints()
		}
		snapshot.RoundPoints = roundPoints
	}

	if forPlayer != "" {
		if player := state.playersByID[forPlayer]; player != nil {
			snapshot.Hand = game.CardStrings(player.Hand())
			if state.round != nil {
				snapshot.PassSubmitted = player.HasSubmittedPass()
				snapshot.PassReady = player.PassReady()
				snapshot.PassSent = game.CardStrings(player.PassSent())
				snapshot.PassReceived = game.CardStrings(player.PassReceived())
			}
		}
	}

	return snapshot
}

func (r *Runtime) publishPublic(eventType protocol.EventType, payload any) {
	r.emit(StreamEvent{Type: eventType, Data: payload})
}

func (r *Runtime) publishPrivate(playerID protocol.PlayerID, eventType protocol.EventType, payload any) {
	r.emit(StreamEvent{Type: eventType, Data: payload, PrivateTo: playerID})
}

func (r *Runtime) emit(event StreamEvent) {
	r.subsMu.RLock()
	defer r.subsMu.RUnlock()

	for _, sub := range r.subs {
		select {
		case sub <- event:
		default:
		}
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
