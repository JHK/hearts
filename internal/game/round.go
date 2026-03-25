package game

import "fmt"

const PlayersPerTable = 4

// RoundPhase represents the current phase of a round.
type RoundPhase int

const (
	PhasePassing    RoundPhase = iota // players submit 3 cards to pass
	PhasePassReview                   // players review received cards
	PhasePlaying                      // trick-by-trick card play
	PhaseComplete                     // all 13 tricks played
)

// TrickResult is returned by Round.Play when a trick completes.
type TrickResult struct {
	WinnerSeat  int
	Points      Points
	TrickNumber int // 0-12
}

// RoundScores holds raw and shoot-the-moon-adjusted points for a completed round.
type RoundScores struct {
	Raw      [PlayersPerTable]Points
	Adjusted [PlayersPerTable]Points
}

// TurnInput is the game state delivered to a decision-maker when it must choose a card to play.
type TurnInput struct {
	Hand        []Card
	Trick       []Card
	HeartsBroken bool
	FirstTrick  bool
	PlayedCards []Card                    // all cards played in completed tricks this round
	RoundPoints [PlayersPerTable]Points   // penalty points captured per seat this round
	MySeat      int                       // this decision-maker's seat index
}

// PassInput is the game state delivered to a decision-maker when it must choose cards to pass.
type PassInput struct {
	Hand      []Card
	Direction PassDirection
}

// Round is a step-at-a-time state machine for one round of Hearts.
// It owns all per-round state: hands, tricks, passes, and scoring.
// Callers drive phase transitions explicitly, giving them hooks for event emission.
type Round struct {
	phase   RoundPhase
	passDir PassDirection

	hands [PlayersPerTable][]Card

	// pass state
	passSubmitted [PlayersPerTable][]Card
	passReceived  [PlayersPerTable][]Card
	passReady     [PlayersPerTable]bool

	// play state
	trickNumber  int
	turnSeat     int
	heartsBroken bool
	currentTrick []Play
	playedCards  []Card
	roundPoints  [PlayersPerTable]Points
}

// NewRound creates a Round with dealt hands and a pass direction.
// The initial phase is PhasePassing.
func NewRound(hands [PlayersPerTable][]Card, passDir PassDirection) *Round {
	r := &Round{
		phase:   PhasePassing,
		passDir: passDir,
	}
	for i, h := range hands {
		r.hands[i] = append([]Card(nil), h...)
	}
	return r
}

// NewTestRound creates a Round in PhaseComplete with preset round points.
// Intended for test setup only.
func NewTestRound(roundPoints [PlayersPerTable]Points) *Round {
	return &Round{phase: PhaseComplete, roundPoints: roundPoints}
}

// --- Queries ---

func (r *Round) Phase() RoundPhase            { return r.phase }
func (r *Round) PassDirection() PassDirection { return r.passDir }
func (r *Round) TurnSeat() int                { return r.turnSeat }
func (r *Round) TrickNumber() int             { return r.trickNumber }
func (r *Round) HeartsBroken() bool           { return r.heartsBroken }
func (r *Round) PlayedCards() []Card          { return r.playedCards }
func (r *Round) CurrentTrick() []Play         { return r.currentTrick }
func (r *Round) RoundPoints(seat int) Points  { return r.roundPoints[seat] }

func (r *Round) Hand(seat int) []Card           { return r.hands[seat] }
func (r *Round) HasSubmittedPass(seat int) bool { return r.passSubmitted[seat] != nil }
func (r *Round) IsPassReady(seat int) bool      { return r.passReady[seat] }
func (r *Round) PassSent(seat int) []Card       { return r.passSubmitted[seat] }
func (r *Round) PassReceived(seat int) []Card   { return r.passReceived[seat] }

// TurnInput builds the decision input for a bot or UI at the given seat.
func (r *Round) TurnInput(seat int) TurnInput {
	return TurnInput{
		Hand:         r.hands[seat],
		Trick:        r.currentTrickCards(),
		HeartsBroken: r.heartsBroken,
		FirstTrick:   r.trickNumber == 0,
		PlayedCards:  r.playedCards,
		RoundPoints:  r.roundPoints,
		MySeat:       seat,
	}
}

// PassInput builds the pass decision input for a bot or UI at the given seat.
func (r *Round) PassInput(seat int) PassInput {
	return PassInput{
		Hand:      append([]Card(nil), r.hands[seat]...),
		Direction: r.passDir,
	}
}

// --- Pass phase mutations ---

// SubmitPass records the cards a seat intends to pass.
func (r *Round) SubmitPass(seat int, cards []Card) error {
	if r.phase != PhasePassing {
		return errWrongPhase("submit pass", "passing")
	}
	if r.passSubmitted[seat] != nil {
		return ErrPassAlreadySubmitted
	}
	for _, card := range cards {
		if !ContainsCard(r.hands[seat], card) {
			return errNotInHand(card)
		}
	}
	r.passSubmitted[seat] = append([]Card(nil), cards...)
	return nil
}

