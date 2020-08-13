// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package magnet

import (
	"errors"
	"github.com/xfali/magnet/pkg/installer"
	"github.com/xfali/magnet/pkg/watcher"
	"sync"
)

type Magnet struct {
	installer  installer.Installer
	recorder   installer.Recorder
	listener   watcher.PackageListener
	watcherFac watcher.Factory
	watchers   map[string]watcher.Watcher

	watchLock sync.Mutex
}

type Opt func(m *Magnet)

func New(opts ...Opt) *Magnet {
	ret := &Magnet{
		watcherFac: watcher.NewWatcher,
		watchers:   map[string]watcher.Watcher{},
	}
	for i := range opts {
		opts[i](ret)
	}
	return ret
}

func (m *Magnet) Close() error {
	m.watchLock.Lock()
	for _, v := range m.watchers {
		v.Stop()
	}
	m.watchLock.Unlock()
	return nil
}

func (m *Magnet) Install(path string) (err error) {
	pkg, err := m.installer.Install(path)
	if err != nil {
		if pkg != nil {
			pkg.Uninstall(true)
			return err
		}
	}
	err = m.recorder.Save(pkg)
	if err != nil {
		return
	}

	w := m.watcherFac()
	w.AddListener(m.listener)
	w.Watch(pkg)

	m.watchLock.Lock()
	defer m.watchLock.Unlock()
	m.watchers[pkg.GetInstallPath()] = w

	return nil
}

func (m *Magnet) Uninstall(name string) (err error) {
	pkg := m.recorder.GetPackage(name)
	if pkg == nil {
		return errors.New("package: " + name + " not found")
	}
	err = pkg.Uninstall(false)
	if err != nil {
		return err
	}

	func() {
		m.watchLock.Lock()
		defer m.watchLock.Unlock()
		w := m.watchers[pkg.GetInstallPath()]
		if w != nil {
			w.Stop()
		}
		delete(m.watchers, pkg.GetInstallPath())
	}()

	return m.recorder.Remove(pkg)
}

func (m *Magnet) ListPackage() []installer.Package {
	return m.recorder.ListPackage()
}

func SetInstaller(i installer.Installer) Opt {
	return func(m *Magnet) {
		m.installer = i
	}
}

func SetRecorder(r installer.Recorder) Opt {
	return func(m *Magnet) {
		m.recorder = r
	}
}

func SetListener(l watcher.PackageListener) Opt {
	return func(m *Magnet) {
		m.listener = l
	}
}

func SetWatchFactory(fac watcher.Factory) Opt {
	return func(m *Magnet) {
		m.watcherFac = fac
	}
}

func Default(installDir, recordFile string) Opt {
	return func(m *Magnet) {
		var err error
		m.installer, err = installer.CreateInstaller(installDir)
		if err != nil {
			panic(err)
		}
		m.recorder, err = installer.CreateRecorder(recordFile)
		if err != nil {
			panic(err)
		}
		m.listener = &watcher.DummyListener{}
	}
}
