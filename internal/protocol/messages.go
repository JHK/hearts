package protocol

import (
	"encoding/json"
	"time"
)

const (
	EventPlayerJoined   = "player_joined"
	EventGameStarted    = "game_started"
	EventTurnChanged    = "turn_changed"
	EventCardPlayed     = "card_played"
	EventTrickCompleted = "trick_completed"
	EventRoundCompleted = "round_completed"
	EventHandUpdated    = "hand_updated"
)

type JoinRequest struct {
	PlayerID string `json:"player_id"`
	Name     string `json:"name"`
}

type JoinResponse struct {
	Accepted bool   `json:"accepted"`
	Reason   string `json:"reason,omitempty"`
	TableID  string `json:"table_id,omitempty"`
	PlayerID string `json:"player_id,omitempty"`
	Seat     int    `json:"seat,omitempty"`
}

type StartRequest struct {
	PlayerID string `json:"player_id"`
}

type PlayCardRequest struct {
	PlayerID string `json:"player_id"`
	Card     string `json:"card"`
}

type CommandResponse struct {
	Accepted bool   `json:"accepted"`
	Reason   string `json:"reason,omitempty"`
}

type Event struct {
	Type    string          `json:"type"`
	TableID string          `json:"table_id"`
	Data    json.RawMessage `json:"data,omitempty"`
	At      time.Time       `json:"at"`
}

func EncodeEvent(tableID, eventType string, payload any) ([]byte, error) {
	encodedPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	event := Event{
		Type:    eventType,
		TableID: tableID,
		Data:    encodedPayload,
		At:      time.Now().UTC(),
	}

	return json.Marshal(event)
}

type PlayerInfo struct {
	PlayerID string `json:"player_id"`
	Name     string `json:"name"`
	Seat     int    `json:"seat"`
}

type PlayerJoinedData struct {
	Player PlayerInfo `json:"player"`
}

type GameStartedData struct {
	Players []PlayerInfo `json:"players"`
}

type TurnChangedData struct {
	PlayerID    string `json:"player_id"`
	TrickNumber int    `json:"trick_number"`
}

type CardPlayedData struct {
	PlayerID string `json:"player_id"`
	Card     string `json:"card"`
}

type TrickCompletedData struct {
	WinnerPlayerID string `json:"winner_player_id"`
	Points         int    `json:"points"`
	TrickNumber    int    `json:"trick_number"`
}

type RoundCompletedData struct {
	RoundPoints map[string]int `json:"round_points"`
	TotalPoints map[string]int `json:"total_points"`
}

type HandUpdatedData struct {
	Cards []string `json:"cards"`
}
