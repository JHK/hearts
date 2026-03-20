package protocol

// PlayerID is the unique identifier for a player, assigned by the table runtime
// and included in all protocol events that reference a specific player.
type PlayerID string

func (id PlayerID) String() string {
	return string(id)
}
