package bot

import (
	"fmt"
	"strings"
)

type StrategyKind string

const (
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
		return "", fmt.Errorf("unknown strategy %q (available: %s, %s)", raw, StrategyRandom, StrategyFirstLegal)
	}

	return kind, nil
}

func (k StrategyKind) Valid() bool {
	switch k {
	case StrategyRandom, StrategyFirstLegal:
		return true
	default:
		return false
	}
}

func (k StrategyKind) New() Strategy {
	switch k {
	case StrategyFirstLegal:
		return NewFirstLegalBot()
	case StrategyRandom:
		fallthrough
	default:
		return NewRandomBot(nil)
	}
}
