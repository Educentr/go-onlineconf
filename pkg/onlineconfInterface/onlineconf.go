package onlineconfInterface

import (
	"context"
	"time"

	"golang.org/x/exp/mmap"
)

type Option interface {
	Apply(oi Instance)
}

type Logger interface {
	Warn(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
	Fatal(ctx context.Context, msg string, args ...any)
}

type SubscriptionCallback interface {
	GetPaths() []string
	InvokeCallback() error
}

type Module interface {
	GetStringIfExists(string) (string, bool, error)
	GetIntIfExists(string) (int64, bool, error)
	GetDurationIfExists(string) (time.Duration, bool, error)
	GetBoolIfExists(string) (bool, bool, error)
	GetString(string, ...string) (string, error)
	GetInt(string, ...int64) (int64, error)
	GetDuration(string, ...time.Duration) (time.Duration, error)
	GetBool(string, ...bool) (bool, error)
	GetStrings(string, []string) ([]string, error)
	GetStruct(string, interface{}) (bool, error)
	RegisterSubscription(subscription SubscriptionCallback)
	Reopen(mmappedFile *mmap.ReaderAt) (*mmap.ReaderAt, error)
	GetMmappedFile() *mmap.ReaderAt
	Clone(name string) Module
}

type Instance interface {
	GetConfigDir() string
	RegisterSubscription(string, []string, func() error) error
	StartWatcher(ctx context.Context) error
	StopWatcher() error
	GetModuleByFile(string) (Module, bool)
	GetModule(string) Module
	GetModuleNames() []string
	GetOrAddModule(string) (Module, error)
	GetStringIfExists(string) (string, bool, error)
	GetIntIfExists(string) (int64, bool, error)
	GetDurationIfExists(string) (time.Duration, bool, error)
	GetBoolIfExists(string) (bool, bool, error)
	GetString(string, ...string) (string, error)
	GetInt(string, ...int64) (int64, error)
	GetDuration(string, ...time.Duration) (time.Duration, error)
	GetBool(string, ...bool) (bool, error)
	GetStrings(string, []string) ([]string, error)
	GetStruct(string, interface{}) (bool, error)
	Clone() (Instance, error)
	Release(context.Context, Instance) error
}
