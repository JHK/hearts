package protocol

import (
	"fmt"
	"strings"

	"github.com/JHK/hearts/internal/game"
)

const rootSubject = "hearts"

func DiscoverSubject() string {
	return rootSubject + ".discovery"
}

func JoinSubject(tableID string) string {
	return fmt.Sprintf("%s.table.%s.join", rootSubject, tableID)
}

func StartSubject(tableID string) string {
	return fmt.Sprintf("%s.table.%s.start", rootSubject, tableID)
}

func PlaySubject(tableID string) string {
	return fmt.Sprintf("%s.table.%s.play", rootSubject, tableID)
}

func EventsSubject(tableID string) string {
	return fmt.Sprintf("%s.table.%s.events", rootSubject, tableID)
}

func PlayerEventsSubject(tableID string, playerID game.PlayerID) string {
	return fmt.Sprintf("%s.table.%s.player.%s.events", rootSubject, tableID, playerID)
}

func PlayerEventsWildcardSubject(tableID string) string {
	return fmt.Sprintf("%s.table.%s.player.*.events", rootSubject, tableID)
}

func ParsePlayerEventsSubject(tableID, subject string) (game.PlayerID, bool) {
	prefix := fmt.Sprintf("%s.table.%s.player.", rootSubject, tableID)
	if !strings.HasPrefix(subject, prefix) {
		return game.PlayerID(""), false
	}

	trimmed := strings.TrimPrefix(subject, prefix)
	if !strings.HasSuffix(trimmed, ".events") {
		return game.PlayerID(""), false
	}

	playerID := strings.TrimSuffix(trimmed, ".events")
	if playerID == "" {
		return game.PlayerID(""), false
	}

	return game.PlayerID(playerID), true
}
