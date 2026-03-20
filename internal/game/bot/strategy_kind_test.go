package bot

import "testing"

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

func TestStrategyKindWrapPlayerPreservesState(t *testing.T) {
	b := StrategyDumb.NewBot()
	b.DealCards(mustParseCards(t, []string{"AC", "KH"}))
	b.AddTrickPoints(5)

	wrapped := StrategyRandom.WrapPlayer(b.Unwrap())
	if len(wrapped.Hand()) != 2 {
		t.Fatalf("expected 2 cards in hand, got %d", len(wrapped.Hand()))
	}
	if wrapped.RoundPoints() != 5 {
		t.Fatalf("expected round points 5, got %d", wrapped.RoundPoints())
	}
}
