package bot

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
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
				require.ErrorIs(t, err, ErrUnknownStrategy)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestStrategyKindNew(t *testing.T) {
	bot := StrategySmart.New()
	require.IsType(t, &Smart{}, bot)

	bot = StrategyRandom.New()
	require.IsType(t, &Random{}, bot)

	bot = StrategyFirstLegal.New()
	require.IsType(t, &FirstLegal{}, bot)
}

func TestBotNameAvoidsCollisions(t *testing.T) {
	taken := map[string]bool{}

	// Add 4 bots of the same strategy — all names must be unique.
	for range 4 {
		name := StrategyRandom.BotName(taken)
		require.False(t, taken[name], "duplicate bot name: %s", name)
		taken[name] = true
	}
}

func TestBotNameFallsBackToSuffix(t *testing.T) {
	// FirstLegal has only one name ("Fritz"). Exhaust the pool.
	taken := map[string]bool{"Fritz": true}
	name := StrategyFirstLegal.BotName(taken)
	require.NotEqual(t, "Fritz", name)
	require.True(t, strings.HasPrefix(name, "Fritz"), "expected suffixed Fritz variant, got %s", name)
}

func TestBotNameAvoidsHumanNames(t *testing.T) {
	taken := map[string]bool{"Lucky": true}
	for range 20 {
		name := StrategyRandom.BotName(taken)
		require.NotEqual(t, "Lucky", name, "bot name collided with human name")
	}
}
