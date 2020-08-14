// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package test

import (
	"github.com/xfali/magnet/pkg/installer"
	"github.com/xfali/magnet/pkg/watcher"
	"testing"
	"time"
)

func TestWatcher(t *testing.T) {
	w := watcher.NewWatcher()
	w.AddListener(&watcher.DummyListener{})
	pkg := installer.ZipPackage{
		InstallPath: "target/test",
	}
	w.Watch(&pkg)
	//<-time.NewTimer(30 * time.Second).C
	select {}
}

func TestPackageBatchWatcher(t *testing.T) {
	w := watcher.NewPackageBatchWatcher(1 * time.Second)
	w.AddListener(&watcher.DummyListener{})
	pkg := installer.ZipPackage{
		InstallPath: "target/test",
	}
	w.Watch(&pkg)
	//<-time.NewTimer(30 * time.Second).C
	select {}
}
