package game

func TrickWinnerIndex(lead Suit, trick []Card) int {
	winnerIdx := -1
	winnerRank := -1

	for i, card := range trick {
		if card.Suit != lead {
			continue
		}
		if card.Rank > winnerRank {
			winnerRank = card.Rank
			winnerIdx = i
		}
	}

	return winnerIdx
}

func TrickPoints(trick []Card) int {
	points := 0
	for _, card := range trick {
		if card.Suit == Hearts {
			points++
		}
		if card.Suit == Spades && card.Rank == 12 {
			points += 13
		}
	}
	return points
}
