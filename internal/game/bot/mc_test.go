package bot

import (
	"math/rand"
	"testing"

	"github.com/JHK/hearts/internal/game"
)

func TestDetectSeatVoids(t *testing.T) {
	// Trick: seat 0 leads 5C, seat 1 follows 3C, seat 2 plays 7H (void in C), seat 3 follows 9C.
	plays := []game.Play{
		{Seat: 0, Card: c("5C")},
		{Seat: 1, Card: c("3C")},
		{Seat: 2, Card: c("7H")},
		{Seat: 3, Card: c("9C")},
	}

	voids := detectSeatVoids(plays, nil)

	if voids[2] == nil || !voids[2][game.SuitClubs] {
		t.Error("seat 2 should be void in clubs")
	}
	if voids[0] != nil {
		t.Error("seat 0 should have no voids")
	}
	if voids[1] != nil {
		t.Error("seat 1 should have no voids")
	}
}

func TestDetectSeatVoids_MultipleTricks(t *testing.T) {
	// Trick 1: seat 0 leads clubs, seat 2 discards a heart (void in clubs).
	// Trick 2: seat 3 leads diamonds, seat 1 discards a spade (void in diamonds).
	plays := []game.Play{
		{Seat: 0, Card: c("5C")}, {Seat: 1, Card: c("3C")}, {Seat: 2, Card: c("7H")}, {Seat: 3, Card: c("9C")},
		{Seat: 3, Card: c("4D")}, {Seat: 0, Card: c("8D")}, {Seat: 1, Card: c("2S")}, {Seat: 2, Card: c("TD")},
	}

	voids := detectSeatVoids(plays, nil)

	if voids[2] == nil || !voids[2][game.SuitClubs] {
		t.Error("seat 2 should be void in clubs")
	}
	if voids[1] == nil || !voids[1][game.SuitDiamonds] {
		t.Error("seat 1 should be void in diamonds")
	}
	// Seats 0 and 3 followed suit in both tricks — no voids.
	if voids[0] != nil {
		t.Error("seat 0 should have no voids")
	}
	if voids[3] != nil {
		t.Error("seat 3 should have no voids")
	}
}

func TestDetectSeatVoids_CurrentTrick(t *testing.T) {
	currentTrick := []game.Play{
		{Seat: 1, Card: c("KD")},
		{Seat: 2, Card: c("QS")}, // void in diamonds
	}

	voids := detectSeatVoids(nil, currentTrick)

	if voids[2] == nil || !voids[2][game.SuitDiamonds] {
		t.Error("seat 2 should be void in diamonds")
	}
}

func TestSampleOpponentHands_BasicConstraints(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	mySeat := 0
	remaining := cards("2C", "3C", "4C", "5C", "6C", "7C", "8C", "9C", "TC")
	handSizes := [4]int{0, 3, 3, 3}
	var noVoids [4]map[game.Suit]bool

	hands, ok := sampleOpponentHands(rng, mySeat, remaining, handSizes, noVoids)
	if !ok {
		t.Fatal("sampling should succeed without void constraints")
	}

	if len(hands[0]) != 0 {
		t.Errorf("my seat should have no cards, got %d", len(hands[0]))
	}
	for seat := 1; seat <= 3; seat++ {
		if len(hands[seat]) != 3 {
			t.Errorf("seat %d should have 3 cards, got %d", seat, len(hands[seat]))
		}
	}

	// All remaining cards should be distributed.
	var all []game.Card
	for _, h := range hands {
		all = append(all, h...)
	}
	if len(all) != len(remaining) {
		t.Errorf("expected %d total cards, got %d", len(remaining), len(all))
	}
}

func TestSampleOpponentHands_RespectsVoids(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	mySeat := 0
	remaining := cards("2C", "3C", "4D", "5D", "6H", "7H")
	handSizes := [4]int{0, 2, 2, 2}

	var voids [4]map[game.Suit]bool
	voids[1] = map[game.Suit]bool{game.SuitClubs: true} // seat 1 void in clubs

	for range 20 {
		hands, ok := sampleOpponentHands(rng, mySeat, remaining, handSizes, voids)
		if !ok {
			t.Fatal("sampling should succeed")
		}
		for _, card := range hands[1] {
			if card.Suit == game.SuitClubs {
				t.Fatalf("seat 1 should not hold clubs, got %s", card)
			}
		}
	}
}

