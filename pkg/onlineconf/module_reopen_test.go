package onlineconf

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/Educentr/go-onlineconf/pkg/onlineconf_dev"
	"github.com/colinmarc/cdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/mmap"
)

// setupModule creates a temp dir, generates a CDB file from data,
// and returns a Module ready for testing Reopen.
func setupModule(t *testing.T, data map[string]interface{}) (m *Module, dir string) {
	t.Helper()

	dir, err := os.MkdirTemp("", "onlineconf_reopen_test*")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(dir) })

	moduleName := "test"
	onlineconf_dev.GenerateCDB(dir, moduleName, data)

	mmapFile, err := mmap.Open(filepath.Join(dir, moduleName+".cdb"))
	require.NoError(t, err)

	cdbReader, err := cdb.New(mmapFile, nil)
	require.NoError(t, err)

	m = &Module{
		name:        moduleName,
		filename:    filepath.Join(dir, moduleName+".cdb"),
		cdb:         cdbReader,
		mmappedFile: mmapFile,
		cache:       make(map[string][]interface{}, startCacheSize),
	}

	return m, dir
}

// reopenWithData generates a new CDB and calls Reopen on the module.
func reopenWithData(t *testing.T, m *Module, dir string, newData map[string]interface{}) {
	t.Helper()

	moduleName := "test"
	onlineconf_dev.GenerateCDB(dir, moduleName, newData)

	newMmap, err := mmap.Open(filepath.Join(dir, moduleName+".cdb"))
	require.NoError(t, err)

	oldMmap, err := m.Reopen(newMmap)
	require.NoError(t, err)

	if oldMmap != nil {
		oldMmap.Close()
	}
}

type atomicCounter struct {
	val atomic.Int32
}

func (c *atomicCounter) callback() error {
	c.val.Add(1)
	return nil
}

func (c *atomicCounter) count() int32 {
	return c.val.Load()
}

// BUG: callback is fired even when the subscribed key's value hasn't changed.
// errors.Is(nil, nil) returns true, so the first condition in Reopen always triggers.
func TestReopen_CallbackNotCalledWhenValueUnchanged(t *testing.T) {
	data := map[string]interface{}{"key1": "value1", "key2": "value2"}
	m, dir := setupModule(t, data)

	counter := &atomicCounter{}
	m.RegisterSubscription(NewSubscription([]string{"key1"}, counter.callback))

	// Reopen with identical data — callback should NOT fire
	reopenWithData(t, m, dir, data)

	assert.Equal(t, int32(0), counter.count(),
		"callback should NOT be called when subscribed key value is unchanged")
}

// Sanity check: callback SHOULD fire when the subscribed key's value changes.
func TestReopen_CallbackCalledWhenValueChanged(t *testing.T) {
	data := map[string]interface{}{"key1": "value1", "key2": "value2"}
	m, dir := setupModule(t, data)

	counter := &atomicCounter{}
	m.RegisterSubscription(NewSubscription([]string{"key1"}, counter.callback))

	newData := map[string]interface{}{"key1": "new_value1", "key2": "value2"}
	reopenWithData(t, m, dir, newData)

	assert.Equal(t, int32(1), counter.count(),
		"callback SHOULD be called when subscribed key value changed")
}

// BUG: callback fires when only unrelated keys change.
func TestReopen_CallbackNotCalledWhenUnrelatedKeyChanges(t *testing.T) {
	data := map[string]interface{}{"key1": "value1", "key2": "value2"}
	m, dir := setupModule(t, data)

	counter := &atomicCounter{}
	m.RegisterSubscription(NewSubscription([]string{"key1"}, counter.callback))

	// Only key2 changes
	newData := map[string]interface{}{"key1": "value1", "key2": "new_value2"}
	reopenWithData(t, m, dir, newData)

	assert.Equal(t, int32(0), counter.count(),
		"callback should NOT be called when only an unrelated key changed")
}

