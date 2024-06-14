package onlineconf_test

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/Educentr/go-onlineconf/pkg/onlineconf"
	"github.com/Educentr/go-onlineconf/pkg/onlineconf_dev"
	"github.com/stretchr/testify/assert"
)

func BenchmarkClone(b *testing.B) {
	tmpConfDir, _ := os.MkdirTemp("", "onlineconf*")
	var globalCtx, _ = onlineconf.Initialize(context.Background(), onlineconf.WithConfigDir(tmpConfDir))
	for i := 0; i < b.N; i++ {
		onlineconf.Clone(globalCtx, context.Background())
	}
}

func BenchmarkGetStringIfExistsCtx(b *testing.B) {
	tmpConfDir, _ := os.MkdirTemp("", "onlineconf*")
	var globalCtx, _ = onlineconf.Initialize(context.Background(), onlineconf.WithConfigDir(tmpConfDir))
	for i := 0; i < b.N; i++ {
		onlineconf.GetStringIfExists(globalCtx, "bla")
	}
}

func BenchmarkGetStringIfExistsDirect(b *testing.B) {
	tmpConfDir, _ := os.MkdirTemp("", "onlineconf*")
	inst := onlineconf.Create(onlineconf.WithConfigDir(tmpConfDir))
	m, _ := inst.GetOrAddModule("TREE")
	for i := 0; i < b.N; i++ {
		m.GetStringIfExists("bla")
	}
}

type callbackStatus struct {
	status bool
	sync.Mutex
}

func (c *callbackStatus) callback() error {
	c.Lock()
	defer c.Unlock()
	c.status = true
	return nil
}

func (c *callbackStatus) getStatus() bool {
	c.Lock()
	defer c.Unlock()
	return c.status
}

func TestGetDefaultModuleB(t *testing.T) {
	tmpConfDir, _ := os.MkdirTemp("", "onlineconf*")
	var globalCtx, _ = onlineconf.Initialize(context.Background(), onlineconf.WithConfigDir(tmpConfDir))

	onlineconf_dev.GenerateCDB(tmpConfDir, "TREE", map[string]interface{}{"bla": "blav"})

	err := onlineconf.StartWatcher(globalCtx)
	if err != nil {
		t.Errorf("can't start watcher: %s", err)
	}

	v, err := onlineconf.GetString(globalCtx, "bla")
	if err != nil {
		t.Error("error get string", err)
	}

	if v != "blav" {
		t.Error("invalid value", v)
	}

	status := callbackStatus{}

	onlineconf.RegisterSubscription(globalCtx, "TREE", []string{"bla"}, status.callback)

	newCtx := context.Background()
	newCtx, _ = onlineconf.Clone(globalCtx, newCtx)

	onlineconf_dev.GenerateCDB(tmpConfDir, "TREE", map[string]interface{}{"bla": "blav1"})
	time.Sleep(time.Millisecond * 100)

	if !status.getStatus() {
		t.Error("callback not called")
	}

	v, err = onlineconf.GetString(globalCtx, "bla")
	if err != nil {
		t.Error("error get string after update", err)
	}

	if v != "blav1" {
		t.Error("invalid value after update", v)
	}

	v, err = onlineconf.GetString(newCtx, "bla")
	if err != nil {
		t.Error("error get string", err)
	}

	if v != "blav" {
		t.Error("invalid value", v)
	}

	onlineconf.Release(globalCtx, newCtx)

	m := onlineconf.FromContext(newCtx).GetModule("TREE")
	if m != nil {
		t.Errorf("module exists after release")
	}

	if err = onlineconf.StopWatcher(globalCtx); err != nil {
		t.Errorf("can't sdtop watcher: %s", err)
	}
}

func TestOnlineconfInstance_GetOrAdd(t *testing.T) {
	tmpConfDir, _ := os.MkdirTemp("", "onlineconf*")
	// Create a new OnlineconfInstance
	oi := onlineconf.Create(onlineconf.WithConfigDir(tmpConfDir))

	onlineconf_dev.GenerateCDB(tmpConfDir, "testModule", map[string]interface{}{"bla": "sblav"})
	// Get or add a module by name
	module, err := oi.GetOrAddModule("testModule")

	// Assert that there is no error and the module is not nil
	assert.NoError(t, err)
	assert.NotNil(t, module)
}
