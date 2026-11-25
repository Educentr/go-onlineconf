package onlineconf

import (
	"context"
	"fmt"

	"github.com/Educentr/go-onlineconf/pkg/onlineconfInterface"
)

type options func(onlineconfInterface.Instance)

func (o options) Apply(oi onlineconfInterface.Instance) {
	o(oi)
}

func WithLogger(logger onlineconfInterface.Logger) options {
	return options(func(oii onlineconfInterface.Instance) {
		oi, ok := oii.(*OnlineconfInstance)
		if !ok {
			panic("onlineconf: invalid instance type")
		}

		oi.logger = logger
	})
}

func WithConfigDir(path string) options {
	return options(func(oii onlineconfInterface.Instance) {
		oi, ok := oii.(*OnlineconfInstance)
		if !ok {
			panic("onlineconf: invalid instance type")
		}

		oi.configDir = path
	})
}

// WithModules opens the named modules right away, instead of leaving them to the first read
// that needs them.
//
// Opening a module fails when its file is missing from the config directory or when the file
// cannot serve as a CDB database (ErrInvalidCDB). What happens then is what required selects:
//
//   - required == true: the option panics, so Create and Initialize never hand out an
//     instance that misses a module the caller declared mandatory.
//   - required == false: the error is reported to the logger of the instance and the module
//     is skipped. The names that follow it in the list are still opened, and a later
//     GetOrAddModule for the skipped module tries to open it again.
//
// Options are applied in the order they are given, and this one reads configDir and logger,
// so keep it after WithConfigDir and WithLogger.
func WithModules(moduleNames []string, required bool) options {
	return options(func(oii onlineconfInterface.Instance) {
		oi, ok := oii.(*OnlineconfInstance)
		if !ok {
			panic("onlineconf: invalid instance type")
		}

		for _, moduleName := range moduleNames {
			// GetOrAddModule returns a module whenever it returns no error, so the error is
			// the only thing worth looking at here.
			if _, err := oi.GetOrAddModule(moduleName); err != nil {
				if required {
					panic(fmt.Sprintf("onlineconf: can't open required module `%s`: %s", moduleName, err))
				}

				oi.logger.Error(context.Background(),
					"onlineconf: can't open optional module `"+moduleName+"`, skipping it", err)
			}
		}
	})
}
