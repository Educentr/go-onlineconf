package onlineconf_dev

import (
	"log"
	"os"
	"path"
	"strings"

	"github.com/colinmarc/cdb"
	"gopkg.in/yaml.v3"
)

func GenerateCDBFromYaml(dir, modulename string, yamlfilename string) {
	yamlfile, err := os.ReadFile(yamlfilename)
	if err != nil {
		log.Fatal(err)
	}

	m := make(map[string]interface{})

	err = yaml.Unmarshal(yamlfile, &m)
	if err != nil {
		log.Fatal(err)
	}

	GenerateCDB(dir, modulename, m)
}

func GenerateCDB(dir, modulename string, m map[string]interface{}) {
	if strings.Contains(modulename, ".cdb") {
		panic("modulename should not contain .cdb")
	}

	modulename += ".cdb"

	writer, err := cdb.Create(path.Join(dir, modulename+".tmp"))
	if err != nil {
		log.Fatal(err)
	}

	for k, v := range m {
		writer.Put([]byte(k), Encode(v))
	}

	writer.Close()

	os.Rename(path.Join(dir, modulename+".tmp"), path.Join(dir, modulename))
}
