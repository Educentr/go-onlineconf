package onlineconf

import (
	"testing"

	"github.com/stretchr/testify/assert"
	testCDB "gitlab.educentr.info/godev/onlinecof-test/test"
)

func TestOnlineconfInstance_RegisterSubscription(t *testing.T) {
	// Create a new OnlineconfInstance
	oi := Create()

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

	if oi.byName[module].changeSubscription[0].path[0] != "param1" {
		assert.Fail(t, "Unexpected subscription path")
	}
}

func TestOnlineconfInstance_Get(t *testing.T) {
	// Create a new OnlineconfInstance
	oi := Create()

	oi.byName["testModule"] = &Module{}

	// Get a module by name
	module := oi.Get("testModule")

	// Assert that the module is not nil
	assert.NotNil(t, module)
}

func TestOnlineconfInstance_GetOrAdd(t *testing.T) {
	d := t.TempDir()
	// Create a new OnlineconfInstance
	oi := Create(WithConfigDir(d))

	testCDB.Generate(d+"/testModule.cdb", map[string][]byte{"bla": []byte("sblav")})
	// Get or add a module by name
	module, err := oi.GetOrAdd("testModule")

	// Assert that there is no error and the module is not nil
	assert.NoError(t, err)
	assert.NotNil(t, module)
}

// Add more test cases for other functions in the OnlineconfInstance struct
