package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/JHK/hearts/internal/game/bot"
	"github.com/JHK/hearts/internal/sim"
)

func main() {
	n := flag.Int("n", 1000, "number of games to simulate (max 250000)")
	sample := flag.Int("sample", 0, "number of games to capture full logs for")
	sampleFile := flag.String("sample-file", "samples.json", "file to write sampled game logs to")
	flag.Parse()

	if *n > 250000 {
		fmt.Println("capping -n to 250000 (see strategies.md for statistical power reference)")
		*n = 250000
	}

	strategies := [4]bot.StrategyKind{
		bot.StrategyHard,
		bot.StrategyMedium,
		bot.StrategyEasy,
		bot.StrategyRandom,
	}

	s := sim.New(strategies, *n).
		WithBotOptions(bot.BotOptions{MCSamples: 3})

	var result sim.Result
	if *sample > 0 {
		var samples []sim.GameLog
		result, samples = s.RunWithSamples(*sample)

		f, err := os.Create(*sampleFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating sample file: %v\n", err)
			os.Exit(1)
		}
		enc := json.NewEncoder(f)
		enc.SetIndent("", "  ")
		if err := enc.Encode(samples); err != nil {
			fmt.Fprintf(os.Stderr, "error encoding samples: %v\n", err)
		}
		f.Close()
		fmt.Printf("Sampled %d games to %s\n", len(samples), *sampleFile)
	} else {
		result = s.Run()
	}

	fmt.Printf("Results after %d games:\n", *n)
	for i, wins := range result.Wins {
		fmt.Printf("  slot %d (%s): %d wins (%.1f%%), %d moon shots\n", i, strategies[i], wins, 100*float64(wins)/float64(*n), result.MoonShots[i])
	}
}
