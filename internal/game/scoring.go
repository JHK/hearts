package game

func ApplyShootMoon(roundPoints []int) []int {
	out := make([]int, len(roundPoints))
	copy(out, roundPoints)

	shooter := -1
	for i, points := range roundPoints {
		if points == 26 {
			shooter = i
			break
		}
	}

	if shooter == -1 {
		return out
	}

	for i := range out {
		if i == shooter {
			out[i] = 0
			continue
		}
		out[i] = 26
	}

	return out
}
