// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package installer

import (
	"archive/zip"
	"encoding/json"
	io2 "github.com/xfali/goutils/io"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	ZIP_INFO_FILENAME = "pkg.info"
)

type ZipPackageInfo struct {
	ProtocolVersion int    `json:"protocolVersion" yaml:"protocolVersion"`
	AppVersion      int    `json:"appVersion" yaml:"appVersion"`
	Name            string `json:"name" yaml:"name"`
	ExecCmd         string `json:"execCmd" yaml:"execCmd"`
	Info            string `json:"info" yaml:"info"`
	ExecName        string `json:"execName" yaml:"execName"`
}

type ZipPackage struct {
	Name    string `json:"name" yaml:"name"`
	Version int    `json:"version" yaml:"version"`
	Info    string `json:"info" yaml:"info"`

	PkgPath     string `json:"pkgPath" yaml:"pkgPath"`
	InstallPath string `json:"installPath" yaml:"installPath"`
}

type ZipInstaller struct {
	installDir string
}

func CreateInstaller(installDir string) (*ZipInstaller, error){
	ret := &ZipInstaller{
		installDir: installDir,
	}
	if !io2.IsPathExists(installDir) {
		err := io2.Mkdir(installDir)
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}

func (inst *ZipInstaller) Install(path string) (Package, error) {
	pkg := &ZipPackage{}
	pluginName := filepath.Base(path)
	lastIndex := strings.LastIndex(pluginName, ".")
	if lastIndex != -1 {
		pluginName = pluginName[:lastIndex]
	}
	saveDir := filepath.Join(inst.installDir, pluginName)
	if !io2.IsPathExists(saveDir) {
		err := io2.Mkdir(saveDir)
		if err != nil {
			return nil, err
		}
	}
	pkg.PkgPath = path
	pkg.InstallPath = saveDir

	reader, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	for _, file := range reader.File {
		err := func() error {
			rc, err := file.Open()
			if err != nil {
				return err
			}
			defer rc.Close()
			filename := filepath.Join(saveDir, file.Name)
			dir := filepath.Dir(filename)
			if dir != "" && dir != "." {
				err = os.MkdirAll(dir, 0755)
				if err != nil {
					return err
				}
			}

			w, err := os.Create(filename)
			if err != nil {
				return err
			}
			defer w.Close()
			_, err = io.Copy(w, rc)
			if err != nil {
				return err
			}
			w.Close()
			rc.Close()

			if filepath.Base(filename) == ZIP_INFO_FILENAME {
				info, err := readZipInfo(filename)
				if err != nil {
					return err
				}
				pkg.Name = info.Name
				pkg.Version = info.AppVersion
				pkg.Info = info.Info
			}

			return nil
		}()
		if err != nil {
			return pkg, err
		}
	}
	return pkg, nil
}

func (inst *ZipInstaller) Uninstall(pkg Package, del bool) error {
	return pkg.Uninstall(del)
}

type ZipRecorder struct {
	path string
	pkgs map[string]Package
}

func CreateRecorder(path string) (*ZipRecorder, error) {
	ret := &ZipRecorder{
		path: path,
		pkgs: map[string]Package{},
	}
	if io2.IsPathExists(path) {
		d, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}
		tmp := map[string]*ZipPackage{}
		err = json.Unmarshal(d, &tmp)
		if err != nil {
			return nil, err
		}
		for k, v := range tmp {
			ret.pkgs[k] = v
		}
	}

	return ret, nil
}

func (r *ZipRecorder) flush() error {
	d, err := json.Marshal(r.pkgs)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(r.path, d, os.ModePerm)
}

func (r *ZipRecorder) Save(pkg Package) error {
	r.pkgs[pkg.GetName()] = pkg
	return r.flush()
}

func (r *ZipRecorder) Remove(pkg Package) error {
	delete(r.pkgs, pkg.GetName())
	return r.flush()
}

func (r *ZipRecorder) ListPackage() []Package {
	ret := make([]Package, len(r.pkgs))
	for _, v := range r.pkgs {
		ret = append(ret, v)
	}
	return ret
}

func (r *ZipRecorder) GetPackage(name string) Package {
	return r.pkgs[name]
}

func (pkg *ZipPackage) GetName() string {
	return pkg.Name
}

func (pkg *ZipPackage) GetVersion() int {
	return pkg.Version
}

func (pkg *ZipPackage) GetInfo() string {
	return pkg.Info
}

func (pkg *ZipPackage) Uninstall(delPkg bool) (err error) {
	if io2.IsPathExists(pkg.InstallPath) {
		err = os.RemoveAll(pkg.InstallPath)
	}

	if delPkg && io2.IsPathExists(pkg.PkgPath) {
		err = os.RemoveAll(pkg.PkgPath)
	}
	return err
}

func readZipInfo(path string) (*ZipPackageInfo, error) {
	d, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	ret := &ZipPackageInfo{}
	err = json.Unmarshal(d, ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}
