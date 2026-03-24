package bot

import "github.com/JHK/hearts/internal/game"

// Bot makes autonomous decisions during play.
// Bots hold no player state — game state is managed by game.Round.
type Bot interface {
	ChoosePlay(game.TurnInput) (game.Card, error)
	ChoosePass(game.PassInput) ([]game.Card, error)
	Kind() StrategyKind
}
