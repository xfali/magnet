// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package watcher

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/xfali/magnet/internal/pkg/installer"
	"github.com/xfali/stream"
	"log"
	"sync"
)

type PackageListener interface {
	//OnCreate(p installer.Package, filename string)
	OnUpdate(p installer.Package, filename string)
	OnRemove(p installer.Package, filename string)
}

type Watcher interface {
	AddListener(PackageListener)
	Watch(p installer.Package)
}

type PackageWatcher struct {
	listeners []PackageListener
	stop      chan struct{}

	lock sync.Mutex
}

func (w *PackageWatcher) Stop() {
	close(w.stop)
}

func (w *PackageWatcher) AddListener(l PackageListener) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.listeners = append(w.listeners, l)
}

func (w *PackageWatcher) getListeners() []PackageListener {
	w.lock.Lock()
	defer w.lock.Unlock()
	return w.listeners
}

func (w *PackageWatcher) Watch(p installer.Package) {
	//创建一个监控对象
	watch, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	err = watch.Add(p.GetInstallPath())
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		defer watch.Close()

		for {
			select {
			case ev := <-watch.Events:
				//判断事件发生的类型，如下5种
				// Create 创建
				// Write 写入
				// Remove 删除
				// Rename 重命名
				// Chmod 修改权限
				if ev.Op&fsnotify.Create == fsnotify.Create {
					//stream.Slice(w.getListeners()).Foreach(func(l PackageListener) {
					//	l.OnCreate(p, ev.Name)
					//})
				}
				if ev.Op&fsnotify.Write == fsnotify.Write {
					stream.Slice(w.getListeners()).Foreach(func(l PackageListener) {
						l.OnUpdate(p, ev.Name)
					})
				}
				if ev.Op&fsnotify.Remove == fsnotify.Remove {
					stream.Slice(w.getListeners()).Foreach(func(l PackageListener) {
						l.OnRemove(p, ev.Name)
					})
					if ev.Name == p.GetInstallPath() {
						log.Println("package removed, exit watch")
						return
					}
				}
			case err := <-watch.Errors:
				log.Println("error : ", err)
				return
			case <-w.stop:
				return
			}
		}
	}()
}

type DummyListener struct {}

func (l *DummyListener) OnCreate(p installer.Package, filename string) {
	fmt.Printf("package : %v created, file : %s\n", p, filename)
}

func (l *DummyListener)OnUpdate(p installer.Package, filename string) {
	fmt.Printf("package : %v update, file : %s\n", p, filename)
}

func (l *DummyListener)OnRemove(p installer.Package, filename string) {
	fmt.Printf("package : %v remove, file : %s\n", p, filename)
}
