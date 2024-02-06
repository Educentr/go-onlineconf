package onlineconf_dev

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/Nikolo/go-onlineconf/pkg/onlineconfInterface"
)

type waiter struct {
	ch      chan struct{}
	version int
}

var reopenWaiter = struct {
	sync.Mutex
	waiters map[string]*waiter
}{
	waiters: map[string]*waiter{},
}

func ReopenWaiter(oc onlineconfInterface.Instance, module string, newConf map[string]any) error {
	reopenWaiter.Lock()

	if _, ex := reopenWaiter.waiters[module]; !ex {
		reopenWaiter.waiters[module] = &waiter{ch: make(chan struct{})}

		oc.RegisterSubscription(module, []string{"waiter"}, func() error {
			reopenWaiter.Lock()
			defer reopenWaiter.Unlock()

			v, ex, _ := oc.GetModule(module).GetIntIfExists("waiter")
			if !ex {
				return nil
			}

			if reopenWaiter.waiters[module].version == v && len(reopenWaiter.waiters[module].ch) == 0 {
				reopenWaiter.waiters[module].ch <- struct{}{}
			}

			return nil
		})
	}

	waiterInt := rand.Int()
	newConf["waiter"] = waiterInt

	if reopenWaiter.waiters[module].version == 0 {
		if len(reopenWaiter.waiters[module].ch) != 0 {
			<-reopenWaiter.waiters[module].ch
		}

		reopenWaiter.waiters[module].version = waiterInt
		defer func() { reopenWaiter.waiters[module].version = 0 }()
	}

	GenerateCDB(oc.GetConfigDir(), module, newConf)

	reopenWaiter.Unlock()

	timer := time.NewTimer(5 * time.Second)

	select {
	case <-reopenWaiter.waiters[module].ch:
		return nil
	case <-timer.C:
		return fmt.Errorf("timeout")

	}
}
