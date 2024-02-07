package onlineconf

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Nikolo/go-onlineconf/pkg/onlineconf_dev"
	"github.com/stretchr/testify/assert"
)

const tmpConfDir = "/tmp/onlineconf/"

var _ = os.Mkdir(tmpConfDir, os.ModePerm)

func TestOnlineconfInstance_RegisterSubscription(t *testing.T) {
	// Create a new OnlineconfInstance
	oi, ok := Create().(*OnlineconfInstance)
	if !ok {
		t.Errorf("Unexpected instance: %v", oi)
	}

	// Define the module name, parameters, and callback function
	module := "testModule"

	oi.byName[module] = &Module{}

	params := []string{"param1", "param2"}
	callback := func() error {
		// Callback implementation
		return nil
	}

	// Register the subscription
	err := oi.RegisterSubscription(module, params, callback)

	// Assert that there is no error
	assert.NoError(t, err)

	oiModule, ok := oi.byName[module].(*Module)
	if !ok {
		t.Errorf("Unexpected module: %v", oi.byName[module])
	}

	subscription, ok := oiModule.changeSubscription[0].(*SubscriptionCallback)
	if !ok {
		t.Errorf("Unexpected subscription: %v", oiModule.changeSubscription[0])
	}

	if subscription.path[0] != "param1" {
		assert.Fail(t, "Unexpected subscription path")
	}
}

func TestOnlineconfInstance_Get(t *testing.T) {
	// Create a new OnlineconfInstance
	oi, ok := Create().(*OnlineconfInstance)
	if !ok {
		t.Errorf("Unexpected instance: %v", oi)
	}

	oi.byName["testModule"] = &Module{}

	// Get a module by name
	module := oi.GetModule("testModule")

	// Assert that the module is not nil
	assert.NotNil(t, module)
}

func TestGetDefaultModuleW(t *testing.T) {
	var globalCtx, _ = Initialize(context.Background(), WithConfigDir(tmpConfDir))

	onlineconf_dev.GenerateCDB(tmpConfDir, "TREE", map[string]interface{}{"bla": "blav"})
	err := StartWatcher(globalCtx)
	if err != nil {
		t.Errorf("can't start watcher: %s", err)
	}

	v, err := GetString(globalCtx, "bla")
	if err != nil {
		t.Error("error get string", err)
	}

	if v != "blav" {
		t.Error("invalid value", v)
	}

	newCtx := context.Background()
	newCtx, _ = Clone(globalCtx, newCtx)

	err = onlineconf_dev.ReopenWaiter(FromContext(globalCtx), "TREE", map[string]interface{}{"bla": "blav1"})
	if err != nil {
		t.Errorf("can't reopen waiter: %s", err)
	}

	v, err = GetString(globalCtx, "bla")
	if err != nil {
		t.Error("error get string after update", err)
	}

	if v != "blav1" {
		t.Error("invalid value after update", v)
	}

	v, err = GetString(newCtx, "bla")
	if err != nil {
		t.Error("error get string", err)
	}

	if v != "blav" {
		t.Error("invalid value", v)
	}

	err = Release(globalCtx, newCtx)
	if err != nil {
		t.Errorf("error release: %s", err)
	}

	instance := FromContext(newCtx)
	m := instance.GetModule("TREE")
	if m != nil {
		t.Errorf("module exists after release")
	}

	instance = FromContext(globalCtx)
	if len(instance.mmappedFiles) != 1 {
		t.Errorf("invalid mmaped size: %d", len(instance.mmappedFiles))
	}

	if err = StopWatcher(globalCtx); err != nil {
		t.Errorf("can't sdtop watcher: %s", err)
	}

	if instance.watcher.watcher != nil {
		t.Errorf("watcher exists after stop")
	}
}

func TestGetNonDefaultModuleW(t *testing.T) {
	var globalCtx, _ = Initialize(context.Background(), WithConfigDir(tmpConfDir))

	onlineconf_dev.GenerateCDB(tmpConfDir, "module3", map[string]interface{}{"bla": "blav"})
	err := StartWatcher(globalCtx)
	if err != nil {
		t.Errorf("can't start watcher: %s", err)
	}

	m, err := GetOrAddModule(globalCtx, "module3")
	if err != nil {
		t.Error("error get string", err)
	}

	v, err := m.GetString("bla")
	if err != nil {
		t.Error("error get string", err)
	}

	if v != "blav" {
		t.Error("invalid value", v)
	}

	newCtx := context.Background()
	newCtx, _ = Clone(globalCtx, newCtx)

	onlineconf_dev.GenerateCDB(tmpConfDir, "module3", map[string]interface{}{"bla": "blav1"})
	time.Sleep(time.Millisecond * 100)

	v, err = m.GetString("bla")
	if err != nil {
		t.Error("error get string after update", err)
	}

	if v != "blav1" {
		t.Error("invalid value after update", v)
	}

	instance := FromContext(newCtx)

	mNew := instance.GetModule("module3")

	v, err = mNew.GetString("bla")
	if err != nil {
		t.Error("error get string", err)
	}

	if v != "blav" {
		t.Error("invalid value", v)
	}

	err = Release(globalCtx, newCtx)
	if err != nil {
		t.Errorf("error release: %s", err)
	}

	m = instance.GetModule("module3")
	if m != nil {
		t.Errorf("module exists after release")
	}

	instance = FromContext(globalCtx)
	if len(instance.mmappedFiles) != 1 {
		t.Errorf("invalid mmaped size: %d", len(instance.mmappedFiles))
	}

	if err = StopWatcher(globalCtx); err != nil {
		t.Errorf("can't sdtop watcher: %s", err)
	}

	if instance.watcher.watcher != nil {
		t.Errorf("watcher exists after stop")
	}
}

func TestGetNonDefaultModuleDirectW(t *testing.T) {
	var globalCtx, _ = Initialize(context.Background(), WithConfigDir(tmpConfDir))

	onlineconf_dev.GenerateCDB(tmpConfDir, "module4", map[string]interface{}{"bla": "blav"})
	instance := FromContext(globalCtx)

	m, err := instance.GetOrAddModule("module4")
	if err != nil {
		t.Errorf("Error while geting module: %s\n", err)
		return
	}

	v, err := m.GetString("bla")
	if err != nil {
		t.Errorf("Error while geting param: %s\n", err)
		return
	}

	if v != "blav" {
		t.Errorf("invalid value: %s\n", v)
	}
}
