package main

import (
	"flag"
	"fmt"

	"github.com/JHK/hearts/internal/game/bot"
	"github.com/JHK/hearts/internal/sim"
)

func main() {
	n := flag.Int("n", 1000, "number of games to simulate")
	flag.Parse()

	strategies := [4]bot.StrategyKind{
		bot.StrategyHard,
		bot.StrategyMedium,
		bot.StrategyEasy,
		bot.StrategyRandom,
	}

	result := sim.New(strategies, *n).Run()

	fmt.Printf("Results after %d games:\n", *n)
	for i, wins := range result.Wins {
		fmt.Printf("  slot %d (%s): %d wins (%.1f%%), %d moon shots\n", i, strategies[i], wins, 100*float64(wins)/float64(*n), result.MoonShots[i])
	}
}
