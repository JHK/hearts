package table

import (
	"strings"
	"testing"
)

func TestCreateGeneratedTableIDUsesNamePools(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	runtime, created, err := manager.Create("")
	if err != nil {
		t.Fatalf("create table: %v", err)
	}
	if !created {
		t.Fatalf("expected newly created table")
	}

	parts := strings.Split(runtime.ID(), "-")
	if len(parts) != 2 {
		t.Fatalf("expected generated id with 2 parts, got %q", runtime.ID())
	}
	if !contains(tableIDAdjectives, parts[0]) {
		t.Fatalf("expected adjective part from pool, got %q", parts[0])
	}
	if !contains(tableIDNouns, parts[1]) {
		t.Fatalf("expected noun part from pool, got %q", parts[1])
	}
}

func TestCreateFailsAfterThreeNameCollisions(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	originalAdjectives := tableIDAdjectives
	originalNouns := tableIDNouns
	tableIDAdjectives = []string{"only"}
	tableIDNouns = []string{"name"}
	defer func() {
		tableIDAdjectives = originalAdjectives
		tableIDNouns = originalNouns
	}()

	if _, _, err := manager.Create(""); err != nil {
		t.Fatalf("first create should succeed: %v", err)
	}

	if _, _, err := manager.Create(""); err == nil {
		t.Fatalf("expected create to fail after repeated collisions")
	}
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}

	return false
}
