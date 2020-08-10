// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package test

import (
	"github.com/xfali/magnet/internal/pkg/installer"
	"github.com/xfali/magnet/internal/pkg/watcher"
	"testing"
)

func TestWatcher(t *testing.T) {
	w := watcher.PackageWatcher{}
	w.AddListener(&watcher.DummyListener{})
	pkg := installer.ZipPackage{
		InstallPath: "target/test",
	}
	w.Watch(&pkg)
	//<-time.NewTimer(30 * time.Second).C
	select {}
}
