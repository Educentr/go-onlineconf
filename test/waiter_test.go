package onlineconf_test

import (
	"context"
	"os"
	"testing"

	"github.com/Nikolo/go-onlineconf/pkg/onlineconf"
	"github.com/Nikolo/go-onlineconf/pkg/onlineconf_dev"
	"github.com/stretchr/testify/require"
)

func TestReopenWaiter(t *testing.T) {
	tmpLogFile := "/tmp/onlineconfLogger.log"
	tmpDir := "/tmp/onlineconf"

	os.Remove(tmpLogFile)

	testConfig := map[string]any{"/some/log/level": "debug", "/some/log/output/err": tmpLogFile, "/some/log/output/out": tmpLogFile}

	onlineconf_dev.GenerateCDB(tmpDir, "TREE", testConfig)

	inst := onlineconf.Create(onlineconf.WithConfigDir(tmpDir))
	inst.StartWatcher(context.Background())

	val, err := inst.GetString("/some/log/level")
	require.NoError(t, err, "can't get string")
	require.Equal(t, val, "debug")

	testConfig["/some/log/level"] = "info"
	onlineconf_dev.ReopenWaiter(inst, "TREE", testConfig)

	val, err = inst.GetString("/some/log/level")
	require.NoError(t, err, "can't get string")
	require.Equal(t, val, "info")

	testConfig["/some/foo"] = "bar"
	onlineconf_dev.GenerateCDB(tmpDir, "TREE", testConfig)

	testConfig["/some/log/level"] = "debug"
	onlineconf_dev.ReopenWaiter(inst, "TREE", testConfig)

	val, err = inst.GetString("/some/log/level")
	require.NoError(t, err, "can't get string")
	require.Equal(t, val, "debug")
}
