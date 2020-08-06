// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package test

import (
	"github.com/xfali/magnet"
	"testing"
)

func TestWatcher(t *testing.T) {
	w := magnet.Watcher{}
	w.Watch("C:\\tmp\\dest\\backup\\1")
	//<-time.NewTimer(30 * time.Second).C
	select {}
}
