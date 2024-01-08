package testCDB

import (
	"log"
	"os"

	"github.com/colinmarc/cdb"
)

func Generate(filename string, m map[string][]byte) {
	writer, err := cdb.Create(filename + ".tmp")
	if err != nil {
		log.Fatal(err)
	}

	for k, v := range m {
		writer.Put([]byte(k), v)
	}

	writer.Close()

	os.Rename(filename+".tmp", filename)
}
