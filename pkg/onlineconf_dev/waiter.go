package onlineconf_dev

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/Educentr/go-onlineconf/pkg/onlineconfInterface"
)

const (
	// waiterParam is the name of the parameter ReopenWaiter writes into the module to
	// recognize its own reopen. It is added to every configuration ReopenWaiter generates.
	waiterParam = "waiter"

	// waiterTimeout is how long ReopenWaiter waits for the module to be reopened with
	// the value it has written.
	waiterTimeout = 5 * time.Second
)

// errTimeout is returned by ReopenWaiter when the module has not been reopened with the
// value it has written within waiterTimeout.
var errTimeout = errors.New("timeout")

// errModuleGone is returned by the subscription callback when the module it was
// registered for is no longer part of the instance.
var errModuleGone = errors.New("module is not registered in the instance anymore")

// waiter is the state of the waiterParam subscription of one module of one onlineconf
// instance.
//
// A subscription cannot be unregistered, so exactly one is registered per instance and
// module, and every ReopenWaiter call on that pair reuses it.
type waiter struct {
	// call serializes the ReopenWaiter calls on the same instance and module: a single
	// waiterParam value cannot describe two reopens awaited at the same time.
	call sync.Mutex

	mu sync.Mutex
	// version is the waiterParam value the pending call waits for, 0 when no call is
	// waiting. Guarded by mu.
	version int32
	// ch is signalled by the subscription callback once the module has been reopened
	// with version. Buffered, so the callback never blocks on a call that has already
	// given up waiting.
	ch chan struct{}
}

// expect makes the subscription callback signal the waiter when the module is reopened
// with the waiterParam value version; version 0 stops it from signalling at all.
//
// A signal left over from an earlier reopen is dropped: it says nothing about the value
// the next call is about to write.
func (w *waiter) expect(version int32) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.version = version

	select {
	case <-w.ch:
	default:
	}
}

// notify signals the pending ReopenWaiter call if mod has just been reopened with the
// waiterParam value that call is waiting for.
func (w *waiter) notify(mod onlineconfInterface.Module) error {
	if mod == nil {
		return errModuleGone
	}

	v, ex, err := mod.GetIntIfExists(waiterParam)
	if err != nil {
		return fmt.Errorf("can't read %s parameter: %w", waiterParam, err)
	}

	if !ex {
		return nil
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.version == 0 || w.version != int32(v) {
		return nil
	}

	// A non-blocking send on a buffered channel: the callback is invoked by Reopen and
	// must not block it when the waiting call has already timed out, and a repeated
	// reopen with the same value adds nothing to the signal already queued.
	select {
	case w.ch <- struct{}{}:
	default:
	}

	return nil
}

// waiterKey identifies the waiter of a module of a particular instance. Keying by module
// name alone would reuse the waiter and the subscription of the first instance for every
// later instance with a module of that name, leaving those instances without a
// subscription that could ever signal their waiter.
type waiterKey struct {
	instance onlineconfInterface.Instance
	module   string
}

var reopenWaiter = struct {
	sync.Mutex
	waiters map[waiterKey]*waiter
}{
	waiters: map[waiterKey]*waiter{},
}

// getWaiter returns the waiter of the module of oc, registering the waiterParam
// subscription on the first call for that pair.
func getWaiter(oc onlineconfInterface.Instance, module string) (*waiter, error) {
	key := waiterKey{instance: oc, module: module}

	reopenWaiter.Lock()
	defer reopenWaiter.Unlock()

	if w, ex := reopenWaiter.waiters[key]; ex {
		return w, nil
	}

	w := &waiter{ch: make(chan struct{}, 1)}

	if err := oc.RegisterSubscription(module, []string{waiterParam}, func() error {
		return w.notify(oc.GetModule(module))
	}); err != nil {
		return nil, fmt.Errorf("can't register subscription: %w", err)
	}

	// The waiter is stored only after the subscription has been registered: an entry
	// left behind by a failed registration would be reused by every later call, which
	// would then wait for a signal nobody can send.
	reopenWaiter.waiters[key] = w

	return w, nil
}

// ReopenWaiter writes newConf into the CDB file of the module of oc and blocks until oc
// has reopened the module with it, instead of making the caller sleep until the watcher
// notices the new file. It is meant for tests and local development.
//
// A "waiter" entry is added to newConf (the map of the caller is modified) and is
// awaited through a change subscription, so oc has to be watching its config directory
// (see Instance.StartWatcher) and the module has to be registered in it already.
//
// Calls on the same instance and module are serialized. Every instance gets a subscription
// of its own, so instances with a module of the same name do not wait for each other's
// reopens. Instances sharing the config directory share the module file as well: a call
// on one of them changes the configuration of the others, and calls on such instances
// must not overlap, as they are not serialized and write the same file.
//
// It returns an error if the subscription cannot be registered, or "timeout" if the
// module has not been reopened with the written value within five seconds.
func ReopenWaiter(oc onlineconfInterface.Instance, module string, newConf map[string]any) error {
	w, err := getWaiter(oc, module)
	if err != nil {
		return err
	}

	w.call.Lock()
	defer w.call.Unlock()

	version := nonZeroInt31()
	newConf[waiterParam] = version

	w.expect(version)
	defer w.expect(0)

	GenerateCDB(oc.GetConfigDir(), module, newConf)

	timer := time.NewTimer(waiterTimeout)
	defer timer.Stop()

	select {
	case <-w.ch:
		return nil
	case <-timer.C:
		return errTimeout
	}
}

// nonZeroInt31 returns a random positive int32. The value a call waits for has to differ
// from the one currently stored in the module, otherwise the subscription sees no change
// and the callback is never invoked; 0 is reserved for "no call is waiting".
func nonZeroInt31() int32 {
	for {
		if v := rand.Int31(); v != 0 {
			return v
		}
	}
}
