package onlineconf

import (
	"context"
	"fmt"

	"github.com/Educentr/go-onlineconf/pkg/onlineconfInterface"
)

type ctxKey uint8

const ContextOnlineconfName ctxKey = iota

func ToContext(ctx context.Context, oi onlineconfInterface.Instance) context.Context {
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
func Initialize(ctx context.Context, options ...onlineconfInterface.Option) (context.Context, error) {
	return ToContext(ctx, Create(options...)), nil
}

func Clone(from, to context.Context) (context.Context, error) {
	instance := FromContext(from)

	newInstance, err := instance.Clone()
	if err != nil {
		return nil, fmt.Errorf("can't clone instance: %w", err)
	}

	return ToContext(to, newInstance), nil
}

func Release(main, cloned context.Context) error {
	mainInstance := FromContext(main)

	if mainInstance == nil {
		return fmt.Errorf("can't get main instance from context")
	}

	clonedInstance := FromContext(cloned)

	if clonedInstance == nil {
		return fmt.Errorf("can't get cloned instance from context")
	}

	return mainInstance.Release(main, clonedInstance)
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

func GetModule(ctx context.Context, name string) (onlineconfInterface.Module, error) { //nolint:ireturn
	instance := FromContext(ctx)
	if instance == nil {
		return nil, fmt.Errorf("can't get instance from context")
	}

	return instance.GetModule(name), nil
}

func GetOrAddModule(ctx context.Context, name string) (onlineconfInterface.Module, error) { //nolint:ireturn
	instance := FromContext(ctx)
	if instance == nil {
		return nil, fmt.Errorf("can't get instance from context")
	}

	return instance.GetOrAddModule(name)
}
