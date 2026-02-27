package bot

import "testing"

func TestParseStrategyKind(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    StrategyKind
		wantErr bool
	}{
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
	strategy := StrategyRandom.New()
	if _, ok := strategy.(*Random); !ok {
		t.Fatalf("expected Random strategy")
	}

	strategy = StrategyFirstLegal.New()
	if _, ok := strategy.(*FirstLegal); !ok {
		t.Fatalf("expected FirstLegal strategy")
	}
}
