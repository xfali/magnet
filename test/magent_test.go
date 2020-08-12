// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package test

import (
	"github.com/xfali/magnet"
	"testing"
)

func TestMagnet(t *testing.T) {
	m := magnet.New(magnet.Default("./target", "./target/pkg.rec"))
	t.Run("install", func(t *testing.T) {
		err := m.Install("./assets/hello.pkg")
		if err != nil {
			t.Fatal(err)
		}
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
