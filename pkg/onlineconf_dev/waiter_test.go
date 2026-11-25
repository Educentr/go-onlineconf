package onlineconf_dev //nolint:revive // the underscore comes from the public import path of the package

import (
	"errors"
	"testing"

	"github.com/Educentr/go-onlineconf/pkg/onlineconfInterface"
)

// fakeModule reports a fixed "waiter" parameter. The embedded interface covers the
// methods the waiter never calls.
type fakeModule struct {
	onlineconfInterface.Module

	value int64
	ex    bool
	err   error
}

func (m fakeModule) GetIntIfExists(string) (int64, bool, error) {
	return m.value, m.ex, m.err
}

// fakeInstance counts the subscriptions registered on it. The embedded interface covers
// the methods getWaiter never calls.
type fakeInstance struct {
	onlineconfInterface.Instance

	registered []string
	err        error
}

func (i *fakeInstance) RegisterSubscription(module string, _ []string, _ func() error) error {
	if i.err != nil {
		return i.err
	}

	i.registered = append(i.registered, module)

	return nil
}

func newWaiter() *waiter {
	return &waiter{ch: make(chan struct{}, 1)}
}

func signalled(w *waiter) bool {
	select {
	case <-w.ch:
		return true
	default:
		return false
	}
}

func TestWaiterNotifySignalsExpectedValue(t *testing.T) {
	w := newWaiter()
	w.expect(42)

	if err := w.notify(fakeModule{value: 42, ex: true}); err != nil {
		t.Fatalf("notify: %s", err)
	}

	if !signalled(w) {
		t.Error("waiter was not signalled for the expected value")
	}
}

func TestWaiterNotifyIgnoresOtherValues(t *testing.T) {
	for name, mod := range map[string]fakeModule{
		"another value":   {value: 41, ex: true},
		"no waiter param": {ex: false},
	} {
		t.Run(name, func(t *testing.T) {
			w := newWaiter()
			w.expect(42)

			if err := w.notify(mod); err != nil {
				t.Fatalf("notify: %s", err)
			}

			if signalled(w) {
				t.Error("waiter was signalled for a value it does not wait for")
			}
		})
	}
}

// TestWaiterNotifyDoesNotBlockWithoutWaiter covers the callback of a module nobody waits
// for: it is invoked by Module.Reopen and may not block it, neither when no call is
// waiting nor when a reopen repeats a value already signalled.
func TestWaiterNotifyDoesNotBlockWithoutWaiter(t *testing.T) {
	w := newWaiter()

	// Nobody is waiting: a reopen still invokes the callback.
	if err := w.notify(fakeModule{value: 42, ex: true}); err != nil {
		t.Fatalf("notify without a waiter: %s", err)
	}

	if signalled(w) {
		t.Error("waiter was signalled while no call was waiting")
	}

	// The call that waited for 42 has given up, and the same value is reopened twice.
	w.expect(42)

	for i := 0; i < 3; i++ {
		if err := w.notify(fakeModule{value: 42, ex: true}); err != nil {
			t.Fatalf("notify %d: %s", i, err)
		}
	}

	if !signalled(w) {
		t.Error("waiter was not signalled for the expected value")
	}
}

// TestWaiterExpectDropsStaleSignal covers a call that starts after an unrelated reopen
// has already signalled the channel: the leftover signal says nothing about the value
// the new call is about to write.
func TestWaiterExpectDropsStaleSignal(t *testing.T) {
	w := newWaiter()
	w.expect(42)

	if err := w.notify(fakeModule{value: 42, ex: true}); err != nil {
		t.Fatalf("notify: %s", err)
	}

	w.expect(0)

	if signalled(w) {
		t.Error("stale signal was kept after the waiting call had finished")
	}

	w.expect(43)

	if signalled(w) {
		t.Error("stale signal was kept for the next waiting call")
	}
}

func TestWaiterNotifyModuleErrors(t *testing.T) {
	w := newWaiter()
	w.expect(42)

	if err := w.notify(nil); !errors.Is(err, errModuleGone) {
		t.Errorf("notify of a missing module: expected %s, got %v", errModuleGone, err)
	}

	errRead := errors.New("broken cdb")

	if err := w.notify(fakeModule{err: errRead}); !errors.Is(err, errRead) {
		t.Errorf("notify of an unreadable module: expected %s, got %v", errRead, err)
	}

	if signalled(w) {
		t.Error("waiter was signalled on a module error")
	}
}

// TestGetWaiterPerInstance is the regression test of LIB-6: the waiter and its
// subscription belong to a pair of an instance and a module, not to the module name, so
// a second instance with a module of the same name gets a subscription of its own.
func TestGetWaiterPerInstance(t *testing.T) {
	first := &fakeInstance{}
	second := &fakeInstance{}

	firstWaiter, err := getWaiter(first, "TREE")
	if err != nil {
		t.Fatalf("get waiter of the first instance: %s", err)
	}

	sameWaiter, err := getWaiter(first, "TREE")
	if err != nil {
		t.Fatalf("get waiter of the first instance again: %s", err)
	}

	if sameWaiter != firstWaiter {
		t.Error("the same instance and module got a second waiter")
	}

	secondWaiter, err := getWaiter(second, "TREE")
	if err != nil {
		t.Fatalf("get waiter of the second instance: %s", err)
	}

	if secondWaiter == firstWaiter {
		t.Error("the second instance reused the waiter of the first one")
	}

	otherModuleWaiter, err := getWaiter(first, "OTHER")
	if err != nil {
		t.Fatalf("get waiter of another module: %s", err)
	}

	if otherModuleWaiter == firstWaiter {
		t.Error("another module of the same instance reused its waiter")
	}

	if len(first.registered) != 2 {
		t.Errorf("expected 2 subscriptions on the first instance, got %v", first.registered)
	}

	if len(second.registered) != 1 {
		t.Errorf("expected 1 subscription on the second instance, got %v", second.registered)
	}
}

// TestGetWaiterRegistrationError covers a failed registration: the waiter may not be
// cached, otherwise every later call would wait for a signal nobody can send.
func TestGetWaiterRegistrationError(t *testing.T) {
	errRegister := errors.New("module does not exists")
	inst := &fakeInstance{err: errRegister}

	if _, err := getWaiter(inst, "TREE"); !errors.Is(err, errRegister) {
		t.Fatalf("expected %s, got %v", errRegister, err)
	}

	inst.err = nil

	if _, err := getWaiter(inst, "TREE"); err != nil {
		t.Fatalf("get waiter after a successful registration: %s", err)
	}

	if len(inst.registered) != 1 {
		t.Errorf("expected 1 subscription after the retry, got %v", inst.registered)
	}
}

func TestNonZeroInt31(t *testing.T) {
	for i := 0; i < 1000; i++ {
		if v := nonZeroInt31(); v == 0 {
			t.Fatal("0 is reserved for \"no call is waiting\"")
		}
	}
}