// Callback should fire when a subscribed key is added (didn't exist before).
func TestReopen_CallbackCalledWhenKeyAdded(t *testing.T) {
	data := map[string]interface{}{"key1": "value1"}
	m, dir := setupModule(t, data)

	counter := &atomicCounter{}
	m.RegisterSubscription(NewSubscription([]string{"key2"}, counter.callback))

	// key2 now appears
	newData := map[string]interface{}{"key1": "value1", "key2": "value2"}
	reopenWithData(t, m, dir, newData)

	assert.Equal(t, int32(1), counter.count(),
		"callback SHOULD be called when subscribed key is added")
}

// Callback should fire when a subscribed key is removed.
func TestReopen_CallbackCalledWhenKeyRemoved(t *testing.T) {
	data := map[string]interface{}{"key1": "value1", "key2": "value2"}
	m, dir := setupModule(t, data)

	counter := &atomicCounter{}
	m.RegisterSubscription(NewSubscription([]string{"key2"}, counter.callback))

	// key2 removed
	newData := map[string]interface{}{"key1": "value1"}
	reopenWithData(t, m, dir, newData)

	assert.Equal(t, int32(1), counter.count(),
		"callback SHOULD be called when subscribed key is removed")
}

// BUG: when a subscription has multiple paths and several change,
// the callback is added to the list multiple times (no break).
func TestReopen_CallbackCalledOnceForMultiPathSubscription(t *testing.T) {
	data := map[string]interface{}{"key1": "value1", "key2": "value2", "key3": "value3"}
	m, dir := setupModule(t, data)

	counter := &atomicCounter{}
	m.RegisterSubscription(NewSubscription([]string{"key1", "key2"}, counter.callback))

	// Both key1 and key2 change
	newData := map[string]interface{}{"key1": "new1", "key2": "new2", "key3": "value3"}
	reopenWithData(t, m, dir, newData)

	assert.Equal(t, int32(1), counter.count(),
		"callback should be called exactly once even when multiple subscribed paths change")
}

// nil paths = subscribe to any change in the module.
func TestReopen_NilPathsAlwaysTrigger(t *testing.T) {
	data := map[string]interface{}{"key1": "value1"}
	m, dir := setupModule(t, data)

	counter := &atomicCounter{}
	m.RegisterSubscription(NewSubscription(nil, counter.callback))

	// Reopen with same data — callback should still fire for nil paths
	reopenWithData(t, m, dir, data)

	assert.Equal(t, int32(1), counter.count(),
		"nil paths subscription should always trigger callback on reopen")
}

// Empty string path = subscribe to any change (same as nil).
func TestReopen_EmptyPathAlwaysTriggers(t *testing.T) {
	data := map[string]interface{}{"key1": "value1"}
	m, dir := setupModule(t, data)

	counter := &atomicCounter{}
	m.RegisterSubscription(NewSubscription([]string{""}, counter.callback))

	// Reopen with same data — callback should still fire for empty path
	reopenWithData(t, m, dir, data)

	assert.Equal(t, int32(1), counter.count(),
		"empty string path subscription should always trigger callback on reopen")
}

// Multiple independent subscriptions: only the ones whose keys changed should fire.
func TestReopen_IndependentSubscriptionsFireSelectively(t *testing.T) {
	data := map[string]interface{}{"key1": "value1", "key2": "value2"}
	m, dir := setupModule(t, data)

	counter1 := &atomicCounter{}
	counter2 := &atomicCounter{}
	m.RegisterSubscription(NewSubscription([]string{"key1"}, counter1.callback))
	m.RegisterSubscription(NewSubscription([]string{"key2"}, counter2.callback))

	// Only key2 changes
	newData := map[string]interface{}{"key1": "value1", "key2": "new_value2"}
	reopenWithData(t, m, dir, newData)

	assert.Equal(t, int32(0), counter1.count(),
		"subscription on key1 should NOT fire when key1 is unchanged")
	assert.Equal(t, int32(1), counter2.count(),
		"subscription on key2 SHOULD fire when key2 changed")
}
