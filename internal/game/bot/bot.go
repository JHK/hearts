package bot

import "github.com/JHK/hearts/internal/game"

// TurnInput is the information passed to a bot when it must choose a card to play.
type TurnInput struct {
	Hand         []game.Card
	Trick        []game.Card
	HeartsBroken bool
	FirstTrick   bool
	PlayedCards  []game.Card // all cards played in completed tricks this round
}

// PassInput is the information passed to a bot when it must choose cards to pass.
type PassInput struct {
	Hand      []game.Card
	Direction game.PassDirection
}

// Bot is a game participant that makes autonomous decisions during play.
// It extends game.Participant with strategy methods and the ability to hand back
// its underlying *game.Player (for human reconnection).
type Bot interface {
	game.Participant
	ChoosePlay(TurnInput) (game.Card, error)
	ChoosePass(PassInput) ([]game.Card, error)
	BotName() string
	Kind() StrategyKind
	Unwrap() *game.Player
}
