package onlineconf

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/fsnotify/fsnotify"
)

type OnlineconfWatcher struct {
	sync.Mutex
	watcher   *fsnotify.Watcher
	cancelCtx context.CancelFunc
	doneChan  chan struct{}
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

	ow.doneChan = make(chan struct{})

	go func() {
		defer close(ow.doneChan)

		for {
			select {
			case ev := <-ow.watcher.Events:
				log.Printf("fsnotify event: %s", ev.String())
				callback(ev)
			case err := <-ow.watcher.Errors:
				log.Printf("fsnotify error: %s", err.Error())
				errorCallback(err)
			case <-watcherCtx.Done():
				log.Printf("fsnotify watcher stopped")
				return
			}
		}
	}()

	return nil
}

func (ow *OnlineconfWatcher) stopWatcher() error {
	if ow.doneChan == nil {
		return fmt.Errorf("can't stop inactive watcher")
	}

	ow.Lock()
	defer ow.Unlock()

	if ow.cancelCtx == nil {
		return fmt.Errorf("can't stop inactive watcher")
	}

	ow.cancelCtx()
	<-ow.doneChan
	ow.watcher = nil
	ow.cancelCtx = nil
	ow.doneChan = nil

	return nil
}
