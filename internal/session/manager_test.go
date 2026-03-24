package session

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateGeneratedTableIDUsesNamePools(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	runtime, created, err := manager.Create("")
	require.NoError(t, err, "create table")
	require.True(t, created, "expected newly created table")

	parts := strings.Split(runtime.ID(), "-")
	require.Len(t, parts, 2, "expected generated id with 2 parts, got %q", runtime.ID())
	require.True(t, contains(tableIDAdjectives, parts[0]), "expected adjective part from pool, got %q", parts[0])
	require.True(t, contains(tableIDNouns, parts[1]), "expected noun part from pool, got %q", parts[1])
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

	_, _, err := manager.Create("")
	require.NoError(t, err, "first create should succeed")

	_, _, err = manager.Create("")
	require.Error(t, err, "expected create to fail after repeated collisions")
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}

	return false
}
