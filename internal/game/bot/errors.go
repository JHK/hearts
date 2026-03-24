package bot

import "errors"

var (
	ErrNoLegalPlays    = errors.New("no legal plays")
	ErrNotEnoughCards  = errors.New("not enough cards to pass")
	ErrUnknownStrategy = errors.New("unknown strategy")
)
