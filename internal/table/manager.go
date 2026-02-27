package table

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
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
		for {
			candidate, candidateErr := randomTableID()
			if candidateErr != nil {
				return nil, false, candidateErr
			}
			if _, exists := m.tables[candidate]; !exists {
				id = candidate
				break
			}
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
	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "table-" + hex.EncodeToString(buf), nil
}
