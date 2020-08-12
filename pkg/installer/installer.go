// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package installer

type Package interface {
	// 获得安装包名称
	GetName() string
	// 获得安装包版本号
	GetVersion() int
	// 获得安装包信息
	GetInfo() string
	// 获得安装路径
	GetInstallPath() string

	// 卸载已安装应用文件，delPkg为true则将同时删除安装包
	Uninstall(delPkg bool) error
}

type Installer interface {
	// 安装
	// Param: path安装包路径，请使用绝对路径
	// Return: Package安装包信息， error 安装错误
	Install(path string) (Package, error)

	// 卸载
	// Param: pkg安装包信息，delPkg是否同时删除安装包
	// Return: error 卸载错误
	Uninstall(pkg Package, delPkg bool) error
}

type Recorder interface {
	// 保存安装包信息
	Save(pkg Package) error
	// 删除安装包信息
	Remove(pkg Package) error

	// 列出所有已安装应用
	ListPackage() []Package
	// 根据名称获得安装信息
	GetPackage(name string) Package
}
