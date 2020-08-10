// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package test

import (
	"github.com/xfali/magnet/internal/pkg/installer"
	"github.com/xfali/stream"
	"testing"
)

func TestInstall(t *testing.T) {
	inst, err := installer.CreateInstaller("./target")
	if err != nil {
		t.Fatal(err)
	}
	recorder, err := installer.CreateRecorder("./target/pkg.rec")
	if err != nil {
		t.Fatal(err)
	}
	pkg, err := inst.Install("./assets/hello.pkg")
	if err != nil {
		t.Fatal(err)
	}
	recorder.Save(pkg)
}

func TestLoadPackage(t *testing.T) {
	recorder, err := installer.CreateRecorder("./target/pkg.rec")
	if err != nil {
		t.Fatal(err)
	}
	pkgs := recorder.ListPackage()
	stream.Slice(pkgs).Foreach(func(pkg installer.Package) {
		t.Log(pkg)
	})
}

func TestUninstall(t *testing.T) {
	recorder, err := installer.CreateRecorder("./target/pkg.rec")
	if err != nil {
		t.Fatal(err)
	}
	pkg := recorder.GetPackage("test")
	if pkg != nil {
		err := pkg.Uninstall(false)
		if err != nil {
			t.Fatal(err)
		}
		recorder.Remove(pkg)
	}
}
