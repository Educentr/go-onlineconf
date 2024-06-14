package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/Educentr/go-onlineconf/pkg/onlineconf"
	"github.com/Educentr/go-onlineconf/pkg/onlineconf_dev"
)

func main() {
	flag.String("dir", "/usr/local/etc/onlineconf", "path to the directory where the cdb files will be stored")
	flag.String("module", "TREE", "module name")
	flag.String("command", "generate", "command to execute (generate, help)")
	flag.String("yaml", "./configs/onlineconf.yml", "path to the yaml file for command generate")
	flag.String("config-name", "", "name of the config parameter in the config file. Use with command get")

	flag.Bool("help", false, "show help") // This is a flag that will be used to show the help message if the user doesn't provide any arguments or provides the -help flag

	flag.Parse()

	if flag.Lookup("help").Value.String() == "true" || flag.Lookup("command").Value.String() == "" {
		flag.Usage()
		return
	}

	switch flag.Lookup("command").Value.String() {
	case "generate":
		if flag.Lookup("config-name").Value.String() != "" {
			flag.Usage()
			return
		}

		onlineconf_dev.GenerateCDBFromYaml(flag.Lookup("dir").Value.String(), flag.Lookup("module").Value.String(), flag.Lookup("yaml").Value.String())
	case "get":
		if flag.Lookup("config-name").Value.String() == "" {
			flag.Usage()
			return
		}

		inst := onlineconf.Create(
			onlineconf.WithConfigDir(flag.Lookup("dir").Value.String()),
		)

		module, err := inst.GetOrAddModule(flag.Lookup("module").Value.String())
		if err != nil {
			log.Fatalf("Can't get module %s", flag.Lookup("module").Value.String())
		}

		if module == nil {
			log.Fatalf("Can't get module %s", flag.Lookup("module").Value.String())
		}

		val, ex, err := module.GetStringIfExists(flag.Lookup("config-name").Value.String())
		if err != nil {
			log.Fatalf("Can't get config value %s", err)
		}

		if !ex {
			fmt.Printf("Config parameter %s does not exist\n", flag.Lookup("config-name").Value.String())
		} else {
			fmt.Printf("Value of config parameter %s is %s\n", flag.Lookup("config-name").Value.String(), val)
		}
	}
}
