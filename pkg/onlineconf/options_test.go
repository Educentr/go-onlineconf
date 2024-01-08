package onlineconf

import (
	"testing"
)

func TestWithLogger(t *testing.T) {
	logger := &DefaultLogger{}
	option := WithLogger(logger)

	oi := &OnlineconfInstance{}
	option.apply(oi)

	if oi.logger != logger {
		t.Errorf("Expected logger to be set to %v, but got %v", logger, oi.logger)
	}
}

func TestWithConfigDir(t *testing.T) {
	path := "/path/to/config"
	option := WithConfigDir(path)

	oi := &OnlineconfInstance{}
	option.apply(oi)

	if oi.configDir != path {
		t.Errorf("Expected configDir to be set to %s, but got %s", path, oi.configDir)
	}
}
