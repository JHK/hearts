package bot

import (
	"fmt"
	"strings"
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

func (k StrategyKind) New() Strategy {
	switch k {
	case StrategySmart:
		return NewSmartBot()
	case StrategyDumb:
		return NewDumbBot()
	case StrategyFirstLegal:
		return NewFirstLegalBot()
	case StrategyRandom:
		fallthrough
	default:
		return NewRandomBot(nil)
	}
}
