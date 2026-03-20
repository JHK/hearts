package bot

import (
	"fmt"
	"strings"

	"github.com/JHK/hearts/internal/game"
)

type StrategyKind string

const (
	StrategySmart      StrategyKind = "smart"
	StrategyDumb       StrategyKind = "dumb"
	StrategyRandom     StrategyKind = "random"
	StrategyFirstLegal StrategyKind = "first-legal"
)

func ParseStrategyKind(raw string) (StrategyKind, error) {
	name := strings.ToLower(strings.TrimSpace(raw))
	if name == "" {
		return StrategyRandom, nil
	}

	kind := StrategyKind(name)
	if !kind.Valid() {
		return "", fmt.Errorf("unknown strategy %q (available: %s, %s, %s, %s)", raw, StrategySmart, StrategyDumb, StrategyRandom, StrategyFirstLegal)
	}

	return kind, nil
}

func (k StrategyKind) Valid() bool {
	switch k {
	case StrategySmart, StrategyDumb, StrategyRandom, StrategyFirstLegal:
		return true
	default:
		return false
	}
}

// NewBot creates a fresh bot of this strategy kind.
func (k StrategyKind) NewBot() Bot {
	return k.WrapPlayer(game.NewPlayer())
}

// WrapPlayer wraps an existing *game.Player with this strategy kind's bot logic.
// Use this when converting a human player to a bot mid-round to preserve game state.
func (k StrategyKind) WrapPlayer(p *game.Player) Bot {
	switch k {
	case StrategySmart:
		return &Smart{Player: p}
	case StrategyDumb:
		return &Dumb{Player: p}
	case StrategyFirstLegal:
		return &FirstLegal{Player: p}
	case StrategyRandom:
		fallthrough
	default:
		return newRandomBot(p, nil)
	}
}

// New returns a new bot of this strategy kind.
func (k StrategyKind) New() Bot {
	return k.NewBot()
}
