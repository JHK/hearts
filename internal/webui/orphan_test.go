package webui

import (
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/JHK/hearts/internal/webui/tracker"
	"github.com/stretchr/testify/require"
)

func TestOrphanCleanup_ClosesAfterGracePeriod(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		presence := tracker.NewHumanPresence()
		t.Cleanup(presence.Shutdown)

		var closed atomic.Bool
		closeTable := func(string) bool { closed.Store(true); return true }

		scheduleOrphanCleanup("table-1", 60*time.Second, presence, closeTable)

		// Timer is running but hasn't fired yet.
		synctest.Wait()
		require.False(t, closed.Load(), "should not close before grace period")

		// Advance past the grace period.
		time.Sleep(60 * time.Second)
		synctest.Wait()
		require.True(t, closed.Load(), "should close after grace period")
	})
}

func TestOrphanCleanup_HumanReconnects(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		presence := tracker.NewHumanPresence()
		t.Cleanup(presence.Shutdown)

		var closed atomic.Bool
		closeTable := func(string) bool { closed.Store(true); return true }

		scheduleOrphanCleanup("table-1", 60*time.Second, presence, closeTable)

		// A human reconnects during the grace period.
		time.Sleep(30 * time.Second)
		presence.Join("table-1")

		// Grace period expires, but a human is present.
		time.Sleep(30 * time.Second)
		synctest.Wait()
		require.False(t, closed.Load(), "should not close when humans are present")
	})
}

func TestOrphanCleanup_ShortGracePeriod(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		presence := tracker.NewHumanPresence()
		t.Cleanup(presence.Shutdown)

		var closed atomic.Bool
		closeTable := func(string) bool { closed.Store(true); return true }

		// Pre-game tables use the short grace period.
		scheduleOrphanCleanup("table-1", 500*time.Millisecond, presence, closeTable)

		synctest.Wait()
		require.False(t, closed.Load(), "should not close before grace period")

		time.Sleep(500 * time.Millisecond)
		synctest.Wait()
		require.True(t, closed.Load(), "should close after short grace period")
	})
}
