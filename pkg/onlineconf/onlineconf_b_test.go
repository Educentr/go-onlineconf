package onlineconf_test

import (
	"context"
	"os"
	"path"
	"testing"
	"time"

	"gitlab.educentr.info/godev/onlinecof-test/pkg/onlineconf"
	testCDB "gitlab.educentr.info/godev/onlinecof-test/test"
)

const tmpConfDir = "/tmp/onlineconf/"

var _ = os.Mkdir(tmpConfDir, os.ModePerm)

var globalCtx, _ = onlineconf.Initialize(context.Background(), onlineconf.WithConfigDir(tmpConfDir))

func BenchmarkClone(b *testing.B) {
	for i := 0; i < b.N; i++ {
		onlineconf.Clone(globalCtx, context.Background())
	}
}

func BenchmarkGetStringIfExists(b *testing.B) {
	for i := 0; i < b.N; i++ {
		onlineconf.GetStringIfExists(globalCtx, "bla")
	}
}

func TestGetDefaultModuleB(t *testing.T) {
	testCDB.Generate(path.Join(tmpConfDir, "TREE.cdb"), map[string][]byte{"bla": []byte("sblav")})

	err := onlineconf.StartWatcher(globalCtx)
	if err != nil {
		t.Errorf("can't start watcher: %s", err)
	}

	v, err := onlineconf.GetString(globalCtx, "bla")
	if err != nil {
		t.Error("error get string", err)
	}

	if v != "blav" {
		t.Error("invalid value", v)
	}

	callbackCalled := false

	onlineconf.RegisterCallback(globalCtx, "TREE", []string{"bla"}, func() error { callbackCalled = true; return nil })

	newCtx := context.Background()
	newCtx, _ = onlineconf.Clone(globalCtx, newCtx)

	testCDB.Generate(path.Join(tmpConfDir, "TREE.cdb"), map[string][]byte{"bla": []byte("sblav1")})
	time.Sleep(time.Millisecond * 100)

	if !callbackCalled {
		t.Error("callback not called")
	}

	v, err = onlineconf.GetString(globalCtx, "bla")
	if err != nil {
		t.Error("error get string after update", err)
	}

	if v != "blav1" {
		t.Error("invalid value after update", v)
	}

	v, err = onlineconf.GetString(newCtx, "bla")
	if err != nil {
		t.Error("error get string", err)
	}

	if v != "blav" {
		t.Error("invalid value", v)
	}

	onlineconf.Release(globalCtx, newCtx)

	m := onlineconf.FromContext(newCtx).Get("TREE")
	if m != nil {
		t.Errorf("module exists after release")
	}

	if err = onlineconf.StopWatcher(globalCtx); err != nil {
		t.Errorf("can't sdtop watcher: %s", err)
	}
}