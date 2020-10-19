// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package test

import (
	"github.com/xfali/magnet/pkg/installer"
	"testing"
)

func TestJsonRecord(t *testing.T) {
	r, err := installer.CreateJsonRecorder("./target/pkg.json")
	if err != nil {
		t.Fatal(err)
	}
	pkg := &installer.ZipPackage{
		Name:        "test",
		Version:     0,
		Info:        "xx",
		PkgPath:     "xx",
		InstallPath: "xxx",
	}
	r.Save(pkg)

	v := r.ListPackage()
	if len(v) == 0 {
		t.Fatal("pkg is empty")
	}
	for _, v := range v {
		t.Log(v)
	}

	v = r.GetPackage("test")
	if len(v) == 0 {
		t.Fatal("pkg is empty")
	}
	for _, v := range v {
		t.Log(v)
	}

	r.Remove(pkg)
	v = r.GetPackage("test")
	if len(v) != 0 {
		t.Fatal("pkg is not empty")
	}
}

func TestJsonRecord2(t *testing.T) {
	r, err := installer.CreateJsonRecorder("./target/pkg.json")
	if err != nil {
		t.Fatal(err)
	}
	pkg := &installer.ZipPackage{
		Name:        "test",
		Version:     0,
		Info:        "xx",
		PkgPath:     "xx",
		InstallPath: "xxx",
	}
	r.Save(pkg)

	r.Save(pkg)

	v := r.ListPackage()
	if len(v) != 1 {
		t.Fatal("pkg number is not 1")
	}
	for _, v := range v {
		t.Log(v)
	}

	v = r.GetPackage("test")
	if len(v) != 1 {
		t.Fatal("pkg number is not 1")
	}
	for _, v := range v {
		t.Log(v)
	}

	pkg.Version = 1
	r.Save(pkg)

	v = r.GetPackage("test")
	if len(v) != 2 {
		t.Fatal("pkg number is not 2")
	}
	for _, v := range v {
		t.Log(v)
	}
}
