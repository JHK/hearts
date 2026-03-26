// See tracker.md for package rationale.
package tracker

import (
	"context"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ConnTracker tracks active WebSocket connections so they can be interrupted
// during shutdown. http.Server.Shutdown does not close hijacked (WebSocket)
// connections, so we expire their read deadlines to unblock ReadJSON loops.
//
// All state is owned by a single goroutine (actor pattern), eliminating
// mutex-based race conditions. Track/Untrack send messages to the actor;
// Shutdown signals the actor to expire all connections. The actor keeps
// processing Untrack ops until all connections are drained, then exits.
type ConnTracker struct {
	ops      chan connOp
	stop     chan struct{}
	stopped  chan struct{}
	stopOnce sync.Once
	drained  chan struct{}
}

type connOpKind int

const (
	opTrack connOpKind = iota
	opUntrack
)

type connOp struct {
	kind connOpKind
	conn *websocket.Conn
}

func NewConnTracker() *ConnTracker {
	ct := &ConnTracker{
		ops:     make(chan connOp),
		stop:    make(chan struct{}),
		stopped: make(chan struct{}),
		drained: make(chan struct{}),
	}
	go ct.run()
	return ct
}

func (ct *ConnTracker) run() {
	defer close(ct.stopped)
	conns := make(map[*websocket.Conn]struct{})
	closing := false
	stopCh := ct.stop

	for {
		select {
		case <-stopCh:
			closing = true
			stopCh = nil // prevent busy loop on closed channel
			for conn := range conns {
				_ = conn.SetReadDeadline(time.Now())
			}
			if len(conns) == 0 {
				close(ct.drained)
				return
			}
		case op := <-ct.ops:
			switch op.kind {
			case opTrack:
				conns[op.conn] = struct{}{}
				if closing {
					_ = op.conn.SetReadDeadline(time.Now())
				}
			case opUntrack:
				delete(conns, op.conn)
				if closing && len(conns) == 0 {
					close(ct.drained)
					return
				}
			}
		}
	}
}

func (ct *ConnTracker) submit(op connOp) bool {
	select {
	case <-ct.stop:
		return false
	case ct.ops <- op:
		return true
	}
}

func (ct *ConnTracker) Track(conn *websocket.Conn) {
	ct.submit(connOp{kind: opTrack, conn: conn})
}

func (ct *ConnTracker) Untrack(conn *websocket.Conn) {
	ct.submit(connOp{kind: opUntrack, conn: conn})
}

// Shutdown signals the actor to expire all tracked connection read deadlines.
// The actor keeps processing Untrack ops until all connections are drained.
// It is safe to call multiple times.
func (ct *ConnTracker) Shutdown() {
	ct.stopOnce.Do(func() { close(ct.stop) })
}

// Wait blocks until all tracked connections have been untracked after
// Shutdown, or the context expires.
func (ct *ConnTracker) Wait(ctx context.Context) {
	select {
	case <-ct.drained:
	case <-ctx.Done():
	}
}
