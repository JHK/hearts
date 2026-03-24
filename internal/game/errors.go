package game

import "errors"

// Round state machine errors.
var (
	ErrWrongPhase           = errors.New("wrong phase")
	ErrPassAlreadySubmitted = errors.New("pass already submitted")
	ErrNotAllPassesSubmitted = errors.New("not all passes submitted")
	ErrNotYourTurn          = errors.New("not your turn")
	ErrCardNotInHand        = errors.New("card is not in hand")
)

// Card parsing errors.
var (
	ErrInvalidCardFormat = errors.New("card must be two characters like QS")
	ErrInvalidCardRank   = errors.New("invalid card rank")
	ErrInvalidCardSuit   = errors.New("invalid card suit")
)

// Play validation errors.
var (
	ErrMustLeadTwoClubs  = errors.New("first trick must be led with 2C")
	ErrHeartsNotBroken   = errors.New("hearts are not broken")
	ErrMustFollowSuit    = errors.New("must follow suit")
	ErrPenaltyCardBlocked = errors.New("penalty cards are blocked on first trick when alternatives exist")
)

// Trick scoring errors.
var (
	ErrEmptyTrick = errors.New("trick has no plays")
)
