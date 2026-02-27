package protocol

type EventType string

const (
	EventPlayerJoined   EventType = "player_joined"
	EventGameStarted    EventType = "game_started"
	EventTurnChanged    EventType = "turn_changed"
	EventYourTurn       EventType = "your_turn"
	EventCardPlayed     EventType = "card_played"
	EventTrickCompleted EventType = "trick_completed"
	EventRoundCompleted EventType = "round_completed"
	EventHandUpdated    EventType = "hand_updated"
)