func TestSampleOpponentHands_VoidFallback(t *testing.T) {
	// Over-constrained: 3 clubs to distribute but all 3 opponents void in clubs.
	// Primary sampling must fail; fallback (no voids) should succeed.
	rng := rand.New(rand.NewSource(42))
	mySeat := 0
	remaining := cards("2C", "3C", "4C")
	handSizes := [4]int{0, 1, 1, 1}

	var voids [4]map[game.Suit]bool
	voids[1] = map[game.Suit]bool{game.SuitClubs: true}
	voids[2] = map[game.Suit]bool{game.SuitClubs: true}
	voids[3] = map[game.Suit]bool{game.SuitClubs: true}

	// Direct call should fail.
	_, ok := sampleOpponentHands(rng, mySeat, remaining, handSizes, voids)
	if ok {
		t.Fatal("sampling should fail when all opponents are void in the only suit")
	}

	// Fallback with no voids should succeed.
	var noVoids [4]map[game.Suit]bool
	hands, ok := sampleOpponentHands(rng, mySeat, remaining, handSizes, noVoids)
	if !ok {
		t.Fatal("fallback sampling without voids should succeed")
	}
	total := 0
	for _, h := range hands {
		total += len(h)
	}
	if total != 3 {
		t.Errorf("expected 3 total cards, got %d", total)
	}
}

func TestSimulateRemaining_CompletesGame(t *testing.T) {
	// Set up a simple state: 2 tricks remaining, all players have 2 cards.
	state := mcState{
		hands: [4][]game.Card{
			cards("AC", "AD"),
			cards("2C", "2D"),
			cards("3C", "3D"),
			cards("4C", "4D"),
		},
		heartsBroken: false,
		trickNumber:  11,
		turnSeat:     0,
		firstTrick:   false,
	}

	final := simulateRemaining(state)

	// No hearts in play, so all round points should be 0.
	for seat, pts := range final {
		if pts != 0 {
			t.Errorf("seat %d should have 0 points, got %d", seat, pts)
		}
	}
}

func TestSimulateRemaining_ScoresHearts(t *testing.T) {
	// Seat 0 leads AH, others follow with low hearts. Then seat 0 leads AD.
	state := mcState{
		hands: [4][]game.Card{
			cards("AH", "AD"),
			cards("2H", "2D"),
			cards("3H", "3D"),
			cards("4H", "4D"),
		},
		heartsBroken: true,
		trickNumber:  11,
		turnSeat:     0,
		firstTrick:   false,
	}

	final := simulateRemaining(state)

	// Seat 0 wins both tricks. First trick has 4 hearts = 4 points.
	if final[0] < 4 {
		t.Errorf("seat 0 should have at least 4 points from hearts trick, got %d", final[0])
	}
}

func TestMCEvaluator_ReturnsLegalPlay(t *testing.T) {
	mc := &mcEvaluator{
		rng:     rand.New(rand.NewSource(42)),
		samples: 20,
	}

	input := game.TurnInput{
		Hand:         cards("KS", "QH", "3D", "7C"),
		Trick:        []game.Play{{Seat: 1, Card: c("5D")}},
		HeartsBroken: false,
		FirstTrick:   false,
		PlayedCards:  nil,
		MySeat:       0,
	}

	legal := game.LegalPlays(input.Hand, input.TrickCards(), input.HeartsBroken, input.FirstTrick)
	card := mc.evaluate(input, legal)

	if !game.ContainsCard(legal, card) {
		t.Errorf("MC returned illegal card %s", card)
	}
}

func TestMCEvaluator_SingleLegalPlay(t *testing.T) {
	mc := &mcEvaluator{
		rng:     rand.New(rand.NewSource(42)),
		samples: 10,
	}

	legal := cards("2C")
	input := game.TurnInput{
		Hand:       cards("2C", "KS", "QH"),
		FirstTrick: true,
		MySeat:     0,
	}

	card := mc.evaluate(input, legal)
	if card != c("2C") {
		t.Errorf("expected 2C with single legal play, got %s", card)
	}
}

func TestMCEvaluator_AvoidsQueenSpades(t *testing.T) {
	// Mid-game: we must follow spades. We have QS and 3S.
	// MC should prefer 3S to avoid winning QS.
	// Fixed seed is load-bearing: keeps this statistical test deterministic.
	mc := &mcEvaluator{
		rng:     rand.New(rand.NewSource(42)),
		samples: 50,
	}

	input := game.TurnInput{
		Hand: cards("QS", "3S", "4H", "5H", "6H"),
		Trick: []game.Play{
			{Seat: 1, Card: c("JS")},
		},
		HeartsBroken: true,
		FirstTrick:   false,
		MySeat:       0,
	}

	legal := game.LegalPlays(input.Hand, input.TrickCards(), input.HeartsBroken, input.FirstTrick)
	card := mc.evaluate(input, legal)

	if card == c("QS") {
		t.Error("MC should avoid playing QS when 3S is available and JS is leading")
	}
}

func cards(strs ...string) []game.Card {
	result := make([]game.Card, len(strs))
	for i, s := range strs {
		result[i] = c(s)
	}
	return result
}
