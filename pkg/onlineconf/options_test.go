package onlineconf

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/Educentr/go-onlineconf/pkg/onlineconf_dev"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// usableModule is the name of a second, openable module used next to DefaultModule.
const usableModule = "OTHER"

func TestWithLogger(t *testing.T) {
	logger := &DefaultLogger{}
	option := WithLogger(logger)

	oi := &OnlineconfInstance{}
	option.Apply(oi)

	if oi.logger != logger {
		t.Errorf("Expected logger to be set to %v, but got %v", logger, oi.logger)
	}
}

func TestWithConfigDir(t *testing.T) {
	path := "/path/to/config"
	option := WithConfigDir(path)

	oi := &OnlineconfInstance{}
	option.Apply(oi)

	if oi.configDir != path {
		t.Errorf("Expected configDir to be set to %s, but got %s", path, oi.configDir)
	}
}

// recordedLog is a single call made to captureLogger.
type recordedLog struct {
	msg  string
	args []any
}

// captureLogger keeps the Error calls an instance makes so that a test can assert what was
// reported instead of what was printed.
type captureLogger struct {
	mu     sync.Mutex
	errors []recordedLog
}

func (l *captureLogger) Warn(_ context.Context, _ string, _ ...any) {}

func (l *captureLogger) Error(_ context.Context, msg string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.errors = append(l.errors, recordedLog{msg: msg, args: args})
}

func (l *captureLogger) Fatal(_ context.Context, _ string, _ ...any) {}

func (l *captureLogger) recordedErrors() []recordedLog {
	l.mu.Lock()
	defer l.mu.Unlock()

	return append([]recordedLog(nil), l.errors...)
}

// moduleOpenFailure is a config directory in which a module cannot be opened.
type moduleOpenFailure struct {
	name string
	// prepare leaves the module file of dir in the state the case is about.
	prepare func(t *testing.T, dir string)
	// errIs is what the error of the failed open must unwrap to.
	errIs error
}

// moduleOpenFailureCases are the two ways opening a module can fail: its file is missing, or
// it is there but cannot serve as a cdb (see ErrInvalidCDB).
func moduleOpenFailureCases() []moduleOpenFailure {
	return []moduleOpenFailure{
		{
			name:    "missing module file",
			prepare: func(_ *testing.T, _ string) {},
			errIs:   os.ErrNotExist,
		},
		{
			name: "unusable cdb file",
			prepare: func(t *testing.T, dir string) {
				t.Helper()
				replaceTreeCDB(t, dir, []byte(corruptCDBContent))
			},
			errIs: ErrInvalidCDB,
		},
	}
}

// openError returns the error a fresh instance gets for the default module of dir, so that a
// test can compare what WithModules reported against what the open actually failed with.
func openError(t *testing.T, dir string) error {
	t.Helper()

	_, err := Create(WithConfigDir(dir)).GetOrAddModule(DefaultModule)
	require.Error(t, err)

	return err
}

// A module that cannot be opened and was not declared required must not stop the
// initialization: the error is logged and the module is left unregistered.
func TestWithModulesOptionalSkipsUnopenableModule(t *testing.T) {
	for _, tc := range moduleOpenFailureCases() {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			tc.prepare(t, dir)

			wantErr := openError(t, dir)
			logger := &captureLogger{}

			var oi *OnlineconfInstance

			require.NotPanics(t, func() {
				oi = Create(WithConfigDir(dir), WithLogger(logger), WithModules([]string{DefaultModule}, false))
			}, "an optional module must not panic when it can't be opened")

			assert.Nil(t, oi.GetModule(DefaultModule), "a module that failed to open must not be registered")
			assert.Empty(t, oi.GetModuleNames())

			reported := logger.recordedErrors()
			require.Len(t, reported, 1, "the error of a skipped module must be reported once")
			assert.Contains(t, reported[0].msg, DefaultModule, "the report must name the module")

			require.Len(t, reported[0].args, 1)
			loggedErr, ok := reported[0].args[0].(error)
			require.True(t, ok, "the error itself must be passed to the logger, got %+v", reported[0].args[0])
			assert.ErrorIs(t, loggedErr, tc.errIs)
			assert.EqualError(t, loggedErr, wantErr.Error())
		})
	}
}

// Skipping a module must not cost the names that come after it in the list.
func TestWithModulesOptionalKeepsOpeningTheRestOfTheList(t *testing.T) {
	dir := t.TempDir()
	replaceTreeCDB(t, dir, []byte(corruptCDBContent))
	onlineconf_dev.GenerateCDB(dir, usableModule, map[string]interface{}{cdbTestKey: cdbNewValue})

	logger := &captureLogger{}

	var oi *OnlineconfInstance

	require.NotPanics(t, func() {
		oi = Create(
			WithConfigDir(dir),
			WithLogger(logger),
			WithModules([]string{DefaultModule, usableModule}, false),
		)
	})

	assert.Nil(t, oi.GetModule(DefaultModule))
	require.NotNil(t, oi.GetModule(usableModule), "the module after the skipped one must be opened")
	assert.Equal(t, []string{usableModule}, oi.GetModuleNames())

	require.Len(t, logger.recordedErrors(), 1, "only the module that failed to open may be reported")
}

// A module declared required keeps panicking when it can't be opened, and the panic has to
// say which module it was and what the open failed with.
func TestWithModulesRequiredPanicsOnUnopenableModule(t *testing.T) {
	for _, tc := range moduleOpenFailureCases() {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			tc.prepare(t, dir)

			wantErr := openError(t, dir)
			logger := &captureLogger{}

			recovered := func() (recovered any) {
				defer func() { recovered = recover() }()

				Create(WithConfigDir(dir), WithLogger(logger), WithModules([]string{DefaultModule}, true))

				return nil
			}()

			require.NotNil(t, recovered, "a required module that can't be opened must panic")

			msg, ok := recovered.(string)
			require.True(t, ok, "expected a string panic value, got %T: %+v", recovered, recovered)
			assert.Contains(t, msg, DefaultModule, "the panic must name the module")
			assert.Contains(t, msg, wantErr.Error(), "the panic must carry the error of the failed open")
		})
	}
}

// An openable module is opened and registered whatever required says, and nothing is
// reported for it.
func TestWithModulesOpensUsableModule(t *testing.T) {
	for _, required := range []bool{false, true} {
		t.Run(fmt.Sprintf("required=%t", required), func(t *testing.T) {
			dir := t.TempDir()
			onlineconf_dev.GenerateCDB(dir, DefaultModule, map[string]interface{}{cdbTestKey: cdbNewValue})

			logger := &captureLogger{}

			var oi *OnlineconfInstance

			require.NotPanics(t, func() {
				oi = Create(WithConfigDir(dir), WithLogger(logger), WithModules([]string{DefaultModule}, required))
			})

			require.NotNil(t, oi.GetModule(DefaultModule))
			assert.Equal(t, []string{DefaultModule}, oi.GetModuleNames())
			assert.Empty(t, logger.recordedErrors(), "nothing may be reported for a module that opened")

			v, err := oi.GetString(cdbTestKey, "")
			require.NoError(t, err)
			assert.Equal(t, cdbNewValue, v)
		})
	}
}
