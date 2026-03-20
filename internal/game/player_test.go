package game

import "testing"

func TestPlayerDealCardsResetsRoundState(t *testing.T) {
	p := NewPlayer()
	p.roundPoints = 5
	p.passSent = []Card{{Suit: SuitClubs, Rank: 2}}
	p.passReady = true

	cards := []Card{{Suit: SuitClubs, Rank: 2}, {Suit: SuitHearts, Rank: 5}}
	p.DealCards(cards)

	if len(p.hand) != 2 {
		t.Fatalf("expected 2 cards, got %d", len(p.hand))
	}
	if p.roundPoints != 0 {
		t.Fatalf("expected roundPoints reset to 0, got %d", p.roundPoints)
	}
	if p.passSent != nil {
		t.Fatalf("expected passSent reset to nil")
	}
	if p.passReady {
		t.Fatalf("expected passReady reset to false")
	}
}

func TestPlayerSubmitAndSendReceivePass(t *testing.T) {
	p := NewPlayer()
	hand := []Card{
		{Suit: SuitClubs, Rank: 2},
		{Suit: SuitClubs, Rank: 3},
		{Suit: SuitClubs, Rank: 4},
		{Suit: SuitHearts, Rank: 5},
	}
	p.DealCards(hand)

	toPass := []Card{hand[0], hand[1], hand[2]}
	p.SubmitPass(toPass)

	if !p.HasSubmittedPass() {
		t.Fatal("expected HasSubmittedPass to be true after SubmitPass")
	}
	if len(p.passSent) != 3 {
		t.Fatalf("expected 3 passSent, got %d", len(p.passSent))
	}

	p.SendPassCards()
	if len(p.hand) != 1 {
		t.Fatalf("expected 1 card remaining after send, got %d", len(p.hand))
	}
	if p.hand[0] != (Card{Suit: SuitHearts, Rank: 5}) {
		t.Fatalf("unexpected remaining card: %v", p.hand[0])
	}

	received := []Card{{Suit: SuitSpades, Rank: 10}, {Suit: SuitSpades, Rank: 11}, {Suit: SuitSpades, Rank: 12}}
	p.ReceivePassCards(received)
	if len(p.hand) != 4 {
		t.Fatalf("expected 4 cards after receive, got %d", len(p.hand))
	}
	if len(p.passReceived) != 3 {
		t.Fatalf("expected 3 passReceived, got %d", len(p.passReceived))
	}
}

func TestPlayerHasSubmittedPassFalseInitially(t *testing.T) {
	p := NewPlayer()
	p.DealCards([]Card{{Suit: SuitClubs, Rank: 2}})
	if p.HasSubmittedPass() {
		t.Fatal("expected HasSubmittedPass to be false before any submission")
	}
}

func TestPlayerPlayCard(t *testing.T) {
	p := NewPlayer()
	card := Card{Suit: SuitClubs, Rank: 2}
	p.DealCards([]Card{card, {Suit: SuitHearts, Rank: 5}})

	if !p.PlayCard(card) {
		t.Fatal("expected PlayCard to return true for card in hand")
	}
	if len(p.hand) != 1 {
		t.Fatalf("expected 1 card remaining, got %d", len(p.hand))
	}

	if p.PlayCard(card) {
		t.Fatal("expected PlayCard to return false for card not in hand")
	}
}

func TestPlayerAddTrickPoints(t *testing.T) {
	p := NewPlayer()
	p.AddTrickPoints(5)
	p.AddTrickPoints(13)
	if p.roundPoints != 18 {
		t.Fatalf("expected roundPoints 18, got %d", p.roundPoints)
	}
}

func TestPlayerFinalizeRound(t *testing.T) {
	p := NewPlayer()
	p.cumulativePoints = 10
	p.roundPoints = 7
	p.passSent = []Card{{Suit: SuitClubs, Rank: 2}}
	p.passReady = true

	p.FinalizeRound(7)

	if p.cumulativePoints != 17 {
		t.Fatalf("expected cumulativePoints 17, got %d", p.cumulativePoints)
	}
	if p.roundPoints != 0 {
		t.Fatalf("expected roundPoints reset to 0, got %d", p.roundPoints)
	}
	if p.passSent != nil {
		t.Fatalf("expected passSent reset to nil")
	}
	if p.passReady {
		t.Fatalf("expected passReady reset to false")
	}
}

func TestPlayerMarkPassReady(t *testing.T) {
	p := NewPlayer()
	if p.passReady {
		t.Fatal("expected passReady false initially")
	}
	p.MarkPassReady()
	if !p.passReady {
		t.Fatal("expected passReady true after MarkPassReady")
	}
}
