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
	inst.StartWatcher(context.Background())

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
