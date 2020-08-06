// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package installer

type Package interface {
	GetName() string
	GetVersion() string
	GetInfo() string

	Uninstall() error
}

type Installer interface {
	Install(path string) (Package, error)
	Uninstall(pkg Package) error
}

type Recorder interface {
	Save(pkg Package) error
	Remove(pkg Package) error

	ListPackage() []Package
	GetPackage(name string) Package
}
