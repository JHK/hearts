package table

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"sync"

	"github.com/JHK/hearts/internal/protocol"
)

type Manager struct {
	mu     sync.RWMutex
	tables map[string]*Runtime
}

func NewManager() *Manager {
	return &Manager{tables: make(map[string]*Runtime)}
}

func (m *Manager) Get(tableID string) (*Runtime, bool) {
	id, err := normalizeTableID(tableID)
	if err != nil || id == "" {
		return nil, false
	}

	m.mu.RLock()
	table, ok := m.tables[id]
	m.mu.RUnlock()
	return table, ok
}

func (m *Manager) Create(tableID string) (*Runtime, bool, error) {
	id, err := normalizeTableID(tableID)
	if err != nil {
		return nil, false, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if id == "" {
		for attempts := 0; attempts < 3; attempts++ {
			candidate, candidateErr := randomTableID()
			if candidateErr != nil {
				return nil, false, candidateErr
			}
			if _, exists := m.tables[candidate]; !exists {
				id = candidate
				break
			}
		}

		if id == "" {
			return nil, false, fmt.Errorf("could not create unique table after 3 attempts")
		}
	}

	if existing, ok := m.tables[id]; ok {
		return existing, false, nil
	}

	created := NewRuntime(id)
	m.tables[id] = created
	return created, true, nil
}

func (m *Manager) List() []protocol.TableInfo {
	m.mu.RLock()
	tables := make([]*Runtime, 0, len(m.tables))
	for _, runtime := range m.tables {
		tables = append(tables, runtime)
	}
	m.mu.RUnlock()

	infos := make([]protocol.TableInfo, 0, len(tables))
	for _, runtime := range tables {
		infos = append(infos, runtime.Info())
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].TableID < infos[j].TableID
	})

	return infos
}

func (m *Manager) Close() {
	m.mu.Lock()
	tables := make([]*Runtime, 0, len(m.tables))
	for id, runtime := range m.tables {
		tables = append(tables, runtime)
		delete(m.tables, id)
	}
	m.mu.Unlock()

	for _, runtime := range tables {
		runtime.Close()
	}
}

func (m *Manager) CloseTable(tableID string) bool {
	id, err := normalizeTableID(tableID)
	if err != nil || id == "" {
		return false
	}

	var toClose *Runtime

	m.mu.Lock()
	if current, ok := m.tables[id]; ok {
		delete(m.tables, id)
		toClose = current
	}
	m.mu.Unlock()

	if toClose == nil {
		return false
	}

	toClose.Close()
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
