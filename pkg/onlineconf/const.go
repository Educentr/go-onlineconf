package onlineconf

import (
	"context"
	"errors"
	"sync"

	"github.com/colinmarc/cdb"
	"golang.org/x/exp/mmap"
)

const startCacheSize = 100
const startModuleCountSize = 100

var ErrFormatIsNotJSON = errors.New("format is not JSON")
var ErrAddModule = errors.New("add module error")
var ErrUnavailableInRO = errors.New("unable to use in readonly instance")
var ErrUnavailableModifyRefcountRO = errors.New("can't modify refcount in RO instance")

const defaultConfigDir = "/usr/local/etc/onlineconf"
const DefaultModule = "TREE"

type OnlineconfLogger interface {
	Warn(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
	Fatal(ctx context.Context, msg string, args ...any)
}

type mmapedFiles struct {
	reader   *mmap.ReaderAt
	refcount uint
}

type OnlineconfInstance struct {
	sync.Mutex
	logger       OnlineconfLogger
	ro           bool
	names        []string
	byName       map[string]*Module
	byFile       map[string]*Module
	watcher      OnlineconfWatcher
	configDir    string
	mmappedFiles map[string]*mmapedFiles
}

type SubscriptionCallback struct {
	path     []string
	callback func() error
}

type Module struct {
	sync.RWMutex
	ro                 bool
	name               string
	filename           string
	cdb                *cdb.CDB
	cache              map[string][]interface{}
	cacheMutex         sync.RWMutex
	mmappedFile        *mmap.ReaderAt
	changeSubscription []SubscriptionCallback
}
