package onlineconf

import (
	"github.com/Nikolo/go-onlineconf/pkg/onlineconfInterface"
)

type options func(onlineconfInterface.Instance)

func (o options) Apply(oi onlineconfInterface.Instance) {
	o(oi)
}

func WithLogger(logger onlineconfInterface.Logger) onlineconfInterface.Option {
	return options(func(oii onlineconfInterface.Instance) {
		oi, ok := oii.(*OnlineconfInstance)
		if !ok {
			panic("onlineconf: invalid instance type")
		}

		oi.logger = logger
	})
}

func WithConfigDir(path string) onlineconfInterface.Option {
	return options(func(oii onlineconfInterface.Instance) {
		oi, ok := oii.(*OnlineconfInstance)
		if !ok {
			panic("onlineconf: invalid instance type")
		}

		oi.configDir = path
	})
}

func WithModules(moduleNames []string, required bool) onlineconfInterface.Option {
	return options(func(oi onlineconfInterface.Instance) {
		for _, moduleName := range moduleNames {
			m, err := oi.GetOrAddModule(moduleName)
			if (required && m == nil) || err != nil {
				panic("onlineconf: `" + moduleName + "` module not found or error + " + err.Error())
			}
		}
	})
}
