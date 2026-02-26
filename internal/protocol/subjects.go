package protocol

import "fmt"

func JoinSubject(tableID string) string {
	return fmt.Sprintf("hearts.table.%s.join", tableID)
}

func StartSubject(tableID string) string {
	return fmt.Sprintf("hearts.table.%s.start", tableID)
}

func PlaySubject(tableID string) string {
	return fmt.Sprintf("hearts.table.%s.play", tableID)
}

func EventsSubject(tableID string) string {
	return fmt.Sprintf("hearts.table.%s.events", tableID)
}

func PlayerEventsSubject(tableID, playerID string) string {
	return fmt.Sprintf("hearts.table.%s.player.%s.events", tableID, playerID)
}
