// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package task

import (
	"testing"
	"time"
)

func TestController(t *testing.T) {
	ctrl := NewController()
	h1, err := ctrl.AddTask("test")
	if err != nil {
		t.Fatal(err)
	}
	h1.Add(1)

	_, err = ctrl.AddTask("test")
	if err == nil {
		t.Fatal(err)
	} else {
		t.Log(err)
	}

	h1.Done()
	_, err = ctrl.AddTask("test")
	if err != nil {
		t.Fatal(err)
	}
}

func TestWait(t *testing.T) {
	ctrl := NewController()
	h1, err := ctrl.AddTask("test")
	if err != nil {
		t.Fatal(err)
	}
	h1.Add(1)

	now := time.Now()
	go func() {
		<-time.NewTimer(time.Second).C
		h1.Done()
	}()

	h1.Wait(500 * time.Millisecond)
	t.Log(time.Since(now).Milliseconds())
	_, err = ctrl.AddTask("test")
	if err == nil {
		t.Fatal(err)
	} else {
		t.Log(err)
	}

	h1.Wait(-1)
	t.Log(time.Since(now).Milliseconds())
	_, err = ctrl.AddTask("test")
	if err != nil {
		t.Fatal(err)
	}
}
