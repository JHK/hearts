package session

import "errors"

var (
	ErrGameOver       = errors.New("game is over")
	ErrRoundInProgress = errors.New("round already in progress")
	ErrTableFull      = errors.New("table is full")
	ErrTableStopping  = errors.New("table is stopping")
)
