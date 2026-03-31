// Command analyze reads sampled game logs and identifies strategy weaknesses.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/JHK/hearts/internal/sim"
)

func main() {
	file := flag.String("file", "samples.json", "path to sampled game logs")
	flag.Parse()

	data, err := os.ReadFile(*file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %s: %v\n", *file, err)
		os.Exit(1)
	}

	var games []sim.GameLog
	if err := json.Unmarshal(data, &games); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Loaded %d games\n\n", len(games))

	// Find which strategy slot is "hard".
	hardSlot := -1
	for i, s := range games[0].Strategies {
		if s == "hard" {
			hardSlot = i
			break
		}
	}
	if hardSlot < 0 {
		fmt.Println("no hard bot found")
		os.Exit(1)
	}

	// Categorize games.
	var hardWins, hardLosses int
	var lostGames []int
	for i, g := range games {
		won := false
		for _, w := range g.Winners {
			if w == hardSlot {
				won = true
			}
		}
		if won {
			hardWins++
		} else {
			hardLosses++
			lostGames = append(lostGames, i)
		}
	}
	fmt.Printf("Hard bot: %d wins, %d losses (%.1f%% win rate)\n\n", hardWins, hardLosses, 100*float64(hardWins)/float64(len(games)))

	// Analyze lost games.
	analyzePassingWeaknesses(games, lostGames, hardSlot)
	analyzeMoonShotVulnerability(games, lostGames, hardSlot)
	analyzePointDumping(games, lostGames, hardSlot)
	analyzeQueenOfSpades(games, lostGames, hardSlot)
	analyzeLateGameDecisions(games, lostGames, hardSlot)
}

// hardSeat returns the seat index where the hard bot sits in a game.
func hardSeat(g sim.GameLog, hardSlot int) int {
	for seat, slot := range g.SeatToStrategy {
		if slot == hardSlot {
			return seat
		}
	}
	return -1
}

func analyzePassingWeaknesses(games []sim.GameLog, lostGames []int, hardSlot int) {
	fmt.Println("=== PASSING ANALYSIS (lost games) ===")

	// Track what the hard bot passed vs kept.
	passedHighSpades := 0
	keptHighSpades := 0
	passedQueenSpades := 0
	keptQueenSpades := 0
	passedHighHearts := 0
	totalPassRounds := 0

	for _, gi := range lostGames {
		g := games[gi]
		seat := hardSeat(g, hardSlot)
		for _, rl := range g.Rounds {
			if rl.PassDirection == "hold" {
				continue
			}
			totalPassRounds++
			passed := rl.Passes[seat]
			dealt := rl.DealtHands[seat]

			for _, cs := range passed {
				if cs == "QS" {
					passedQueenSpades++
				}
				if cs == "AS" || cs == "KS" {
					passedHighSpades++
				}
				if isHighHeart(cs) {
					passedHighHearts++
				}
			}
			for _, cs := range dealt {
				if cs == "QS" && !slices.Contains(passed, "QS") {
					keptQueenSpades++
				}
				if (cs == "AS" || cs == "KS") && !slices.Contains(passed, cs) {
					keptHighSpades++
				}
			}
		}
	}

	fmt.Printf("  Pass rounds analyzed: %d\n", totalPassRounds)
	fmt.Printf("  Passed QS: %d, Kept QS: %d\n", passedQueenSpades, keptQueenSpades)
	fmt.Printf("  Passed A/KS: %d, Kept A/KS: %d\n", passedHighSpades, keptHighSpades)
	fmt.Printf("  Passed high hearts (T+): %d\n", passedHighHearts)
	fmt.Println()
}

func analyzeMoonShotVulnerability(games []sim.GameLog, lostGames []int, hardSlot int) {
	fmt.Println("=== MOON SHOT ANALYSIS (lost games) ===")

	// Count how often the hard bot got shot by each opponent.
	shotByStrat := map[string]int{}
	hardShots := 0
	hardShotsFailed := 0 // hard bot took 20+ points but not 26

	for _, gi := range lostGames {
		g := games[gi]
		seat := hardSeat(g, hardSlot)

		for _, rl := range g.Rounds {
			// Check if anyone shot the moon.
			for s := range 4 {
				if rl.RawScores[s] == 26 {
					if s == seat {
						// Hard bot shot the moon - shouldn't happen in lost games typically
						hardShots++
					} else {
						strat := g.Strategies[g.SeatToStrategy[s]]
						shotByStrat[strat]++
					}
				}
			}
			// Check if hard bot nearly shot (20+ but not 26).
			if rl.RawScores[seat] >= 20 && rl.RawScores[seat] < 26 {
				hardShotsFailed++
			}
		}
	}

	fmt.Printf("  Hard bot got shot by opponents:\n")
	for strat, count := range shotByStrat {
		fmt.Printf("    %s: %d times\n", strat, count)
	}
	fmt.Printf("  Hard bot successful moon shots (in lost games): %d\n", hardShots)
	fmt.Printf("  Hard bot failed moon attempts (20-25 pts): %d\n", hardShotsFailed)
	fmt.Println()
}

