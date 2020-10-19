// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package magnet

import (
	"errors"
	"fmt"
	"github.com/xfali/magnet/pkg/installer"
	"github.com/xfali/magnet/pkg/task"
	"github.com/xfali/magnet/pkg/watcher"
	"github.com/xfali/xlog"
	"sync"
)

const (
	// 如果不存在则安装
	InstallFlagNotExists = 0
	// 使用新版本覆盖安装
	InstallFlagNewVersion = 1
	// 强制安装，不做任何检查
	InstallFlagForce = 1 << 1
	// 卸载已存在的所有安装包
	InstallFlagUninstallExists = 1 << 2
	// 卸载已存在的旧版本
	InstallFlagUninstallOld = 1 << 3
)

type Magnet struct {
	strategy   installer.Strategy
	installer  installer.Installer
	recorder   installer.Recorder
	listener   watcher.PackageListener
	watcherFac watcher.Factory

	taskCtrl task.Controller
	log      xlog.Logger
	watchers map[string]watcher.Watcher

	watchLock sync.Mutex
}

type Opt func(m *Magnet)

func New(opts ...Opt) *Magnet {
	ret := &Magnet{
		strategy:   installer.NewStrategy(),
		watcherFac: watcher.NewWatcher,
		watchers:   map[string]watcher.Watcher{},

		taskCtrl: task.NewController(),
		log:      xlog.GetLogger(),
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
func (m *Magnet) Install(path string, flag int) (installer.Package, error) {
	info, err := m.installer.ReadInfo(path)
	if err != nil {
		return nil, err
	}

	handle, err := m.taskCtrl.AddTask(info.GetName())
	if err != nil {
		return nil, err
	}
	// for install sub task
	handle.Add(1)
	defer handle.Done()

	if flag&InstallFlagForce == 0 {
		pkgs := m.recorder.GetPackage(info.GetName())
		if len(pkgs) > 0 {
			var pkg2remove []installer.Package
			// 非删除所有已存在安装包
			if flag&InstallFlagUninstallExists == 0 {
				//仅安装新版本
				if flag&InstallFlagNewVersion != 0 {
					for _, pkg := range pkgs {
						//安装包比现有安装更老
						if info.GetVersion() <= pkg.GetVersion() {
							return pkg, fmt.Errorf("Package: %s Exists version: %d is Newer than Install version %d ",
								pkg.GetName(), pkg.GetVersion(), info.GetVersion())
						} else {
							if flag&InstallFlagUninstallOld != 0 {
								pkg2remove = append(pkg2remove, pkg)
							}
						}
					}
				} else {
					// 默认非删除所有已存在安装包及非更新安装都选择不安装
					// 由于同名程序已存在，不做卸载处理可能出现问题。
					return nil, errors.New("Package: " + info.GetName() + " Exists")
				}
			} else {
				pkg2remove = pkgs
			}
			if len(pkg2remove) > 0 {
				m.uninstallPkgs(handle, false, pkg2remove...)
			}
		}
	}

	pkg, err := m.installer.Install(path, m.strategy)
	if err != nil {
		if pkg != nil {
			pkg.Uninstall(true)
			return nil, err
		}
	}
	err = m.recorder.Save(pkg)
	if err != nil {
		return nil, err
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
func (m *Magnet) Uninstall(name string, delPkg bool) error {
	pkgs := m.recorder.GetPackage(name)
	if len(pkgs) > 0 {
		return m.UninstallPkgs(delPkg, pkgs...)
	}
	return nil
}

func (m *Magnet) UninstallPkgs(delPkg bool, pkgs ...installer.Package) (err error) {
	if len(pkgs) > 0 {
		h, err := m.taskCtrl.AddTask(pkgs[0].GetName())
		if err != nil {
			return err
		}
		h.Add(len(pkgs))
		return m.uninstallPkgs(h, delPkg, pkgs...)
	}
	return nil
}

func (m *Magnet) uninstallOne(handle task.Handle, delPkg bool, pkg installer.Package) (err error) {
	defer handle.Done()

	m.log.Infof("Uninstall package: %s Exists version: %d delPkg: %d\n", pkg.GetName(), pkg.GetVersion(), delPkg)
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

func (m *Magnet) uninstallPkgs(handle task.Handle, delPkg bool, pkgs ...installer.Package) error {
	for _, pkg := range pkgs {
		err := m.uninstallOne(handle, delPkg, pkg)
		if err != nil {
			m.log.Infof("Uninstall package: %s error: %v\n", pkg.GetName(), err)
			continue
		}
	}
	return nil
}

// 根据安装包名称获得安装信息
func (m *Magnet) GetPackage(name string) []installer.Package {
	return m.recorder.GetPackage(name)
}

// 获得所有安装信息
func (m *Magnet) ListPackage() []installer.Package {
	return m.recorder.ListPackage()
}

// 设置安装策略，控制安装的行为
func SetInstallStrategy(s installer.Strategy) Opt {
	return func(m *Magnet) {
		m.strategy = s
	}
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

// 设置安装应用监听器，用于监听安装应用的状态，包括更新、删除
func SetLogger(l xlog.Logger) Opt {
	return func(m *Magnet) {
		m.log = l
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
