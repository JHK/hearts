// See tracker.md for package rationale.
package tracker

import "sync"

// HumanPresence counts active human WebSocket connections per table.
type HumanPresence struct {
	mu     sync.Mutex
	counts map[string]int
}

func NewHumanPresence() *HumanPresence {
	return &HumanPresence{counts: make(map[string]int)}
}

func (t *HumanPresence) Join(tableID string) {
	t.mu.Lock()
	t.counts[tableID]++
	t.mu.Unlock()
}

func (t *HumanPresence) Leave(tableID string) int {
	t.mu.Lock()
	defer t.mu.Unlock()

	remaining := t.counts[tableID] - 1
	if remaining <= 0 {
		delete(t.counts, tableID)
		return 0
	}

	t.counts[tableID] = remaining
	return remaining
}

func (t *HumanPresence) Count(tableID string) int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.counts[tableID]
}

// PlayerPresence counts active WebSocket connections per player per table.
// This prevents spurious Leave calls when a player has multiple tabs open.
type PlayerPresence struct {
	mu     sync.Mutex
	counts map[string]int // key: "tableID\x00playerID"
}

func NewPlayerPresence() *PlayerPresence {
	return &PlayerPresence{counts: make(map[string]int)}
}

func (t *PlayerPresence) Join(tableID string, playerID string) {
	t.mu.Lock()
	t.counts[tableID+"\x00"+playerID]++
	t.mu.Unlock()
}

// Leave decrements the count and returns the remaining connections.
func (t *PlayerPresence) Leave(tableID string, playerID string) int {
	t.mu.Lock()
	defer t.mu.Unlock()

	key := tableID + "\x00" + playerID
	remaining := t.counts[key] - 1
	if remaining <= 0 {
		delete(t.counts, key)
		return 0
	}
	t.counts[key] = remaining
	return remaining
}