func analyzePointDumping(games []sim.GameLog, lostGames []int, hardSlot int) {
	fmt.Println("=== POINT ACCUMULATION ANALYSIS (lost games) ===")

	// Analyze rounds where the hard bot took a lot of points.
	highPointRounds := 0   // rounds where hard bot took 10+ points
	queenSpadesCaught := 0 // rounds where hard bot ate the queen
	type bigRound struct {
		gameIdx  int
		roundIdx int
		points   int
	}
	var worstRounds []bigRound

	for _, gi := range lostGames {
		g := games[gi]
		seat := hardSeat(g, hardSlot)

		for ri, rl := range g.Rounds {
			pts := rl.RawScores[seat]
			if pts >= 10 {
				highPointRounds++
				worstRounds = append(worstRounds, bigRound{gi, ri, pts})
			}
			if pts >= 13 {
				queenSpadesCaught++
			}
		}
	}

	sort.Slice(worstRounds, func(i, j int) bool {
		return worstRounds[i].points > worstRounds[j].points
	})

	fmt.Printf("  Rounds with 10+ points: %d\n", highPointRounds)
	fmt.Printf("  Rounds catching QS (13+ pts): %d\n", queenSpadesCaught)

	// Show details of worst rounds.
	limit := min(len(worstRounds), 10)
	fmt.Printf("\n  Top %d worst rounds:\n", limit)
	for i := range limit {
		wr := worstRounds[i]
		g := games[wr.gameIdx]
		rl := g.Rounds[wr.roundIdx]
		seat := hardSeat(g, hardSlot)
		fmt.Printf("    Game %d, Round %d: %d pts (pass=%s, dealt=%s)\n",
			wr.gameIdx, wr.roundIdx, wr.points, rl.PassDirection,
			summarizeHand(rl.DealtHands[seat]))
		// Show which tricks the hard bot won with penalty cards.
		for _, tl := range rl.Tricks {
			if tl.Winner == seat && tl.Points > 0 {
				fmt.Printf("      Trick %d: won %d pts - plays: %s\n", tl.TrickNumber, tl.Points, strings.Join(tl.Plays, " "))
			}
		}
	}
	fmt.Println()
}

func analyzeQueenOfSpades(games []sim.GameLog, lostGames []int, hardSlot int) {
	fmt.Println("=== QUEEN OF SPADES ANALYSIS (lost games) ===")

	qsCaughtByHard := 0
	qsDumpedOnHard := 0 // opponent threw QS when hard was winning the trick
	totalRoundsWithQS := 0

	for _, gi := range lostGames {
		g := games[gi]
		seat := hardSeat(g, hardSlot)

		for _, rl := range g.Rounds {
			for _, tl := range rl.Tricks {
				hasQS := false
				for _, ps := range tl.Plays {
					card := strings.SplitN(ps, ":", 2)[1]
					if card == "QS" {
						hasQS = true
					}
				}
				if hasQS {
					totalRoundsWithQS++
					if tl.Winner == seat {
						qsCaughtByHard++
						// Check if QS was dumped (not led).
						leadCard := strings.SplitN(tl.Plays[0], ":", 2)[1]
						if leadCard != "QS" {
							qsDumpedOnHard++
						}
					}
				}
			}
		}
	}

	fmt.Printf("  Tricks with QS: %d\n", totalRoundsWithQS)
	fmt.Printf("  Hard bot caught QS: %d (%.1f%%)\n", qsCaughtByHard, pct(qsCaughtByHard, totalRoundsWithQS))
	fmt.Printf("  QS dumped on hard (not led): %d\n", qsDumpedOnHard)
	fmt.Println()
}

func analyzeLateGameDecisions(games []sim.GameLog, lostGames []int, hardSlot int) {
	fmt.Println("=== LATE GAME ANALYSIS (lost games) ===")

	// Look at games where hard bot was close but lost due to final rounds.
	comebackLosses := 0 // was winning at some point but lost
	for _, gi := range lostGames {
		g := games[gi]
		seat := hardSeat(g, hardSlot)

		wasLeading := false
		for _, rl := range g.Rounds {
			scores := rl.CumScores
			hardScore := scores[seat]
			minOther := 999
			for s, sc := range scores {
				if s != seat && sc < minOther {
					minOther = sc
				}
			}
			if hardScore <= minOther {
				wasLeading = true
			}
		}
		if wasLeading {
			comebackLosses++
		}
	}

	fmt.Printf("  Games where hard bot led at some point but lost: %d / %d losses\n", comebackLosses, len(lostGames))

	// Analyze margin of loss.
	var margins []int
	for _, gi := range lostGames {
		g := games[gi]
		seat := hardSeat(g, hardSlot)
		hardScore := g.FinalScores[seat]
		winnerScore := 999
		for s, sc := range g.FinalScores {
			if s != seat && sc < winnerScore {
				winnerScore = sc
			}
		}
		margins = append(margins, hardScore-winnerScore)
	}
	sort.Ints(margins)

	if len(margins) > 0 {
		sum := 0
		for _, m := range margins {
			sum += m
		}
		fmt.Printf("  Average loss margin: %.1f pts\n", float64(sum)/float64(len(margins)))
		fmt.Printf("  Median loss margin: %d pts\n", margins[len(margins)/2])
		p90 := margins[int(float64(len(margins))*0.9)]
		fmt.Printf("  90th percentile loss margin: %d pts\n", p90)
	}
	fmt.Println()
}

func isHighHeart(card string) bool {
	if len(card) != 2 || card[1] != 'H' {
		return false
	}
	return card[0] == 'T' || card[0] == 'J' || card[0] == 'Q' || card[0] == 'K' || card[0] == 'A'
}

func summarizeHand(hand []string) string {
	// Group by suit.
	suits := map[byte][]string{}
	for _, c := range hand {
		suits[c[len(c)-1]] = append(suits[c[len(c)-1]], string(c[0]))
	}
	var parts []string
	for _, s := range []byte{'C', 'D', 'S', 'H'} {
		if cards, ok := suits[s]; ok {
			parts = append(parts, fmt.Sprintf("%c:%s", s, strings.Join(cards, "")))
		}
	}
	return strings.Join(parts, " ")
}

func pct(num, denom int) float64 {
	if denom == 0 {
		return 0
	}
	return 100 * float64(num) / float64(denom)
}
