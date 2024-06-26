package onlineconf_dev

import (
	"github.com/Educentr/go-onlineconf/pkg/onlineconfInterface"
)

type optionsDev func(onlineconfInterface.Instance)

func (o optionsDev) Apply(oi onlineconfInterface.Instance) {
	o(oi)
}

func WithGenerateFromYAML(moduleName, path string) func(onlineconfInterface.Instance) {
	return optionsDev(func(oi onlineconfInterface.Instance) {
		GenerateCDBFromYaml(oi.GetConfigDir(), moduleName, path)
	})
}
