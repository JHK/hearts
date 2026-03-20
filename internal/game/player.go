package game

// Player holds all per-player game state for a single round and across rounds.
type Player struct {
	hand             []Card
	roundPoints      Points
	cumulativePoints Points
	passSent         []Card // nil = not yet submitted; non-nil = submitted
	passReceived     []Card
	passReady        bool
}

// NewPlayer creates a Player.
func NewPlayer() *Player {
	return &Player{}
}

// NewPlayerState creates a Player with preset cumulative and round points.
// Intended for test setup only.
func NewPlayerState(roundPoints, cumulativePoints Points) *Player {
	return &Player{roundPoints: roundPoints, cumulativePoints: cumulativePoints}
}

// Hand returns the player's current hand.
func (p *Player) Hand() []Card { return p.hand }

// RoundPoints returns the points accumulated this round.
func (p *Player) RoundPoints() Points { return p.roundPoints }

// CumulativePoints returns the total points across all rounds.
func (p *Player) CumulativePoints() Points { return p.cumulativePoints }

// PassSent returns the cards this player submitted to pass (nil = not yet submitted).
func (p *Player) PassSent() []Card { return p.passSent }

// PassReceived returns the cards this player received from the pass.
func (p *Player) PassReceived() []Card { return p.passReceived }

// PassReady reports whether the player has confirmed they've reviewed their passed cards.
func (p *Player) PassReady() bool { return p.passReady }

// DealCards sets the hand and resets all round-level state.
func (p *Player) DealCards(cards []Card) {
	p.hand = append(p.hand[:0], cards...)
	p.roundPoints = 0
	p.passSent = nil
	p.passReceived = nil
	p.passReady = false
}

// SubmitPass records the cards this player intends to pass.
func (p *Player) SubmitPass(cards []Card) {
	p.passSent = append([]Card(nil), cards...)
}

// HasSubmittedPass reports whether the player has submitted a pass this round.
func (p *Player) HasSubmittedPass() bool {
	return p.passSent != nil
}

// SendPassCards removes the submitted pass cards from the player's hand.
func (p *Player) SendPassCards() {
	for _, card := range p.passSent {
		p.hand, _ = RemoveCard(p.hand, card)
	}
}

// ReceivePassCards adds the received cards to the player's hand and records them.
func (p *Player) ReceivePassCards(cards []Card) {
	p.passReceived = append([]Card(nil), cards...)
	p.hand = append(p.hand, cards...)
	SortCards(p.hand)
}

// MarkPassReady marks the player as ready to proceed after reviewing passed cards.
func (p *Player) MarkPassReady() {
	p.passReady = true
}

// PlayCard removes the card from the player's hand and reports whether it was found.
func (p *Player) PlayCard(card Card) bool {
	updated, removed := RemoveCard(p.hand, card)
	if !removed {
		return false
	}
	p.hand = updated
	return true
}

// AddTrickPoints adds points won from a trick to the player's round total.
func (p *Player) AddTrickPoints(pts Points) {
	p.roundPoints += pts
}

// FinalizeRound adds the given adjusted points to the cumulative total and resets round state.
// The caller is responsible for applying shoot-the-moon adjustments before calling this.
func (p *Player) FinalizeRound(adjustedPoints Points) {
	p.cumulativePoints += adjustedPoints
	p.roundPoints = 0
	p.passSent = nil
	p.passReceived = nil
	p.passReady = false
}
