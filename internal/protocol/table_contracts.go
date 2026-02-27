package protocol

import "github.com/JHK/hearts/internal/game"

type TableInfo struct {
	TableID    string `json:"table_id"`
	Players    int    `json:"players"`
	MaxPlayers int    `json:"max_players"`
	Started    bool   `json:"started"`
}

type DiscoverRequest struct{}

type DiscoverResponse struct {
	Tables []TableInfo `json:"tables"`
}

type JoinRequest struct {
	Name string `json:"name"`
}

type JoinResponse struct {
	Accepted bool          `json:"accepted"`
	Reason   string        `json:"reason,omitempty"`
	TableID  string        `json:"table_id,omitempty"`
	PlayerID game.PlayerID `json:"player_id,omitempty"`
	Seat     int           `json:"seat,omitempty"`
}