// ApplyPasses exchanges submitted cards and transitions to PhasePassReview.
func (r *Round) ApplyPasses() error {
	if r.phase != PhasePassing {
		return errWrongPhase("apply passes", "passing")
	}
	for i := range r.passSubmitted {
		if r.passSubmitted[i] == nil {
			return ErrNotAllPassesSubmitted
		}
	}

	var passes [PlayersPerTable][]Card
	copy(passes[:], r.passSubmitted[:])
	received := ExchangePasses(passes, r.passDir)

	for seat := range r.hands {
		for _, card := range r.passSubmitted[seat] {
			r.hands[seat], _ = RemoveCard(r.hands[seat], card)
		}
		r.passReceived[seat] = append([]Card(nil), received[seat]...)
		r.hands[seat] = append(r.hands[seat], received[seat]...)
		SortCards(r.hands[seat])
	}

	r.phase = PhasePassReview
	return nil
}

// MarkReady marks a seat as ready after reviewing passed cards.
func (r *Round) MarkReady(seat int) error {
	if r.phase != PhasePassReview {
		return errWrongPhase("mark ready", "pass review")
	}
	r.passReady[seat] = true
	return nil
}

// --- Play phase transitions ---

// StartPlaying transitions from PhasePassing (hold rounds) or PhasePassReview
// to PhasePlaying. It finds the Two of Clubs holder and sets the turn.
func (r *Round) StartPlaying() error {
	if r.phase != PhasePassing && r.phase != PhasePassReview {
		return errWrongPhase("start playing", "passing or pass review")
	}
	r.phase = PhasePlaying
	r.trickNumber = 0
	r.heartsBroken = false
	r.currentTrick = nil
	r.playedCards = nil

	twoClubs := Card{Suit: SuitClubs, Rank: 2}
	for seat, hand := range r.hands {
		if ContainsCard(hand, twoClubs) {
			r.turnSeat = seat
			break
		}
	}
	return nil
}

// --- Play phase mutations ---

// Play validates and applies a card play for the given seat.
// Returns a non-nil TrickResult when the trick completes (4 cards played).
// Advances turn, tracks hearts broken, and transitions to PhaseComplete after trick 12.
func (r *Round) Play(seat int, card Card) (*TrickResult, error) {
	if r.phase != PhasePlaying {
		return nil, errWrongPhase("play", "playing")
	}
	if seat != r.turnSeat {
		return nil, ErrNotYourTurn
	}

	if err := ValidatePlay(ValidatePlayInput{
		Hand:         r.hands[seat],
		Card:         card,
		Trick:        r.currentTrickCards(),
		HeartsBroken: r.heartsBroken,
		FirstTrick:   r.trickNumber == 0,
	}); err != nil {
		return nil, err
	}

	r.hands[seat], _ = RemoveCard(r.hands[seat], card)
	if card.Suit == SuitHearts {
		r.heartsBroken = true
	}
	r.currentTrick = append(r.currentTrick, Play{Seat: seat, Card: card})

	if len(r.currentTrick) < PlayersPerTable {
		r.turnSeat = (r.turnSeat + 1) % PlayersPerTable
		return nil, nil
	}

	// Trick complete — find winner and assign points.
	winnerSeat, points, _ := TrickWinner(r.currentTrick)
	r.roundPoints[winnerSeat] += points

	result := &TrickResult{
		WinnerSeat:  winnerSeat,
		Points:      points,
		TrickNumber: r.trickNumber,
	}

	if len(r.hands[winnerSeat]) == 0 {
		r.phase = PhaseComplete
		return result, nil
	}

	// Advance to next trick.
	for _, p := range r.currentTrick {
		r.playedCards = append(r.playedCards, p.Card)
	}
	r.currentTrick = nil
	r.trickNumber++
	r.turnSeat = winnerSeat

	return result, nil
}

// --- Scoring ---

// Scores returns raw and shoot-the-moon-adjusted points for the completed round.
func (r *Round) Scores() RoundScores {
	return RoundScores{
		Raw:      r.roundPoints,
		Adjusted: ApplyShootTheMoon(r.roundPoints),
	}
}

// --- Internal helpers ---

func (r *Round) currentTrickCards() []Card {
	cards := make([]Card, len(r.currentTrick))
	for i, p := range r.currentTrick {
		cards[i] = p.Card
	}
	return cards
}

func errWrongPhase(action, expected string) error {
	return fmt.Errorf("cannot %s: round is not in %s phase: %w", action, expected, ErrWrongPhase)
}

func errNotInHand(card Card) error {
	return fmt.Errorf("card %s is not in hand: %w", card, ErrCardNotInHand)
}
