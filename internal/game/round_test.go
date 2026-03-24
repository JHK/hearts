package game

import "testing"

func TestRoundDealAndHand(t *testing.T) {
	hand0 := []Card{{Suit: SuitClubs, Rank: 2}, {Suit: SuitHearts, Rank: 5}}
	hand1 := []Card{{Suit: SuitDiamonds, Rank: 3}}
	hand2 := []Card{{Suit: SuitSpades, Rank: 7}}
	hand3 := []Card{{Suit: SuitHearts, Rank: 10}}

	r := NewRound([PlayersPerTable][]Card{hand0, hand1, hand2, hand3}, PassDirectionLeft)

	if len(r.Hand(0)) != 2 {
		t.Fatalf("expected 2 cards in seat 0, got %d", len(r.Hand(0)))
	}
	if len(r.Hand(1)) != 1 {
		t.Fatalf("expected 1 card in seat 1, got %d", len(r.Hand(1)))
	}
}

func TestRoundSubmitAndApplyPasses(t *testing.T) {
	hands := [PlayersPerTable][]Card{
		{{Suit: SuitClubs, Rank: 2}, {Suit: SuitClubs, Rank: 3}, {Suit: SuitClubs, Rank: 4}, {Suit: SuitHearts, Rank: 5}},
		{{Suit: SuitDiamonds, Rank: 2}, {Suit: SuitDiamonds, Rank: 3}, {Suit: SuitDiamonds, Rank: 4}, {Suit: SuitSpades, Rank: 5}},
		{{Suit: SuitHearts, Rank: 2}, {Suit: SuitHearts, Rank: 3}, {Suit: SuitHearts, Rank: 4}, {Suit: SuitClubs, Rank: 5}},
		{{Suit: SuitSpades, Rank: 2}, {Suit: SuitSpades, Rank: 3}, {Suit: SuitSpades, Rank: 4}, {Suit: SuitDiamonds, Rank: 5}},
	}
	r := NewRound(hands, PassDirectionLeft)

	for seat := 0; seat < PlayersPerTable; seat++ {
		if err := r.SubmitPass(seat, r.Hand(seat)[:3]); err != nil {
			t.Fatalf("submit pass seat %d: %v", seat, err)
		}
	}

	if !r.HasSubmittedPass(0) {
		t.Fatal("expected HasSubmittedPass true after SubmitPass")
	}

	if err := r.ApplyPasses(); err != nil {
		t.Fatalf("apply passes: %v", err)
	}

	if r.Phase() != PhasePassReview {
		t.Fatalf("expected PhasePassReview, got %d", r.Phase())
	}

	// After left pass, seat 0 sent to seat 1; seat 0 received from seat 3.
	if len(r.PassReceived(0)) != 3 {
		t.Fatalf("expected 3 received cards for seat 0, got %d", len(r.PassReceived(0)))
	}
	if len(r.Hand(0)) != 4 {
		t.Fatalf("expected 4 cards in hand after pass, got %d", len(r.Hand(0)))
	}
}

func TestRoundMarkReadyAndStartPlaying(t *testing.T) {
	hands := dealTestHands()
	r := NewRound(hands, PassDirectionHold)

	if err := r.StartPlaying(); err != nil {
		t.Fatalf("start playing: %v", err)
	}
	if r.Phase() != PhasePlaying {
		t.Fatalf("expected PhasePlaying, got %d", r.Phase())
	}
}

func TestRoundPlayValidation(t *testing.T) {
	hands := dealTestHands()
	r := NewRound(hands, PassDirectionHold)
	_ = r.StartPlaying()

	// Wrong seat should fail.
	wrongSeat := (r.TurnSeat() + 1) % PlayersPerTable
	_, err := r.Play(wrongSeat, r.Hand(wrongSeat)[0])
	if err == nil {
		t.Fatal("expected error for wrong seat play")
	}
}

func TestRoundPlayTrickCompletion(t *testing.T) {
	// Build hands where seat 0 has 2C (leads first trick).
	hands := [PlayersPerTable][]Card{
		{{Suit: SuitClubs, Rank: 2}},
		{{Suit: SuitClubs, Rank: 3}},
		{{Suit: SuitClubs, Rank: 4}},
		{{Suit: SuitClubs, Rank: 5}},
	}
	r := NewRound(hands, PassDirectionHold)
	_ = r.StartPlaying()

	if r.TurnSeat() != 0 {
		t.Fatalf("expected seat 0 to lead (has 2C), got %d", r.TurnSeat())
	}

	// Play all 4 cards.
	for seat := 0; seat < PlayersPerTable; seat++ {
		result, err := r.Play(r.TurnSeat(), r.Hand(r.TurnSeat())[0])
		if err != nil {
			t.Fatalf("play seat %d: %v", seat, err)
		}
		if seat < 3 && result != nil {
			t.Fatalf("expected nil result for card %d", seat+1)
		}
		if seat == 3 && result == nil {
			t.Fatal("expected trick result after 4th card")
		}
		if seat == 3 {
			if result.WinnerSeat != 3 {
				t.Fatalf("expected seat 3 (5C) to win, got %d", result.WinnerSeat)
			}
			if result.Points != 0 {
				t.Fatalf("expected 0 points (all clubs), got %d", result.Points)
			}
		}
	}

	if r.Phase() != PhaseComplete {
		t.Fatalf("expected PhaseComplete after last trick, got %d", r.Phase())
	}
}

func TestRoundScores(t *testing.T) {
	r := NewTestRound([PlayersPerTable]Points{5, 7, 3, 11})
	scores := r.Scores()

	if scores.Raw[0] != 5 || scores.Raw[3] != 11 {
		t.Fatalf("unexpected raw scores: %v", scores.Raw)
	}
	if scores.Adjusted[0] != 5 {
		t.Fatalf("expected adjusted[0]=5 (no shoot the moon), got %d", scores.Adjusted[0])
	}
}

func TestRoundScoresShootTheMoon(t *testing.T) {
	r := NewTestRound([PlayersPerTable]Points{26, 0, 0, 0})
	scores := r.Scores()

	if scores.Adjusted[0] != 0 || scores.Adjusted[1] != 26 {
		t.Fatalf("unexpected adjusted scores: %v", scores.Adjusted)
	}
}

func dealTestHands() [PlayersPerTable][]Card {
	return [PlayersPerTable][]Card{
		{{Suit: SuitClubs, Rank: 2}},
		{{Suit: SuitClubs, Rank: 3}},
		{{Suit: SuitClubs, Rank: 4}},
		{{Suit: SuitClubs, Rank: 5}},
	}
}
