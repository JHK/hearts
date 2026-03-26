package webui

import (
	"slices"
	"sync"
)

// lobbyHub tracks players currently browsing the lobby and broadcasts
// presence changes to all connected lobby WebSocket clients.
// It runs as a single goroutine (actor pattern); all operations are
// serialized through its ops channel.
type lobbyHub struct {
	ops      chan lobbyOp
	stop     chan struct{}
	stopped  chan struct{}
	stopOnce sync.Once
}

type lobbyOpKind int

const (
	lobbyJoinOp lobbyOpKind = iota
	lobbyLeaveOp
	lobbyUpdateNameOp
	lobbySubscribeOp
	lobbyUnsubscribeOp
	lobbySnapshotOp
	lobbyBroadcastOp
)

type lobbyOp struct {
	kind  lobbyOpKind
	token string
	name  string

	sub  chan lobbySnapshot // unsubscribe: channel to remove
	done chan struct{}      // unsubscribe: signals completion

	replyInt  chan int              // Join
	replySnap chan lobbySnapshot    // Snapshot
	replySub  chan lobbySubResult   // Subscribe
}

type lobbySubResult struct {
	ch    <-chan lobbySnapshot
	unsub func()
}

type lobbyPlayer struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type lobbySnapshot struct {
	Players []lobbyPlayer `json:"players"`
}

func newLobbyHub() *lobbyHub {
	h := &lobbyHub{
		ops:     make(chan lobbyOp),
		stop:    make(chan struct{}),
		stopped: make(chan struct{}),
	}
	go h.run()
	return h
}

func (h *lobbyHub) run() {
	defer close(h.stopped)

	players := make(map[string]string)
	refs := make(map[string]int)
	subs := make(map[chan lobbySnapshot]struct{})
	ids := make(map[string]int)
	seq := 0

	snapshot := func() lobbySnapshot {
		list := make([]lobbyPlayer, 0, len(players))
		for token, name := range players {
			list = append(list, lobbyPlayer{ID: ids[token], Name: name})
		}
		slices.SortFunc(list, func(a, b lobbyPlayer) int { return a.ID - b.ID })
		return lobbySnapshot{Players: list}
	}

	broadcast := func() {
		snap := snapshot()
		for ch := range subs {
			select {
			case ch <- snap:
			default:
			}
		}
	}

	for {
		select {
		case <-h.stop:
			for ch := range subs {
				close(ch)
			}
			return
		case op := <-h.ops:
			switch op.kind {
			case lobbyJoinOp:
				players[op.token] = op.name
				refs[op.token]++
				if _, ok := ids[op.token]; !ok {
					seq++
					ids[op.token] = seq
				}
				op.replyInt <- ids[op.token]

			case lobbyLeaveOp:
				refs[op.token]--
				if refs[op.token] > 0 {
					continue
				}
				delete(players, op.token)
				delete(refs, op.token)
				delete(ids, op.token)
				broadcast()

			case lobbyUpdateNameOp:
				if _, ok := players[op.token]; ok {
					players[op.token] = op.name
					broadcast()
				}

			case lobbySubscribeOp:
				ch := make(chan lobbySnapshot, 8)
				subs[ch] = struct{}{}
				op.replySub <- lobbySubResult{
					ch: ch,
					unsub: func() {
						done := make(chan struct{})
						if h.submit(lobbyOp{kind: lobbyUnsubscribeOp, sub: ch, done: done}) {
							<-done
						}
						for range ch {
						}
					},
				}

			case lobbyUnsubscribeOp:
				delete(subs, op.sub)
				close(op.sub)
				close(op.done)

			case lobbySnapshotOp:
				op.replySnap <- snapshot()

			case lobbyBroadcastOp:
				broadcast()
			}
		}
	}
}

func (h *lobbyHub) submit(op lobbyOp) bool {
	select {
	case <-h.stop:
		return false
	case h.ops <- op:
		return true
	}
}

// Join registers a player (or increments ref count) and returns their lobby ID.
// It does not broadcast; the caller must call Broadcast separately.
func (h *lobbyHub) Join(token, name string) int {
	reply := make(chan int, 1)
	if !h.submit(lobbyOp{kind: lobbyJoinOp, token: token, name: name, replyInt: reply}) {
		return 0
	}
	return <-reply
}

// Broadcast sends the current player list to all subscribers.
func (h *lobbyHub) Broadcast() {
	h.submit(lobbyOp{kind: lobbyBroadcastOp})
}

func (h *lobbyHub) UpdateName(token, name string) {
	h.submit(lobbyOp{kind: lobbyUpdateNameOp, token: token, name: name})
}

// Leave decrements the ref count and only removes the player when all
// connections for that token have closed (handles multi-tab).
func (h *lobbyHub) Leave(token string) {
	h.submit(lobbyOp{kind: lobbyLeaveOp, token: token})
}

// Subscribe returns a channel that receives presence snapshots and an
// unsubscribe function.
func (h *lobbyHub) Subscribe() (<-chan lobbySnapshot, func()) {
	reply := make(chan lobbySubResult, 1)
	if !h.submit(lobbyOp{kind: lobbySubscribeOp, replySub: reply}) {
		ch := make(chan lobbySnapshot)
		close(ch)
		return ch, func() {}
	}
	result := <-reply
	return result.ch, result.unsub
}

// Snapshot returns the current player list.
func (h *lobbyHub) Snapshot() lobbySnapshot {
	reply := make(chan lobbySnapshot, 1)
	if !h.submit(lobbyOp{kind: lobbySnapshotOp, replySnap: reply}) {
		return lobbySnapshot{}
	}
	return <-reply
}

// Shutdown stops the actor goroutine, closing all subscriber channels.
// It is safe to call multiple times.
func (h *lobbyHub) Shutdown() {
	h.stopOnce.Do(func() { close(h.stop) })
	<-h.stopped
}
