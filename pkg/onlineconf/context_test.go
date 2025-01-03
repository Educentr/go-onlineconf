package onlineconf

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/Educentr/go-onlineconf/pkg/onlineconfInterface"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/mmap"
)

func TestFromContext(t *testing.T) {
	// Create a dummy OnlineconfInstance
	oi := &OnlineconfInstance{}

	// Create a context with the OnlineconfInstance as a value
	ctx := context.WithValue(context.Background(), ContextOnlineconfName, oi)

	// Call the FromContext function
	result := FromContext(ctx)

	// Check if the result is equal to the dummy OnlineconfInstance
	if result != oi {
		t.Errorf("Expected %v, but got %v", oi, result)
	}

	// Create a context without the OnlineconfInstance as a value
	emptyCtx := context.Background()

	// Call the FromContext function with the empty context
	emptyResult := FromContext(emptyCtx)

	// Check if the result is nil
	if emptyResult != nil {
		t.Errorf("Expected nil, but got %v", emptyResult)
	}
}
func TestCloneRelease(t *testing.T) {
	oneOfMmapedFile := &mmap.ReaderAt{}
	oneOfMmapedFileAddr := fmt.Sprintf("%p", oneOfMmapedFile)
	// Create a dummy OnlineconfInstance
	oi := &OnlineconfInstance{
		ro:           false,
		logger:       nil,
		byName:       map[string]onlineconfInterface.Module{"module1": &Module{name: "module1", filename: "/etc/onlineconf/module1.conf", mmappedFile: oneOfMmapedFile}, "module2": &Module{name: "module2", filename: "/etc/onlineconf/module2.conf", mmappedFile: &mmap.ReaderAt{}}},
		byFile:       map[string]onlineconfInterface.Module{"/etc/onlineconf/module1.conf": &Module{name: "module1", filename: "/etc/onlineconf/module1.conf", mmappedFile: oneOfMmapedFile}, "/etc/onlineconf/module2.conf": &Module{name: "module2", filename: "/etc/onlineconf/module2.conf", mmappedFile: &mmap.ReaderAt{}}},
		names:        []string{"module1", "module2"},
		mmappedFiles: map[string]*mmapedFiles{oneOfMmapedFileAddr: {reader: &mmap.ReaderAt{}, refcount: 1}},
	}

	// Create a context with the OnlineconfInstance as a value
	ctx := context.WithValue(context.Background(), ContextOnlineconfName, oi)

	// Create a new context for cloning
	cloneCtx := context.Background()

	// Call the Clone function
	newCtx, err := Clone(ctx, cloneCtx)

	// Check if the cloning was successful
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if oi.mmappedFiles[oneOfMmapedFileAddr].refcount != 2 {
		t.Error("Expected refcount to be 2 but got", oi.mmappedFiles["module1"].refcount)
	}

	// Check if the cloned context has the correct value
	clonedInstance := FromContext(newCtx)
	if clonedInstance == nil || clonedInstance.ro != true || clonedInstance.logger != nil || len(clonedInstance.byName) != 2 {
		t.Errorf("Unexpected cloned instance: %v", clonedInstance)
	}

	// Check if the cloned context has the correct modules
	for _, name := range oi.names {
		moduleI := clonedInstance.GetModule(name)
		module, ok := moduleI.(*Module)
		if !ok {
			t.Errorf("Unexpected module: %v", moduleI)
		}

		oiModuleI := oi.GetModule(name)
		oiModule, ok := oiModuleI.(*Module)
		if !ok {
			t.Errorf("Unexpected module: %v", oiModuleI)
		}

		if module == nil || module.ro != true || module.name != name || module.filename != oiModule.filename {
			t.Errorf("Unexpected cloned module: %v", module)
		}
	}

	err = Release(ctx, newCtx)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if oi.mmappedFiles[oneOfMmapedFileAddr].refcount != 1 {
		t.Error("Expected refcount to be 1")
	}
}
func TestStopWatcher(t *testing.T) {
	// Create a context without the OnlineconfInstance as a value
	emptyCtx := context.Background()

	// Call the StopWatcher function with the empty context
	err := StopWatcher(emptyCtx)

	// Check if the error is not nil
	if err == nil {
		t.Errorf("Expected error, but got nil")
	}
}
func TestRegisterCallback(t *testing.T) {
	module := "testModule"

	// Create a dummy OnlineconfInstance
	oi := &OnlineconfInstance{
		byName: map[string]onlineconfInterface.Module{module: &Module{name: module, filename: "/etc/onlineconf/" + module + ".conf", mmappedFile: &mmap.ReaderAt{}}},
	}

	// Create a context with the OnlineconfInstance as a value
	ctx := context.WithValue(context.Background(), ContextOnlineconfName, oi)

	// Define the test module and parameters
	params := []string{"param1", "param2"}

	// Define the test callback function
	callback := func() error {
		// Test callback logic here
		return nil
	}

	// Call the RegisterCallback function
	err := RegisterSubscription(ctx, module, params, callback)

	// Check if the error is nil
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	// Check if the subscription was registered correctly
	oiModuleI := oi.GetModule(module)
	oiModule, ok := oiModuleI.(*Module)
	if !ok {
		t.Errorf("Unexpected module: %v", oiModuleI)
	}

	subscriptionI := oiModule.changeSubscription[0]
	subscription, ok := subscriptionI.(*SubscriptionCallback)
	if !ok {
		t.Errorf("Unexpected subscription: %v", subscriptionI)
	}

	if subscription.path == nil || subscription.callback == nil {
		t.Errorf("Unexpected subscription: %v", subscription)
	}
}
func TestStartWatcher(t *testing.T) {
	// Create a dummy OnlineconfInstance
	oi := &OnlineconfInstance{}

	// Create a context with the OnlineconfInstance as a value
	ctx := context.WithValue(context.Background(), ContextOnlineconfName, oi)

	// Call the StartWatcher function
	err := StartWatcher(ctx)

	// Check if the error is nil
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Call the StopWatcher function
	err = StopWatcher(ctx)

	// Check if the error is nil
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestInitialize(t *testing.T) {
	// Create a context
	ctx := context.Background()

	// Call the Initialize function
	newCtx, err := Initialize(ctx)

	// Check if the initialization was successful
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	// Check if the new context has the OnlineconfInstance value
	oi := FromContext(newCtx)
	if oi == nil {
		t.Error("Expected OnlineconfInstance value in the context, but got nil")
	}
}

func TestEnvConfig(t *testing.T) {
	ctx := context.Background()

	tmpConfDir, _ := os.MkdirTemp("", "onlineconf*")
	defer os.RemoveAll(tmpConfDir)

	// Call the EnvConfig function
	os.Setenv("ONLINECONFIG_FROM_ENV", "true")
	os.Setenv("OC_TEST", "test")
	os.Setenv("OC_TEST__TEST", "test2")

	// Call the Initialize function
	newCtx, err := Initialize(ctx, WithConfigDir(tmpConfDir))
	assert.NoError(t, err)

	testParam, err := GetString(newCtx, "/TEST", "")
	assert.NoError(t, err)
	assert.Equal(t, "test", testParam)

	test2Param, err := GetString(newCtx, "/TEST/TEST", "")
	assert.NoError(t, err)
	assert.Equal(t, "test2", test2Param)
}
