package onlineconf

import (
	"context"
	"fmt"
	"sync"

	"github.com/fsnotify/fsnotify"
)

type OnlineconfWatcher struct {
	sync.Mutex
	watcher   *fsnotify.Watcher
	cancelCtx context.CancelFunc
	path      string
}

func (ow *OnlineconfWatcher) start(path string, callback func(fsnotify.Event), errorCallback func(error)) error {
	ow.Lock()
	defer ow.Unlock()

	if ow.watcher != nil {
		if ow.path == path {
			return nil
		}

		return fmt.Errorf("watcher already inited on the another folder")
	}

	var err error

	ow.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("can't create fsnotify watcher: %w", err)
	}

	err = ow.watcher.Add(path)
	if err != nil {
		return fmt.Errorf("can't add dir to fsnotify watcher: %w", err)
	}

	ow.path = path

	var watcherCtx context.Context

	watcherCtx, ow.cancelCtx = context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case ev := <-ow.watcher.Events:
				callback(ev)
			case err := <-ow.watcher.Errors:
				errorCallback(err)
			case <-watcherCtx.Done():
				return
			}
		}
	}()

	return nil
}

func (ow *OnlineconfWatcher) stopWatcher() error {
	ow.Lock()
	defer ow.Unlock()

	if ow.cancelCtx == nil {
		return fmt.Errorf("can't stop inactive watcher")
	}

	ow.cancelCtx()
	ow.watcher = nil
	ow.cancelCtx = nil

	return nil
}
