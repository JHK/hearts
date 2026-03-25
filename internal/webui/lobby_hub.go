package webui

import (
	"slices"
	"sync"
)

// lobbyHub tracks players currently browsing the lobby and broadcasts
// presence changes to all connected lobby WebSocket clients.
type lobbyHub struct {
	mu      sync.Mutex
	players map[string]string                 // token → name
	refs    map[string]int                    // token → connection count (multi-tab)
	subs    map[chan<- lobbySnapshot]struct{} // subscriber channels
	seq     int                               // monotonic sequence for lobby IDs
	ids     map[string]int                    // token → lobby ID
}

type lobbyPlayer struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type lobbySnapshot struct {
	Players []lobbyPlayer `json:"players"`
}

func newLobbyHub() *lobbyHub {
	return &lobbyHub{
		players: make(map[string]string),
		refs:    make(map[string]int),
		subs:    make(map[chan<- lobbySnapshot]struct{}),
		ids:     make(map[string]int),
	}
}

// Join registers a player (or increments ref count) and returns their lobby ID.
func (h *lobbyHub) Join(token, name string) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.players[token] = name
	h.refs[token]++
	if _, ok := h.ids[token]; !ok {
		h.seq++
		h.ids[token] = h.seq
	}
	h.broadcastLocked()
	return h.ids[token]
}

func (h *lobbyHub) UpdateName(token, name string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.players[token]; ok {
		h.players[token] = name
		h.broadcastLocked()
	}
}

// Leave decrements the ref count and only removes the player when all
// connections for that token have closed (handles multi-tab).
func (h *lobbyHub) Leave(token string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.refs[token]--
	if h.refs[token] > 0 {
		return
	}
	delete(h.players, token)
	delete(h.refs, token)
	delete(h.ids, token)
	h.broadcastLocked()
}

// Subscribe returns a channel that receives presence snapshots and an
// unsubscribe function.
func (h *lobbyHub) Subscribe() (<-chan lobbySnapshot, func()) {
	ch := make(chan lobbySnapshot, 8)
	h.mu.Lock()
	h.subs[ch] = struct{}{}
	h.mu.Unlock()
	return ch, func() {
		h.mu.Lock()
		delete(h.subs, ch)
		h.mu.Unlock()
		// Close after removing from subs so broadcastLocked won't send
		// to a closed channel. Drain any buffered items.
		close(ch)
		for range ch {
		}
	}
}

// Snapshot returns the current player list.
func (h *lobbyHub) Snapshot() lobbySnapshot {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.snapshotLocked()
}

func (h *lobbyHub) snapshotLocked() lobbySnapshot {
	list := make([]lobbyPlayer, 0, len(h.players))
	for token, name := range h.players {
		list = append(list, lobbyPlayer{ID: h.ids[token], Name: name})
	}
	slices.SortFunc(list, func(a, b lobbyPlayer) int { return a.ID - b.ID })
	return lobbySnapshot{Players: list}
}

func (h *lobbyHub) broadcastLocked() {
	snap := h.snapshotLocked()
	for ch := range h.subs {
		select {
		case ch <- snap:
		default:
		}
	}
}
