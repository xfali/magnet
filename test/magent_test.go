// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package test

import (
	"github.com/xfali/magnet"
	"github.com/xfali/magnet/pkg/watcher"
	"testing"
	"time"
)

func TestMagnet(t *testing.T) {
	m := magnet.New(magnet.Default("./target", "./target/pkg.rec"), magnet.SetWatchFactory(func() watcher.Watcher {
		return watcher.NewPackageBatchWatcher(1 * time.Second)
	}))
	t.Run("install", func(t *testing.T) {
		err := m.Install("./assets/hello.pkg")
		if err != nil {
			t.Fatal(err)
		}
		//time.Sleep(time.Minute)
	})

	t.Run("list", func(t *testing.T) {
		pkgs := m.ListPackage()
		for _, v := range pkgs {
			t.Log(v)
		}
	})

	t.Run("uninstall", func(t *testing.T) {
		err := m.Uninstall("test")
		if err != nil {
			t.Fatal(err)
		}
	})
}
