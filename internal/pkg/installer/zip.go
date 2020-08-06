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
	ProtocolVersion string
	AppVersion      string
	Name            string
	ExecCmd         string
	Info            string
}

type ZipPackage struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Info    string `json:"info"`

	pkgPath     string `json:"pkg_path"`
	installPath string `json:"install_path"`
}

type ZipInstaller struct {
	installDir string
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
	pkg.pkgPath = path
	pkg.installPath = saveDir

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

func (inst *ZipInstaller) Uninstall(pkg Package) error {
	return pkg.Uninstall()
}

func (inst *ZipInstaller) ListPackage() []Package {
	return nil
}

func (inst *ZipInstaller) GetPackage(name string) Package {
	return nil
}

func (pkg *ZipPackage) GetName() string {
	return pkg.Name
}

func (pkg *ZipPackage) GetVersion() string {
	return pkg.Version
}

func (pkg *ZipPackage) GetInfo() string {
	return pkg.Info
}

func (pkg *ZipPackage) Uninstall() error {
	return nil
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
