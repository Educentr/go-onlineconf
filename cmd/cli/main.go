package main

import (
	"flag"

	"github.com/Nikolo/go-onlineconf/pkg/onlineconf_dev"
)

func main() {
	flag.String("yaml", "./configs/onlineconf.yml", "path to the yaml file")
	flag.String("dir", "/usr/local/etc/onlineconf", "path to the directory where the cdb files will be stored")
	flag.String("module", "TREE", "module name")
	flag.Parse()
	onlineconf_dev.GenerateCDBFromYaml(flag.Lookup("dir").Value.String(), flag.Lookup("module").Value.String(), flag.Lookup("yaml").Value.String())
}
