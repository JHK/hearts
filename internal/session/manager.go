package session

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"sort"
	"strings"
	"sync"

	"github.com/JHK/hearts/internal/protocol"
)

type Manager struct {
	tables map[string]*Table
	subs   map[chan<- struct{}]struct{}

	cmds      chan func()
	stop      chan struct{}
	stopped   chan struct{}
	closeOnce sync.Once
}

func NewManager() *Manager {
	m := &Manager{
		tables:  make(map[string]*Table),
		subs:    make(map[chan<- struct{}]struct{}),
		cmds:    make(chan func()),
		stop:    make(chan struct{}),
		stopped: make(chan struct{}),
	}
	go m.run()
	return m
}

func (m *Manager) run() {
	defer close(m.stopped)
	for {
		select {
		case cmd := <-m.cmds:
			cmd()
		case <-m.stop:
			return
		}
	}
}

// send executes cmd on the actor goroutine and returns true.
// Returns false if the manager is shutting down.
func (m *Manager) send(cmd func()) bool {
	select {
	case m.cmds <- cmd:
		return true
	case <-m.stop:
		return false
	}
}

// Subscribe returns a channel that receives a signal whenever the table list
// changes (table created, closed, or lobby-visible state changes) and an
// unsubscribe function.
func (m *Manager) Subscribe() (<-chan struct{}, func()) {
	ch := make(chan struct{}, 8)
	reply := make(chan struct{}, 1)
	if !m.send(func() {
		m.subs[ch] = struct{}{}
		reply <- struct{}{}
	}) {
		close(ch)
		return ch, func() {}
	}
	<-reply
	return ch, func() {
		done := make(chan struct{}, 1)
		if !m.send(func() {
			delete(m.subs, ch)
			done <- struct{}{}
		}) {
			return
		}
		<-done
		close(ch)
		for range ch {
		}
	}
}

// notifySubs sends a signal to all subscribers. Only call from the actor goroutine.
func (m *Manager) notifySubs() {
	for ch := range m.subs {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

// onTableChange is the callback passed to tables; it sends a notification
// command to the actor.
func (m *Manager) onTableChange() {
	m.send(func() {
		m.notifySubs()
	})
}

func (m *Manager) Get(tableID string) (*Table, bool) {
	id, err := normalizeTableID(tableID)
	if err != nil || id == "" {
		return nil, false
	}

	type result struct {
		table *Table
		ok    bool
	}
	reply := make(chan result, 1)
	if !m.send(func() {
		t, ok := m.tables[id]
		reply <- result{t, ok}
	}) {
		return nil, false
	}
	r := <-reply
	return r.table, r.ok
}

func (m *Manager) Create(tableID string) (*Table, bool, error) {
	id, err := normalizeTableID(tableID)
	if err != nil {
		return nil, false, err
	}

	type result struct {
		table   *Table
		created bool
		err     error
	}
	reply := make(chan result, 1)
	if !m.send(func() {
		if id == "" {
			for range 3 {
				candidate, candidateErr := randomTableID()
				if candidateErr != nil {
					reply <- result{err: candidateErr}
					return
				}
				if _, exists := m.tables[candidate]; !exists {
					id = candidate
					break
				}
			}

			if id == "" {
				reply <- result{err: fmt.Errorf("could not create unique table after 3 attempts")}
				return
			}
		}

		if existing, ok := m.tables[id]; ok {
			reply <- result{table: existing}
			return
		}

		created := NewTable(id, m.onTableChange)
		m.tables[id] = created
		slog.Info("table created", "event", "table_created", "table_id", id)
		m.notifySubs()
		reply <- result{table: created, created: true}
	}) {
		return nil, false, fmt.Errorf("manager is shutting down")
	}
	r := <-reply
	return r.table, r.created, r.err
}

func (m *Manager) List() []protocol.TableInfo {
	reply := make(chan []*Table, 1)
	if !m.send(func() {
		tables := make([]*Table, 0, len(m.tables))
		for _, runtime := range m.tables {
			tables = append(tables, runtime)
		}
		reply <- tables
	}) {
		return nil
	}
	tables := <-reply

	infos := make([]protocol.TableInfo, 0, len(tables))
	for _, runtime := range tables {
		info := runtime.Info()
		if info.GameOver {
			continue
		}
		infos = append(infos, info)
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].TableID < infos[j].TableID
	})

	return infos
}

func (m *Manager) Close() {
	reply := make(chan []*Table, 1)
	sent := m.send(func() {
		tables := make([]*Table, 0, len(m.tables))
		for id, runtime := range m.tables {
			tables = append(tables, runtime)
			delete(m.tables, id)
		}
		reply <- tables
	})

	var tables []*Table
	if sent {
		tables = <-reply
	}

	m.closeOnce.Do(func() { close(m.stop) })
	<-m.stopped

	for _, runtime := range tables {
		slog.Info("table destroyed", "event", "table_destroyed", "table_id", runtime.ID())
		runtime.Close()
	}
}

func (m *Manager) CloseTable(tableID string) bool {
	id, err := normalizeTableID(tableID)
	if err != nil || id == "" {
		return false
	}

	type result struct {
		table *Table
		found bool
	}
	reply := make(chan result, 1)
	if !m.send(func() {
		t, ok := m.tables[id]
		if ok {
			delete(m.tables, id)
			m.notifySubs()
		}
		reply <- result{t, ok}
	}) {
		return false
	}
	r := <-reply
	if !r.found {
		return false
	}

	slog.Info("table destroyed", "event", "table_destroyed", "table_id", id)
	r.table.Close()
	return true
}

func normalizeTableID(raw string) (string, error) {
	id := strings.ToLower(strings.TrimSpace(raw))
	if id == "" {
		return "", nil
	}

	for _, ch := range id {
		isLetter := ch >= 'a' && ch <= 'z'
		isDigit := ch >= '0' && ch <= '9'
		if isLetter || isDigit || ch == '-' {
			continue
		}
		return "", fmt.Errorf("invalid table id %q (allowed: a-z, 0-9, -)", raw)
	}

	return id, nil
}

func randomTableID() (string, error) {
	first, err := randomItem(tableIDAdjectives)
	if err != nil {
		return "", err
	}

	second, err := randomItem(tableIDNouns)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s-%s", first, second), nil
}

var tableIDAdjectives = []string{
	"amber",
	"brisk",
	"cedar",
	"clear",
	"coral",
	"dawn",
	"ember",
	"gold",
	"honey",
	"ivory",
	"jade",
	"lucky",
	"maple",
	"merry",
	"misty",
	"opal",
	"pearl",
	"river",
	"silver",
	"sunny",
}

var tableIDNouns = []string{
	"acorn",
	"anchor",
	"brook",
	"cabin",
	"canyon",
	"comet",
	"field",
	"forest",
	"harbor",
	"island",
	"lantern",
	"meadow",
	"orchard",
	"pavilion",
	"summit",
	"tavern",
	"thicket",
	"valley",
	"willow",
	"windmill",
}

func randomItem(items []string) (string, error) {
	if len(items) == 0 {
		return "", fmt.Errorf("random source has no items")
	}

	limit := big.NewInt(int64(len(items)))
	index, err := rand.Int(rand.Reader, limit)
	if err != nil {
		return "", err
	}

	return items[index.Int64()], nil
}
