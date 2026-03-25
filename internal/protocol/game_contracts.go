package protocol

import "github.com/JHK/hearts/internal/game"

type CommandResponse struct {
	Accepted bool   `json:"accepted"`
	Reason   string `json:"reason,omitempty"`
}

type PlayerInfo struct {
	PlayerID PlayerID `json:"player_id"`
	Name     string   `json:"name"`
	Seat     int      `json:"seat"`
}

type PlayerJoinedData struct {
	Player PlayerInfo `json:"player"`
}

type PlayerLeftData struct {
	Player PlayerInfo `json:"player"`
}

type TurnChangedData struct {
	PlayerID    PlayerID `json:"player_id"`
	TrickNumber int      `json:"trick_number"`
}

type YourTurnData struct {
	PlayerID    PlayerID `json:"player_id"`
	TrickNumber int      `json:"trick_number"`
}

type CardPlayedData struct {
	PlayerID     PlayerID `json:"player_id"`
	Card         string   `json:"card"`
	BreaksHearts bool     `json:"breaks_hearts,omitempty"`
}

type PassStatusData struct {
	Submitted int                `json:"submitted"`
	Total     int                `json:"total"`
	Direction game.PassDirection `json:"direction,omitempty"`
}

type PassReadyData struct {
	Ready int `json:"ready"`
	Total int `json:"total"`
}

type TrickCompletedData struct {
	TrickNumber    int         `json:"trick_number"`
	WinnerPlayerID PlayerID    `json:"winner_player_id"`
	Points         game.Points `json:"points"`
}

type RoundCompletedData struct {
	RoundPoints map[PlayerID]game.Points `json:"round_points"`
	TotalPoints map[PlayerID]game.Points `json:"total_points"`
}

type HandUpdatedData struct {
	Cards []string `json:"cards"`
}

type GameOverData struct {
	FinalScores map[PlayerID]game.Points `json:"final_scores"`
	Winners     []PlayerID               `json:"winners"`
}

type GamePausedData struct {
	Player PlayerInfo `json:"player"`
}

type GameResumedData struct{}

type RematchVoteData struct {
	PlayerID PlayerID `json:"player_id"`
	Votes    int      `json:"votes"`
	Total    int      `json:"total"`
}

type RematchStartingData struct{}

type SeatClaimedData struct {
	Player  PlayerInfo `json:"player"`
	OldName string     `json:"old_name"`
}
