package onlineconf_dev

import (
	"github.com/Nikolo/go-onlineconf/pkg/onlineconfInterface"
)

type optionsDev func(onlineconfInterface.Instance)

func (o optionsDev) Apply(oi onlineconfInterface.Instance) {
	o(oi)
}

func WithGenerateFromYAML(moduleName, path string) onlineconfInterface.Option {
	return optionsDev(func(oi onlineconfInterface.Instance) {
		GenerateCDBFromYaml(oi.GetConfigDir(), moduleName, path)
	})
}
