package bot

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type StrategyKind string

const (
	StrategyHard       StrategyKind = "hard"
	StrategyMedium     StrategyKind = "medium"
	StrategyEasy       StrategyKind = "easy"
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
		return "", fmt.Errorf("%w %q (available: %s, %s, %s, %s, %s)", ErrUnknownStrategy, raw, StrategyHard, StrategyMedium, StrategyEasy, StrategyRandom, StrategyFirstLegal)
	}

	return kind, nil
}

func (k StrategyKind) Valid() bool {
	switch k {
	case StrategyHard, StrategyMedium, StrategyEasy, StrategyRandom, StrategyFirstLegal:
		return true
	default:
		return false
	}
}

// BotOptions configures bot creation. Zero values use defaults.
type BotOptions struct {
	MCSamples int // Monte Carlo samples for Hard bot (0 = defaultMCSamples)
}

// NewBot creates a fresh bot of this strategy kind with default options.
func (k StrategyKind) NewBot() Bot {
	return k.NewBotWithOptions(BotOptions{})
}

// NewBotWithOptions creates a fresh bot of this strategy kind with the given options.
func (k StrategyKind) NewBotWithOptions(opts BotOptions) Bot {
	switch k {
	case StrategyHard:
		samples := opts.MCSamples
		if samples <= 0 {
			samples = defaultMCSamples
		}
		return &Hard{mc: newMCEvaluator(samples)}
	case StrategyMedium:
		return &Medium{}
	case StrategyEasy:
		return &Easy{}
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
	StrategyHard:       hardBotNames,
	StrategyMedium:     mediumBotNames,
	StrategyEasy:       easyBotNames,
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
