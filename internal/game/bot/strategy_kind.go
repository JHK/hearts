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

// BotName returns a randomly chosen name for a bot of this strategy that
// does not collide with any of the provided taken names. If every name in
// the pool is taken, a numeric suffix is appended to make the name unique.
func (k StrategyKind) BotName(taken map[string]bool) string {
	pool := botNames[k]
	if len(pool) == 0 {
		return uniqueName("Bot", taken)
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Shuffle and pick the first available name.
	perm := rng.Perm(len(pool))
	for _, i := range perm {
		if !taken[pool[i]] {
			return pool[i]
		}
	}

	// All pool names taken — append a suffix to a random one.
	base := pool[perm[0]]
	return uniqueName(base, taken)
}

func uniqueName(base string, taken map[string]bool) string {
	if !taken[base] {
		return base
	}
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s %d", base, i)
		if !taken[candidate] {
			return candidate
		}
	}
}
