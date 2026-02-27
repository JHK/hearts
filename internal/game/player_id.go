package game

type PlayerID string

func (id PlayerID) String() string {
	return string(id)
}
