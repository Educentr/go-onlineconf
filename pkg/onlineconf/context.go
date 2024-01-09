package onlineconf

import (
	"context"
	"fmt"
)

type ctxKey uint8

const ContextOnlineconfName ctxKey = iota

func ToContext(ctx context.Context, oi *OnlineconfInstance) context.Context {
	return context.WithValue(ctx, ContextOnlineconfName, oi)
}

func FromContext(ctx context.Context) *OnlineconfInstance {
	ctxVal := ctx.Value(ContextOnlineconfName)
	if ctxVal != nil {
		oi, ok := ctxVal.(*OnlineconfInstance)
		if !ok {
			return nil
		}

		return oi
	}

	return nil
}

// Initialize sets config directory for onlineconf modules.
// Default value is "/usr/local/etc/onlineconf"
func Initialize(ctx context.Context, options ...Option) (context.Context, error) {
	return ToContext(ctx, Create(options...)), nil
}

func Clone(from, to context.Context) (context.Context, error) {
	instance := FromContext(from)

	if instance.ro {
		return nil, fmt.Errorf("can't clone RO instance")
	}

	existsModules := instance.names

	newInstance := &OnlineconfInstance{
		ro:     true,
		logger: instance.logger,
		byName: make(map[string]*Module, len(existsModules)),
		names:  instance.names,
	}

	initMutex.Lock()
	defer initMutex.Unlock()

	for _, name := range existsModules {
		m := instance.GetModule(name)
		if m == nil {
			return nil, fmt.Errorf("module %s not found", name)
		}

		if err := instance.incRefcount(m.mmappedFile); err != nil {
			return nil, err
		}

		newInstance.byName[name] = &Module{
			ro:          true,
			name:        name,
			filename:    m.filename,
			cache:       make(map[string][]interface{}, startCacheSize),
			cdb:         m.cdb,
			mmappedFile: m.mmappedFile,
		}
	}

	return ToContext(to, newInstance), nil
}

func Release(main, cloned context.Context) error {
	mainInstance := FromContext(main)

	if mainInstance == nil {
		return fmt.Errorf("can't get main instance from context")
	}

	if mainInstance.ro {
		return fmt.Errorf("can't clone RO instance")
	}

	clonedInstance := FromContext(cloned)

	if clonedInstance == nil {
		return fmt.Errorf("can't get cloned instance from context")
	}

	initMutex.Lock()
	defer initMutex.Unlock()

	for _, name := range clonedInstance.names {
		m := clonedInstance.GetModule(name)
		mainInstance.decRefcount(m.mmappedFile)
	}

	clonedInstance.names = []string{}
	clonedInstance.byFile = map[string]*Module{}
	clonedInstance.byName = map[string]*Module{}

	return nil
}

func StartWatcher(ctx context.Context) error {
	instance := FromContext(ctx)
	if instance == nil {
		return fmt.Errorf("can't get instance from context")
	}

	return instance.StartWatcher(ctx)
}

func StopWatcher(ctx context.Context) error {
	instance := FromContext(ctx)
	if instance == nil {
		return fmt.Errorf("can't get instance from context")
	}

	return instance.StopWatcher()
}

func RegisterSubscription(ctx context.Context, module string, params []string, callback func() error) error {
	instance := FromContext(ctx)
	if instance == nil {
		return fmt.Errorf("can't get instance from context")
	}

	return instance.RegisterSubscription(module, params, callback)
}

func GetModule(ctx context.Context, name string) (*Module, error) {
	instance := FromContext(ctx)
	if instance == nil {
		return nil, fmt.Errorf("can't get instance from context")
	}

	return instance.GetModule(name), nil
}

func GetOrAddModule(ctx context.Context, name string) (*Module, error) {
	instance := FromContext(ctx)
	if instance == nil {
		return nil, fmt.Errorf("can't get instance from context")
	}

	return instance.GetOrAddModule(name)
}
