// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package magnet

import (
	"errors"
	"fmt"
	"github.com/xfali/goutils/log"
	"github.com/xfali/magnet/pkg/installer"
	"github.com/xfali/magnet/pkg/watcher"
	"sync"
)

const (
	//如果不存在则安装
	InstallFlagNotExists = 0
	//使用新版本覆盖安装
	InstallFlagNewVersion = 1
	//强制覆盖安装
	InstallFlagForce = 1 << 1
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

// 关闭Magnet
func (m *Magnet) Close() error {
	m.watchLock.Lock()
	for _, v := range m.watchers {
		v.Stop()
	}
	m.watchLock.Unlock()
	return nil
}

// 获得安装包信息
// param： path安装包路径
func (m *Magnet) ReadInfo(path string) (pkg installer.PackageInfo, err error) {
	return m.installer.ReadInfo(path)
}

// 安装
// param： path安装包路径， flag 安装标志
func (m *Magnet) Install(path string, flag int) (pkg installer.Package, err error) {
	info, err := m.installer.ReadInfo(path)
	if err != nil {
		return
	}

	pkg = m.recorder.GetPackage(info.GetName())
	if pkg != nil {
		//非强制安装
		if flag&InstallFlagForce == 0 {
			//安装新版本或更新已有版本
			if flag&InstallFlagNewVersion != 0 {
				//安装包比现有安装更老
				if info.GetVersion() < pkg.GetVersion() {
					return pkg, fmt.Errorf("Package: %s Exists version: %d is Newer than Install version %d ",
						pkg.GetName(), pkg.GetVersion(), info.GetVersion())
				}
			} else {
				//默认非强制安装及非更新安装都选择不安装
				return pkg, errors.New("Package: " + pkg.GetName() + " Exists")
			}
		}
		log.Info("Package: %s Exists version: %d , New Package version %d , install flag: %d\n",
			pkg.GetName(), pkg.GetVersion(), info.GetVersion(), flag)
		m.Uninstall(pkg.GetName(), false)
	}

	pkg, err = m.installer.Install(path)
	if err != nil {
		if pkg != nil {
			pkg.Uninstall(true)
			return nil, err
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

	return pkg, nil
}

// 卸载安装
// param: name 安装包名称， delPkg 是否卸载同时删除安装包
func (m *Magnet) Uninstall(name string, delPkg bool) (err error) {
	pkg := m.recorder.GetPackage(name)
	if pkg == nil {
		return errors.New("package: " + name + " not found")
	}
	err = pkg.Uninstall(delPkg)
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

// 根据安装包名称获得安装信息
func (m *Magnet) GetPackage(name string) installer.Package {
	return m.recorder.GetPackage(name)
}

// 获得所有安装信息
func (m *Magnet) ListPackage() []installer.Package {
	return m.recorder.ListPackage()
}

// 设置安装管理器，实现安装的流程
func SetInstaller(i installer.Installer) Opt {
	return func(m *Magnet) {
		m.installer = i
	}
}

// 设置安装记录工具，用于记录安装信息
func SetRecorder(r installer.Recorder) Opt {
	return func(m *Magnet) {
		m.recorder = r
	}
}

// 设置安装应用监听器，用于监听安装应用的状态，包括更新、删除
func SetListener(l watcher.PackageListener) Opt {
	return func(m *Magnet) {
		m.listener = l
	}
}

// 设置应用监听器的生成器
func SetWatchFactory(fac watcher.Factory) Opt {
	return func(m *Magnet) {
		m.watcherFac = fac
	}
}

// 使用默认配置，包括Installer、Recorder、WatcherFactory、Listener
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
		m.watcherFac = watcher.NewWatcher
		m.listener = &watcher.DummyListener{}
	}
}
