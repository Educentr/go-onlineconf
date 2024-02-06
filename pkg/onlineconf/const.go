package onlineconf

import (
	"errors"
	"sync"

	"github.com/Nikolo/go-onlineconf/pkg/onlineconfInterface"
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

type mmapedFiles struct {
	reader   *mmap.ReaderAt
	refcount uint
}

type OnlineconfInstance struct {
	sync.Mutex
	logger       onlineconfInterface.Logger
	ro           bool
	names        []string
	byName       map[string]onlineconfInterface.Module
	byFile       map[string]onlineconfInterface.Module
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
	changeSubscription []onlineconfInterface.SubscriptionCallback
}
