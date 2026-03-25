package protocol

type EventType string

const (
	EventPlayerJoined      EventType = "player_joined"
	EventPlayerLeft        EventType = "player_left"
	EventGameStarted       EventType = "game_started"
	EventPassSubmitted     EventType = "pass_submitted"
	EventPassReviewStarted EventType = "pass_review_started"
	EventPassReadyChanged  EventType = "pass_ready_changed"
	EventTurnChanged       EventType = "turn_changed"
	EventYourTurn          EventType = "your_turn"
	EventCardPlayed        EventType = "card_played"
	EventTrickCompleted    EventType = "trick_completed"
	EventRoundCompleted    EventType = "round_completed"
	EventHandUpdated       EventType = "hand_updated"
	EventGameOver          EventType = "game_over"
	EventGamePaused        EventType = "game_paused"
	EventGameResumed       EventType = "game_resumed"
	EventRematchVote       EventType = "rematch_vote"
	EventRematchStarting   EventType = "rematch_starting"
)
