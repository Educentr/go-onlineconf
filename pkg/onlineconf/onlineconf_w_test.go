package onlineconf

import (
	"context"
	"os"
	"path"
	"testing"
	"time"

	testCDB "gitlab.educentr.info/godev/onlinecof-test/test"
)

const tmpConfDir = "/tmp/onlineconf/"

var _ = os.Mkdir(tmpConfDir, os.ModePerm)

var globalCtx, _ = Initialize(context.Background(), WithConfigDir(tmpConfDir))

func TestGetDefaultModuleW(t *testing.T) {
	testCDB.Generate(path.Join(tmpConfDir, "TREE.cdb"), map[string][]byte{"bla": []byte("sblav")})
	err := StartWatcher(globalCtx)
	if err != nil {
		t.Errorf("can't start watcher: %s", err)
	}

	v, err := GetString(globalCtx, "bla")
	if err != nil {
		t.Error("error get string", err)
	}

	if v != "blav" {
		t.Error("invalid value", v)
	}

	newCtx := context.Background()
	newCtx, _ = Clone(globalCtx, newCtx)

	testCDB.Generate(path.Join(tmpConfDir, "TREE.cdb"), map[string][]byte{"bla": []byte("sblav1")})
	time.Sleep(time.Millisecond * 100)

	v, err = GetString(globalCtx, "bla")
	if err != nil {
		t.Error("error get string after update", err)
	}

	if v != "blav1" {
		t.Error("invalid value after update", v)
	}

	v, err = GetString(newCtx, "bla")
	if err != nil {
		t.Error("error get string", err)
	}

	if v != "blav" {
		t.Error("invalid value", v)
	}

	err = Release(globalCtx, newCtx)
	if err != nil {
		t.Errorf("error release: %s", err)
	}

	instance := FromContext(newCtx)
	m := instance.Get("TREE")
	if m != nil {
		t.Errorf("module exists after release")
	}

	instance = FromContext(globalCtx)
	if len(instance.mmappedFiles) != 1 {
		t.Errorf("invalid mmaped size: %d", len(instance.mmappedFiles))
	}

	if err = StopWatcher(globalCtx); err != nil {
		t.Errorf("can't sdtop watcher: %s", err)
	}

	if instance.watcher.watcher != nil {
		t.Errorf("watcher exists after stop")
	}
}
