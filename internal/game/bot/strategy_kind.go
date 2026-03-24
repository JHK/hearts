package bot

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
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
	switch k {
	case StrategySmart:
		return &Smart{}
	case StrategyDumb:
		return &Dumb{}
	case StrategyFirstLegal:
		return &FirstLegal{}
	case StrategyRandom:
		fallthrough
	default:
		return newRandomBot(nil)
	}
}

// New returns a new bot of this strategy kind.
func (k StrategyKind) New() Bot {
	return k.NewBot()
}

var botNames = map[StrategyKind][]string{
	StrategySmart:      smartBotNames,
	StrategyDumb:       dumbBotNames,
	StrategyRandom:     randomBotNames,
	StrategyFirstLegal: firstLegalBotNames,
}

// BotName returns a randomly chosen name for a bot of this strategy.
func (k StrategyKind) BotName() string {
	pool := botNames[k]
	if len(pool) == 0 {
		return "Bot"
	}
	return pool[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(pool))]
}
