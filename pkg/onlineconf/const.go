package onlineconf

import (
	"errors"
	"sync"

	"github.com/Educentr/go-onlineconf/pkg/onlineconfInterface"
	"github.com/colinmarc/cdb"
	"golang.org/x/exp/mmap"
)

const startCacheSize = 100
const startModuleCountSize = 100

var ErrFormatIsNotJSON = errors.New("format is not JSON")
var ErrAddModule = errors.New("add module error")
var ErrUnavailableInRO = errors.New("unable to use in readonly instance")
var ErrUnavailableModifyRefcountRO = errors.New("can't modify refcount in RO instance")

// ErrInvalidCDB is reported when a file is not usable as a CDB database: it is either not a
// CDB at all or an incomplete one, e.g. left by an interrupted rebuild. Such a file is
// rejected instead of being adopted by a module, so the errors a caller can get it from are
// Reopen and GetOrAddModule.
var ErrInvalidCDB = errors.New("invalid cdb file")

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
