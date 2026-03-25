package protocol

type TableInfo struct {
	TableID    string `json:"table_id"`
	Players    int    `json:"players"`
	MaxPlayers int    `json:"max_players"`
	Started    bool   `json:"started"`
	GameOver   bool   `json:"game_over,omitempty"`
	Paused     bool   `json:"paused,omitempty"`
}

type JoinResponse struct {
	Accepted bool     `json:"accepted"`
	Reason   string   `json:"reason,omitempty"`
	TableID  string   `json:"table_id,omitempty"`
	PlayerID PlayerID `json:"player_id,omitempty"`
	Seat     int      `json:"seat,omitempty"`
}
