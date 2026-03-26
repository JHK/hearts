// See tracker.md for package rationale.
package tracker

import (
	"context"
	"time"

	"github.com/gorilla/websocket"
)

// ConnTracker tracks active WebSocket connections so they can be interrupted
// during shutdown. http.Server.Shutdown does not close hijacked (WebSocket)
// connections, so we expire their read deadlines to unblock ReadJSON loops.
//
// All state is owned by a single goroutine (actor pattern), eliminating
// mutex-based race conditions. Track/Untrack/Shutdown send messages to the
// actor; the drained channel is closed once shutdown has been requested and
// all tracked connections have been untracked.
type ConnTracker struct {
	ops     chan connOp
	drained chan struct{}
}

type connOpKind int

const (
	opTrack connOpKind = iota
	opUntrack
	opShutdown
)

type connOp struct {
	kind connOpKind
	conn *websocket.Conn
}

func NewConnTracker() *ConnTracker {
	ct := &ConnTracker{
		ops:     make(chan connOp),
		drained: make(chan struct{}),
	}
	go ct.run()
	return ct
}

func (ct *ConnTracker) run() {
	conns := make(map[*websocket.Conn]struct{})
	closing := false

	for op := range ct.ops {
		switch op.kind {
		case opTrack:
			conns[op.conn] = struct{}{}
			if closing {
				_ = op.conn.SetReadDeadline(time.Now())
			}
		case opUntrack:
			delete(conns, op.conn)
			if closing && len(conns) == 0 && ct.drained != nil {
				close(ct.drained)
				ct.drained = nil
			}
		case opShutdown:
			if closing {
				continue
			}
			closing = true
			for conn := range conns {
				_ = conn.SetReadDeadline(time.Now())
			}
			if len(conns) == 0 {
				close(ct.drained)
				ct.drained = nil
			}
		}
	}
}

func (ct *ConnTracker) Track(conn *websocket.Conn) {
	ct.ops <- connOp{kind: opTrack, conn: conn}
}

func (ct *ConnTracker) Untrack(conn *websocket.Conn) {
	ct.ops <- connOp{kind: opUntrack, conn: conn}
}

func (ct *ConnTracker) Shutdown() {
	ct.ops <- connOp{kind: opShutdown}
}

// Wait blocks until all tracked connections have been untracked after
// Shutdown, or the context expires.
func (ct *ConnTracker) Wait(ctx context.Context) {
	select {
	case <-ct.drained:
	case <-ctx.Done():
	}
}
