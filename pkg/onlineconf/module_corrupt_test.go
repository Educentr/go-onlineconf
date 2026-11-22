package onlineconf

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Educentr/go-onlineconf/pkg/onlineconf_dev"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/mmap"
)

const (
	// corruptCDBContent is the content of a file that mmap opens happily but cdb rejects.
	corruptCDBContent = "this is not a cdb file at all"

	// blockTimeout is how long a read is given to return before it is considered blocked.
	// The regression these tests guard against blocks forever, so any small value works.
	blockTimeout = 2 * time.Second

	// corruptObserveWindow is how long the reads are watched after a corrupt file has been
	// written, to give the watcher goroutine time to pick the file up.
	corruptObserveWindow = 300 * time.Millisecond

	pollInterval = 5 * time.Millisecond

	cdbTestKey  = "/some/key"
	cdbOldValue = "old"
	cdbNewValue = "new"
)

// writeCorruptTreeCDB replaces <dir>/TREE.cdb with a file that is not a valid CDB the way
// onlineconf-updater would: a temporary file renamed into place, so that a running watcher
// sees a fsnotify.Create event.
func writeCorruptTreeCDB(t *testing.T, dir string) {
	t.Helper()

	target := filepath.Join(dir, DefaultModule+".cdb")
	tmp := target + ".corrupt-tmp"

	require.NoError(t, os.WriteFile(tmp, []byte(corruptCDBContent), 0o600))
	require.NoError(t, os.Rename(tmp, target))
}

// runWithTimeout runs fn in a separate goroutine and fails the test if it has not returned
// within blockTimeout. Only assert (not require) may be used inside fn: it runs outside
// the test goroutine.
func runWithTimeout(t *testing.T, what string, fn func()) {
	t.Helper()

	done := make(chan struct{})

	go func() {
		defer close(done)
		fn()
	}()

	select {
	case <-done:
	case <-time.After(blockTimeout):
		t.Fatalf("%s did not return within %s, it is blocked", what, blockTimeout)
	}
}

// A corrupt file passed to Reopen must be rejected without taking the write lock of the
// module: before the fix the error path returned while holding it, so every later read
// blocked forever on RLock.
func TestReopenCorruptCDBKeepsModuleReadable(t *testing.T) {
	m, dir := setupModule(t, map[string]interface{}{cdbTestKey: cdbOldValue})

	corruptPath := filepath.Join(dir, "corrupt.cdb")
	require.NoError(t, os.WriteFile(corruptPath, []byte(corruptCDBContent), 0o600))

	corruptMmap, err := mmap.Open(corruptPath)
	require.NoError(t, err)

	t.Cleanup(func() { corruptMmap.Close() })

	oldMmap, err := m.Reopen(corruptMmap)
	require.Error(t, err, "Reopen must reject a file that is not a valid cdb")
	require.Nil(t, oldMmap, "no mmap may be handed over for release when Reopen failed")

	runWithTimeout(t, "read after a failed Reopen", func() {
		v, getErr := m.GetString(cdbTestKey, "")
		assert.NoError(t, getErr)
		assert.Equal(t, cdbOldValue, v, "a failed Reopen must not change the values of the module")
	})

	// The module is still usable: a valid file is adopted as before.
	reopenWithData(t, m, dir, map[string]interface{}{cdbTestKey: cdbNewValue})

	v, err := m.GetString(cdbTestKey, "")
	require.NoError(t, err)
	assert.Equal(t, cdbNewValue, v)
}

// A module whose initialization failed must not stay registered: before the fix it was put
// into byName before Reopen, so the next call returned the half-initialized module with a
// nil error and reads blocked on its leaked lock.
func TestGetOrAddModuleCorruptCDBRepeatsError(t *testing.T) {
	dir := t.TempDir()
	writeCorruptTreeCDB(t, dir)

	oi := Create(WithConfigDir(dir))

	for attempt := 1; attempt <= 3; attempt++ {
		runWithTimeout(t, fmt.Sprintf("GetOrAddModule attempt %d", attempt), func() {
			m, err := oi.GetOrAddModule(DefaultModule)
			assert.Error(t, err, "every attempt on a corrupt cdb must report the error")
			assert.Nil(t, m)
		})
	}

	assert.Nil(t, oi.GetModule(DefaultModule), "a module that failed to initialize must not be registered")
	assert.Empty(t, oi.GetModuleNames())
	assert.Empty(t, oi.mmappedFiles, "the mmap of a module that failed to initialize must be released")
}

// The reproduction from the bug report: with a corrupt TREE.cdb the first read returned an
// honest error and every later read blocked forever.
func TestCorruptCDBReadsAlwaysReturnError(t *testing.T) {
	// Initialize must read the corrupt file instead of generating one from the environment.
	t.Setenv("ONLINECONFIG_FROM_ENV", "")

	dir := t.TempDir()
	writeCorruptTreeCDB(t, dir)

	ctx, err := Initialize(context.Background(), WithConfigDir(dir))
	require.NoError(t, err)

	runWithTimeout(t, "first read", func() {
		_, getErr := GetBool(ctx, "some/key", false)
		assert.Error(t, getErr)
	})

	runWithTimeout(t, "second read", func() {
		_, getErr := GetString(ctx, "other/key", "")
		assert.Error(t, getErr)
	})
}

// onlineconf-updater replacing a working module with a corrupt file must not break a
// running process: the error is logged, the module keeps serving the values of its previous
// file and the next valid file is picked up as usual.
func TestWatcherCorruptCDBKeepsPreviousValues(t *testing.T) {
	dir := t.TempDir()
	onlineconf_dev.GenerateCDB(dir, DefaultModule, map[string]interface{}{cdbTestKey: cdbOldValue})

	oi := Create(WithConfigDir(dir))
	require.NoError(t, oi.StartWatcher(context.Background()))

	watcherStopped := false

	t.Cleanup(func() {
		if !watcherStopped {
			_ = oi.StopWatcher()
		}
	})

	v, err := oi.GetString(cdbTestKey, "")
	require.NoError(t, err)
	require.Equal(t, cdbOldValue, v)

	writeCorruptTreeCDB(t, dir)

	runWithTimeout(t, "reads while the module file is corrupt", func() {
		// The watcher picks the file up asynchronously, so the reads are checked during the
		// whole window: none of them may block or see the rejected file.
		for deadline := time.Now().Add(corruptObserveWindow); time.Now().Before(deadline); {
			got, getErr := oi.GetString(cdbTestKey, "")
			if !assert.NoError(t, getErr) || !assert.Equal(t, cdbOldValue, got) {
				return
			}

			time.Sleep(pollInterval)
		}
	})

	onlineconf_dev.GenerateCDB(dir, DefaultModule, map[string]interface{}{cdbTestKey: cdbNewValue})

	runWithTimeout(t, "reads after a valid file replaced the corrupt one", func() {
		for deadline := time.Now().Add(blockTimeout / 2); time.Now().Before(deadline); {
			if got, getErr := oi.GetString(cdbTestKey, ""); getErr == nil && got == cdbNewValue {
				return
			}

			time.Sleep(pollInterval)
		}

		assert.Fail(t, "the valid file written after the corrupt one was not adopted")
	})

	require.NoError(t, oi.StopWatcher())
	watcherStopped = true

	// The rejected file was released and the adopted one replaced the original mmap.
	assert.Len(t, oi.mmappedFiles, 1, "the mmap of the rejected file must be released")
}
