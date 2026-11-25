package onlineconf_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Educentr/go-onlineconf/pkg/onlineconf"
	"github.com/Educentr/go-onlineconf/pkg/onlineconf_dev"
	"github.com/stretchr/testify/require"
)

func TestReopenWaiter(t *testing.T) {
	tmpLogFile := "/tmp/onlineconfLogger.log"
	tmpDir, _ := os.MkdirTemp("", "onlineconf*")

	os.Remove(tmpLogFile)

	testConfig := map[string]any{"/some/log/level": "debug", "/some/log/output/err": tmpLogFile, "/some/log/output/out": tmpLogFile}

	onlineconf_dev.GenerateCDB(tmpDir, "TREE", testConfig)

	inst := onlineconf.Create(onlineconf.WithConfigDir(tmpDir))
	err := inst.StartWatcher(context.Background())
	require.NoError(t, err, "can't start watcher")

	val, err := inst.GetString("/some/log/level")
	require.NoError(t, err, "can't get string")
	require.Equal(t, "debug", val)

	testConfig["/some/log/level"] = "info"
	err = onlineconf_dev.ReopenWaiter(inst, "TREE", testConfig)
	require.NoError(t, err, "reopen waiter error")

	val, err = inst.GetString("/some/log/level")
	require.NoError(t, err, "can't get string")
	require.Equal(t, "info", val)

	testConfig["/some/foo"] = "bar"
	onlineconf_dev.GenerateCDB(tmpDir, "TREE", testConfig)
	time.Sleep(time.Microsecond * 100)

	testConfig["/some/log/level"] = "debug"
	err = onlineconf_dev.ReopenWaiter(inst, "TREE", testConfig)
	require.NoError(t, err, "reopen waiter error")

	val, err = inst.GetString("/some/log/level")
	require.NoError(t, err, "can't get string")
	require.Equal(t, "debug", val)
}

// newWatchedInstance creates an instance watching module TREE with conf in dir and
// returns it with the module already registered, so that ReopenWaiter can subscribe
// to it.
func newWatchedInstance(t *testing.T, dir string, conf map[string]any) *onlineconf.OnlineconfInstance {
	t.Helper()

	onlineconf_dev.GenerateCDB(dir, "TREE", conf)

	inst := onlineconf.Create(onlineconf.WithConfigDir(dir))
	require.NoError(t, inst.StartWatcher(context.Background()), "can't start watcher")

	t.Cleanup(func() { require.NoError(t, inst.StopWatcher(), "can't stop watcher") })

	val, err := inst.GetString("/some/log/level")
	require.NoError(t, err, "can't get string")
	require.Equal(t, "debug", val)

	return inst
}

// reopenLevel changes the log level of the instance through ReopenWaiter and checks that
// the instance serves the new value once ReopenWaiter has returned.
func reopenLevel(t *testing.T, inst *onlineconf.OnlineconfInstance, conf map[string]any, level string) {
	t.Helper()

	conf["/some/log/level"] = level

	require.NoError(t, onlineconf_dev.ReopenWaiter(inst, "TREE", conf), "reopen waiter error")

	val, err := inst.GetString("/some/log/level")
	require.NoError(t, err, "can't get string")
	require.Equal(t, level, val)
}

// TestReopenWaiterSeveralInstances is the regression test of LIB-6: ReopenWaiter kept the
// subscription of a module in a package global keyed by module name only. In a process
// with more than one instance (a repeated run of the same test with -count>1 among them)
// only the first instance ever got a subscription, and every later one timed out.
func TestReopenWaiterSeveralInstances(t *testing.T) {
	for _, tc := range []struct {
		name   string
		shared bool
	}{
		{name: "separate config dirs", shared: false},
		{name: "shared config dir", shared: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			firstDir := t.TempDir()

			secondDir := firstDir
			if !tc.shared {
				secondDir = t.TempDir()
			}

			firstConf := map[string]any{"/some/log/level": "debug"}
			first := newWatchedInstance(t, firstDir, firstConf)

			secondConf := map[string]any{"/some/log/level": "debug"}
			second := newWatchedInstance(t, secondDir, secondConf)

			// The second instance has to get a subscription of its own, even though a
			// module of that name is already watched by the first one.
			reopenLevel(t, second, secondConf, "info")

			// And the instance that subscribed first keeps working afterwards.
			reopenLevel(t, first, firstConf, "warn")

			// Both of them still work on a second round of reopens.
			reopenLevel(t, second, secondConf, "error")
			reopenLevel(t, first, firstConf, "info")
		})
	}
}
