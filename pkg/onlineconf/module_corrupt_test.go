package onlineconf

import (
	"bytes"
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

	// truncatedCDBCut is how many bytes are cut from the tail of a valid CDB file to emulate
	// an interrupted rebuild.
	truncatedCDBCut = 10

	// garbageCDBSize is the size of the generated files that are long enough to hold an index
	// but hold no CDB.
	garbageCDBSize = 4096
)

// replaceTreeCDB replaces <dir>/TREE.cdb with the given content the way onlineconf-updater
// would: a temporary file renamed into place, so that a running watcher sees a
// fsnotify.Create event.
func replaceTreeCDB(t *testing.T, dir string, content []byte) {
	t.Helper()

	target := filepath.Join(dir, DefaultModule+".cdb")
	tmp := target + ".replace-tmp"

	require.NoError(t, os.WriteFile(tmp, content, 0o600))
	require.NoError(t, os.Rename(tmp, target))
}

// writeCorruptTreeCDB replaces <dir>/TREE.cdb with a file that is not a valid CDB at all.
func writeCorruptTreeCDB(t *testing.T, dir string) {
	t.Helper()

	replaceTreeCDB(t, dir, []byte(corruptCDBContent))
}

// validCDBContent generates a CDB file from data and returns its content.
func validCDBContent(t *testing.T, data map[string]interface{}) []byte {
	t.Helper()

	dir := t.TempDir()
	onlineconf_dev.GenerateCDB(dir, DefaultModule, data)

	content, err := os.ReadFile(filepath.Join(dir, DefaultModule+".cdb"))
	require.NoError(t, err)

	return content
}

// unusableCDBCase is a file that mmap opens happily but that cannot serve as a CDB database.
type unusableCDBCase struct {
	name    string
	content []byte
}

// unusableCDBCases returns the files a module must refuse to adopt. Only the first two are
// too short for the index and so are refused by cdb.New as well; the rest kept a full index
// and nothing but the index check stands between them and a module that answers every read
// with an error, because cdb.New reads nothing but those 2048 bytes.
func unusableCDBCases(t *testing.T) []unusableCDBCase {
	t.Helper()

	valid := validCDBContent(t, map[string]interface{}{cdbTestKey: cdbNewValue})
	require.Greater(t, len(valid), cdbIndexSize+truncatedCDBCut,
		"a valid cdb with one key must be longer than its index for the truncation cases to make sense")

	return []unusableCDBCase{
		{name: "not a cdb at all", content: []byte(corruptCDBContent)},
		{name: "shorter than the index", content: bytes.Repeat([]byte{'A'}, cdbIndexSize-1)},
		// An interrupted rebuild: the index survived, the data and the hash tables did not.
		{name: "index only", content: valid[:cdbIndexSize]},
		// The realistic truncation: only the tail of the hash tables is missing.
		{name: "truncated tail", content: valid[:len(valid)-truncatedCDBCut]},
		// A file whose index the writer never got to rewrite.
		{name: "zeroed index", content: make([]byte, garbageCDBSize)},
		{name: "garbage", content: bytes.Repeat([]byte{'A'}, garbageCDBSize)},
	}
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

// A file that is not usable as a CDB must be rejected by Reopen whatever is wrong with it.
// Only the index of a CDB file is read when it is opened, so a file that kept a plausible
// index — what an interrupted rebuild leaves behind — used to be adopted and then failed on
// every read, silently dropping the values the module had been serving.
func TestReopenRejectsUnusableCDB(t *testing.T) {
	for _, tc := range unusableCDBCases(t) {
		t.Run(tc.name, func(t *testing.T) {
			m, dir := setupModule(t, map[string]interface{}{cdbTestKey: cdbOldValue})

			rejectedPath := filepath.Join(dir, "rejected.cdb")
			require.NoError(t, os.WriteFile(rejectedPath, tc.content, 0o600))

			rejectedMmap, err := mmap.Open(rejectedPath)
			require.NoError(t, err)

			t.Cleanup(func() { rejectedMmap.Close() })

			oldMmap, err := m.Reopen(rejectedMmap)
			require.ErrorIs(t, err, ErrInvalidCDB, "Reopen must reject a file that can't serve as a cdb")
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
		})
	}
}

// The same files must keep a module out of the registry when they are found on the first
// open: a registered module backed by such a file answers every read with an error.
func TestGetOrAddModuleRejectsUnusableCDB(t *testing.T) {
	for _, tc := range unusableCDBCases(t) {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			replaceTreeCDB(t, dir, tc.content)

			oi := Create(WithConfigDir(dir))

			m, err := oi.GetOrAddModule(DefaultModule)
			require.ErrorIs(t, err, ErrInvalidCDB)
			require.Nil(t, m)

			assert.Nil(t, oi.GetModule(DefaultModule), "a module that failed to initialize must not be registered")
			assert.Empty(t, oi.GetModuleNames())
			assert.Empty(t, oi.mmappedFiles, "the mmap of a module that failed to initialize must be released")
		})
	}
}

// An interrupted rebuild of a module of a running process: the file kept its index, so it is
// opened without an error, but its hash tables are cut short. The module must keep serving
// the values of its previous file instead of answering every read with an error.
func TestWatcherTruncatedCDBKeepsPreviousValues(t *testing.T) {
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

	valid := validCDBContent(t, map[string]interface{}{cdbTestKey: cdbNewValue})
	replaceTreeCDB(t, dir, valid[:len(valid)-truncatedCDBCut])

	runWithTimeout(t, "reads while the module file is truncated", func() {
		for deadline := time.Now().Add(corruptObserveWindow); time.Now().Before(deadline); {
			got, getErr := oi.GetString(cdbTestKey, "")
			if !assert.NoError(t, getErr) || !assert.Equal(t, cdbOldValue, got) {
				return
			}

			time.Sleep(pollInterval)
		}
	})

	replaceTreeCDB(t, dir, valid)

	runWithTimeout(t, "reads after a complete file replaced the truncated one", func() {
		for deadline := time.Now().Add(blockTimeout / 2); time.Now().Before(deadline); {
			if got, getErr := oi.GetString(cdbTestKey, ""); getErr == nil && got == cdbNewValue {
				return
			}

			time.Sleep(pollInterval)
		}

		assert.Fail(t, "the complete file written after the truncated one was not adopted")
	})

	require.NoError(t, oi.StopWatcher())
	watcherStopped = true

	// The rejected file was released and the adopted one replaced the original mmap.
	assert.Len(t, oi.mmappedFiles, 1, "the mmap of the rejected file must be released")
}

// The check must not reject anything a CDB writer produces, including the degenerate case of
// a database without a single key.
func TestValidateCDBAcceptsGeneratedFiles(t *testing.T) {
	many := make(map[string]interface{}, 500)
	for i := 0; i < 500; i++ {
		many[fmt.Sprintf("/key/%d", i)] = fmt.Sprintf("value %d", i)
	}

	for name, data := range map[string]map[string]interface{}{
		"no keys":   {},
		"one key":   {cdbTestKey: cdbOldValue},
		"many keys": many,
	} {
		t.Run(name, func(t *testing.T) {
			content := validCDBContent(t, data)
			assert.NoError(t, validateCDB(bytes.NewReader(content), len(content)))
		})
	}
}
