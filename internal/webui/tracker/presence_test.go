package tracker

import (
	"testing"
	"testing/synctest"

	"github.com/stretchr/testify/require"
)

// synctest.Test runs the function with virtualized time so channel-based
// actor operations settle instantly via synctest.Wait(). This pattern
// replaces sleep-based polling (assertEventually) for actor tests.

func TestHumanPresence_JoinLeaveCount(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		hp := NewHumanPresence()
		t.Cleanup(hp.Shutdown)

		hp.Join("t1")
		hp.Join("t1")
		synctest.Wait()

		require.Equal(t, 2, hp.Count("t1"))
		require.Equal(t, 0, hp.Count("t2"))

		remaining := hp.Leave("t1")
		require.Equal(t, 1, remaining)

		remaining = hp.Leave("t1")
		require.Equal(t, 0, remaining)
	})
}

func TestHumanPresence_ShutdownDropsOps(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		hp := NewHumanPresence()

		hp.Join("t1")
		synctest.Wait()
		require.Equal(t, 1, hp.Count("t1"))

		hp.Shutdown()

		// After shutdown, Count returns 0 (operation dropped).
		require.Equal(t, 0, hp.Count("t1"))
		require.Equal(t, 0, hp.Leave("t1"))
	})
}

func TestPlayerPresence_MultiTab(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		pp := NewPlayerPresence()
		t.Cleanup(pp.Shutdown)

		// Simulate two browser tabs for the same player.
		pp.Join("t1", "alice")
		pp.Join("t1", "alice")
		synctest.Wait()

		// First tab closes — player still has a connection.
		remaining := pp.Leave("t1", "alice")
		require.Equal(t, 1, remaining)

		// Second tab closes — player fully disconnected.
		remaining = pp.Leave("t1", "alice")
		require.Equal(t, 0, remaining)
	})
}
