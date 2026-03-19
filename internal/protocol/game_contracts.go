package protocol

import "github.com/JHK/hearts/internal/game"

type CommandResponse struct {
	Accepted bool   `json:"accepted"`
	Reason   string `json:"reason,omitempty"`
}

type PlayerInfo struct {
	PlayerID game.PlayerID `json:"player_id"`
	Name     string        `json:"name"`
	Seat     int           `json:"seat"`
}

type PlayerJoinedData struct {
	Player PlayerInfo `json:"player"`
}

type TurnChangedData struct {
	PlayerID    game.PlayerID `json:"player_id"`
	TrickNumber int           `json:"trick_number"`
}

type YourTurnData struct {
	PlayerID    game.PlayerID `json:"player_id"`
	TrickNumber int           `json:"trick_number"`
}

type CardPlayedData struct {
	PlayerID     game.PlayerID `json:"player_id"`
	Card         string        `json:"card"`
	BreaksHearts bool          `json:"breaks_hearts,omitempty"`
}

type PassStatusData struct {
	Submitted int    `json:"submitted"`
	Total     int    `json:"total"`
	Direction game.PassDirection `json:"direction,omitempty"`
}

type PassReadyData struct {
	Ready int `json:"ready"`
	Total int `json:"total"`
}

type TrickCompletedData struct {
	TrickNumber    int           `json:"trick_number"`
	WinnerPlayerID game.PlayerID `json:"winner_player_id"`
	Points         game.Points   `json:"points"`
}

type RoundCompletedData struct {
	RoundPoints map[game.PlayerID]game.Points `json:"round_points"`
	TotalPoints map[game.PlayerID]game.Points `json:"total_points"`
}

type HandUpdatedData struct {
	Cards []string `json:"cards"`
}

type GameOverData struct {
	FinalScores map[game.PlayerID]game.Points `json:"final_scores"`
	Winners     []game.PlayerID               `json:"winners"`
}
