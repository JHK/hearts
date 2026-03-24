package game

// GameOverThreshold is the cumulative point total that triggers the end of a game.
const GameOverThreshold Points = 100

// GameResult holds the final state when a game ends.
type GameResult struct {
	Scores  [PlayersPerTable]Points
	Winners []int // seat indices of the winner(s) (lowest score; ties possible)
}

// Game tracks cumulative scores across rounds and detects game-over conditions.
type Game struct {
	scores      [PlayersPerTable]Points
	roundsPlayed int
}

// NewGame creates a new game tracker with zero scores.
func NewGame() *Game {
	return &Game{}
}

// RoundsPlayed returns the number of completed rounds.
func (g *Game) RoundsPlayed() int { return g.roundsPlayed }

// Scores returns the current cumulative scores.
func (g *Game) Scores() [PlayersPerTable]Points { return g.scores }

// Score returns the cumulative score for a single seat.
func (g *Game) Score(seat int) Points { return g.scores[seat] }

// AddRoundScores adds the adjusted scores from a completed round and returns
// true if the game is now over (any player reached the threshold).
func (g *Game) AddRoundScores(adjusted [PlayersPerTable]Points) bool {
	for i, pts := range adjusted {
		g.scores[i] += pts
	}
	g.roundsPlayed++
	return g.IsOver()
}

// IsOver reports whether any player has reached the game-over threshold.
func (g *Game) IsOver() bool {
	for _, pts := range g.scores {
		if pts >= GameOverThreshold {
			return true
		}
	}
	return false
}

// Winners returns the seat indices of the player(s) with the lowest cumulative score.
func (g *Game) Winners() []int {
	min := g.scores[0]
	for _, pts := range g.scores[1:] {
		if pts < min {
			min = pts
		}
	}
	var out []int
	for i, pts := range g.scores {
		if pts == min {
			out = append(out, i)
		}
	}
	return out
}

// Result returns the final game result. Only meaningful when IsOver() is true.
func (g *Game) Result() GameResult {
	return GameResult{
		Scores:  g.scores,
		Winners: g.Winners(),
	}
}

// NextPassDirection returns the pass direction for the next round.
func (g *Game) NextPassDirection() PassDirection {
	return PassDirectionForRound(g.roundsPlayed)
}
