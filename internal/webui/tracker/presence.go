// See tracker.md for package rationale.
package tracker

import "sync"

// HumanPresence counts active human WebSocket connections per table.
//
// All state is owned by a single goroutine (actor pattern). Join/Leave/Count
// send messages to the actor. After Shutdown, new operations are dropped
// (Leave/Count return 0).
type HumanPresence struct {
	ops      chan humanOp
	stop     chan struct{}
	stopped  chan struct{}
	stopOnce sync.Once
}

type humanOpKind int

const (
	humanJoin humanOpKind = iota
	humanLeave
	humanCount
)

type humanOp struct {
	kind    humanOpKind
	tableID string
	reply   chan int // used by leave and count
}

func NewHumanPresence() *HumanPresence {
	hp := &HumanPresence{
		ops:     make(chan humanOp),
		stop:    make(chan struct{}),
		stopped: make(chan struct{}),
	}
	go hp.run()
	return hp
}

func (hp *HumanPresence) run() {
	defer close(hp.stopped)
	counts := make(map[string]int)
	for {
		select {
		case <-hp.stop:
			return
		case op := <-hp.ops:
			switch op.kind {
			case humanJoin:
				counts[op.tableID]++
			case humanLeave:
				remaining := counts[op.tableID] - 1
				if remaining <= 0 {
					delete(counts, op.tableID)
					remaining = 0
				} else {
					counts[op.tableID] = remaining
				}
				op.reply <- remaining
			case humanCount:
				op.reply <- counts[op.tableID]
			}
		}
	}
}

func (hp *HumanPresence) submit(op humanOp) bool {
	select {
	case <-hp.stop:
		return false
	case hp.ops <- op:
		return true
	}
}

func (hp *HumanPresence) Join(tableID string) {
	hp.submit(humanOp{kind: humanJoin, tableID: tableID})
}

func (hp *HumanPresence) Leave(tableID string) int {
	reply := make(chan int, 1)
	if !hp.submit(humanOp{kind: humanLeave, tableID: tableID, reply: reply}) {
		return 0
	}
	return <-reply
}

func (hp *HumanPresence) Count(tableID string) int {
	reply := make(chan int, 1)
	if !hp.submit(humanOp{kind: humanCount, tableID: tableID, reply: reply}) {
		return 0
	}
	return <-reply
}

// Shutdown stops the actor goroutine. It is safe to call multiple times.
func (hp *HumanPresence) Shutdown() {
	hp.stopOnce.Do(func() { close(hp.stop) })
	<-hp.stopped
}

// PlayerPresence counts active WebSocket connections per player per table.
// This prevents spurious Leave calls when a player has multiple tabs open.
//
// All state is owned by a single goroutine (actor pattern). Join/Leave send
// messages to the actor. After Shutdown, new operations are dropped
// (Leave returns 0).
type PlayerPresence struct {
	ops      chan playerOp
	stop     chan struct{}
	stopped  chan struct{}
	stopOnce sync.Once
}

type playerOpKind int

const (
	playerJoin playerOpKind = iota
	playerLeave
)

type playerOp struct {
	kind     playerOpKind
	tableID  string
	playerID string
	reply    chan int // used by leave
}

func NewPlayerPresence() *PlayerPresence {
	pp := &PlayerPresence{
		ops:     make(chan playerOp),
		stop:    make(chan struct{}),
		stopped: make(chan struct{}),
	}
	go pp.run()
	return pp
}

func (pp *PlayerPresence) run() {
	defer close(pp.stopped)
	counts := make(map[string]int) // key: "tableID\x00playerID"
	for {
		select {
		case <-pp.stop:
			return
		case op := <-pp.ops:
			switch op.kind {
			case playerJoin:
				counts[op.tableID+"\x00"+op.playerID]++
			case playerLeave:
				key := op.tableID + "\x00" + op.playerID
				remaining := counts[key] - 1
				if remaining <= 0 {
					delete(counts, key)
					remaining = 0
				} else {
					counts[key] = remaining
				}
				op.reply <- remaining
			}
		}
	}
}

func (pp *PlayerPresence) submit(op playerOp) bool {
	select {
	case <-pp.stop:
		return false
	case pp.ops <- op:
		return true
	}
}

func (pp *PlayerPresence) Join(tableID string, playerID string) {
	pp.submit(playerOp{kind: playerJoin, tableID: tableID, playerID: playerID})
}

// Leave decrements the count and returns the remaining connections.
func (pp *PlayerPresence) Leave(tableID string, playerID string) int {
	reply := make(chan int, 1)
	if !pp.submit(playerOp{kind: playerLeave, tableID: tableID, playerID: playerID, reply: reply}) {
		return 0
	}
	return <-reply
}

// Shutdown stops the actor goroutine. It is safe to call multiple times.
func (pp *PlayerPresence) Shutdown() {
	pp.stopOnce.Do(func() { close(pp.stop) })
	<-pp.stopped
}
