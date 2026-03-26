// See tracker.md for package rationale.
package tracker

// HumanPresence counts active human WebSocket connections per table.
//
// All state is owned by a single goroutine (actor pattern). Join/Leave/Count
// send messages to the actor. After Shutdown, new Joins are ignored but
// Leave/Count keep working so in-flight teardown completes safely.
type HumanPresence struct {
	ops chan humanOp
}

type humanOpKind int

const (
	humanJoin humanOpKind = iota
	humanLeave
	humanCount
	humanShutdown
)

type humanOp struct {
	kind    humanOpKind
	tableID string
	reply   chan int // used by leave and count
}

func NewHumanPresence() *HumanPresence {
	hp := &HumanPresence{ops: make(chan humanOp)}
	go hp.run()
	return hp
}

func (hp *HumanPresence) run() {
	counts := make(map[string]int)
	closing := false
	for op := range hp.ops {
		switch op.kind {
		case humanJoin:
			if !closing {
				counts[op.tableID]++
			}
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
		case humanShutdown:
			closing = true
		}
	}
}

func (hp *HumanPresence) Join(tableID string) {
	hp.ops <- humanOp{kind: humanJoin, tableID: tableID}
}

func (hp *HumanPresence) Leave(tableID string) int {
	reply := make(chan int, 1)
	hp.ops <- humanOp{kind: humanLeave, tableID: tableID, reply: reply}
	return <-reply
}

func (hp *HumanPresence) Count(tableID string) int {
	reply := make(chan int, 1)
	hp.ops <- humanOp{kind: humanCount, tableID: tableID, reply: reply}
	return <-reply
}

func (hp *HumanPresence) Shutdown() {
	hp.ops <- humanOp{kind: humanShutdown}
}

// PlayerPresence counts active WebSocket connections per player per table.
// This prevents spurious Leave calls when a player has multiple tabs open.
//
// All state is owned by a single goroutine (actor pattern). Join/Leave send
// messages to the actor. After Shutdown, new Joins are ignored but Leave
// keeps working so in-flight teardown completes safely.
type PlayerPresence struct {
	ops chan playerOp
}

type playerOpKind int

const (
	playerJoin playerOpKind = iota
	playerLeave
	playerShutdown
)

type playerOp struct {
	kind     playerOpKind
	tableID  string
	playerID string
	reply    chan int // used by leave
}

func NewPlayerPresence() *PlayerPresence {
	pp := &PlayerPresence{ops: make(chan playerOp)}
	go pp.run()
	return pp
}

func (pp *PlayerPresence) run() {
	counts := make(map[string]int) // key: "tableID\x00playerID"
	closing := false
	for op := range pp.ops {
		switch op.kind {
		case playerJoin:
			if !closing {
				counts[op.tableID+"\x00"+op.playerID]++
			}
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
		case playerShutdown:
			closing = true
		}
	}
}

func (pp *PlayerPresence) Join(tableID string, playerID string) {
	pp.ops <- playerOp{kind: playerJoin, tableID: tableID, playerID: playerID}
}

// Leave decrements the count and returns the remaining connections.
func (pp *PlayerPresence) Leave(tableID string, playerID string) int {
	reply := make(chan int, 1)
	pp.ops <- playerOp{kind: playerLeave, tableID: tableID, playerID: playerID, reply: reply}
	return <-reply
}

func (pp *PlayerPresence) Shutdown() {
	pp.ops <- playerOp{kind: playerShutdown}
}
