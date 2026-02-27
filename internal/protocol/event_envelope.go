package protocol

import "encoding/json"

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

type Event struct {
	Type EventType         `json:"type"`
	Data json.RawMessage   `json:"data,omitempty"`
	Meta map[string]string `json:"meta,omitempty"`
}

func EncodeEvent[T any](eventType EventType, data T) ([]byte, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	event := Event{Type: eventType, Data: payload}
	return json.Marshal(event)
}
