package bot

import (
	"strings"
	"testing"
)

func TestParseStrategyKind(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    StrategyKind
		wantErr bool
	}{
		{name: "smart", raw: "smart", want: StrategySmart},
		{name: "empty defaults to random", raw: "", want: StrategyRandom},
		{name: "random", raw: "random", want: StrategyRandom},
		{name: "first legal", raw: "first-legal", want: StrategyFirstLegal},
		{name: "unknown", raw: "first", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseStrategyKind(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %s, got %s", tt.want, got)
			}
		})
	}
}

func TestStrategyKindNew(t *testing.T) {
	bot := StrategySmart.New()
	if _, ok := bot.(*Smart); !ok {
		t.Fatalf("expected Smart bot")
	}

	bot = StrategyRandom.New()
	if _, ok := bot.(*Random); !ok {
		t.Fatalf("expected Random bot")
	}

	bot = StrategyFirstLegal.New()
	if _, ok := bot.(*FirstLegal); !ok {
		t.Fatalf("expected FirstLegal bot")
	}
}

func TestBotNameAvoidsCollisions(t *testing.T) {
	taken := map[string]bool{}

	// Add 4 bots of the same strategy — all names must be unique.
	for range 4 {
		name := StrategyRandom.BotName(taken)
		if taken[name] {
			t.Fatalf("duplicate bot name: %s", name)
		}
		taken[name] = true
	}
}

func TestBotNameFallsBackToSuffix(t *testing.T) {
	// FirstLegal has only one name ("Fritz"). Exhaust the pool.
	taken := map[string]bool{"Fritz": true}
	name := StrategyFirstLegal.BotName(taken)
	if name == "Fritz" {
		t.Fatal("expected a different name when Fritz is taken")
	}
	if !strings.HasPrefix(name, "Fritz") {
		t.Fatalf("expected suffixed Fritz variant, got %s", name)
	}
}

func TestBotNameAvoidsHumanNames(t *testing.T) {
	taken := map[string]bool{"Lucky": true}
	for range 20 {
		name := StrategyRandom.BotName(taken)
		if name == "Lucky" {
			t.Fatal("bot name collided with human name")
		}
	}
}
