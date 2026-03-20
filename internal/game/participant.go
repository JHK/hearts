package game

// Participant is the interface for any entity that occupies a seat in a Hearts game.
// Both human players (*Player) and bots satisfy this interface.
type Participant interface {
	Hand() []Card
	RoundPoints() Points
	CumulativePoints() Points
	PassSent() []Card
	PassReceived() []Card
	PassReady() bool

	DealCards([]Card)
	SubmitPass([]Card)
	HasSubmittedPass() bool
	SendPassCards()
	ReceivePassCards([]Card)
	MarkPassReady()
	PlayCard(Card) bool
	AddTrickPoints(Points)
	FinalizeRound(Points)
}
