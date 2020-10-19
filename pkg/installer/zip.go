// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package installer

import (
	"archive/zip"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	io2 "github.com/xfali/goutils/io"
	"io"
	"os"
	"path/filepath"
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
	Description     string `json:"description" yaml:"description"`
	ExecName        string `json:"execName" yaml:"execName"`
	Checksum        string `json:"checksum" yaml:"checksum"`
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

func CreateInstaller(installDir string) (*ZipInstaller, error) {
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

func (inst *ZipInstaller) ReadInfo(path string) (PackageInfo, error) {
	return getPackageInfo(path)
}

func (inst *ZipInstaller) Install(path string, strategy Strategy) (Package, error) {
	pkg := &ZipPackage{}
	info, err := getPackageInfo(path)
	if err != nil {
		return nil, err
	}
	pkg.Name = info.Name
	pkg.Version = info.AppVersion
	pkg.Info = info.Info

	saveDir, err := strategy.GenInstallPath(inst.installDir, info)
	if err != nil {
		return nil, err
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

			w, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, file.Mode())
			if err != nil {
				return err
			}
			defer w.Close()
			if filepath.Base(filename) == info.ExecName && info.Checksum != "" {
				h := md5.New()
				multiWriter := io.MultiWriter(w, h)
				_, err = io.Copy(multiWriter, rc)
				if err != nil {
					return err
				}
				if hex.EncodeToString(h.Sum(nil)) != info.Checksum {
					return errors.New("Checksum not match ")
				}
				return nil
			}
			_, err = io.Copy(w, rc)
			return err
		}()
		if err != nil {
			return pkg, err
		}
	}
	return pkg, nil
}

func checkPackage(file string, checksum string) error {
	if checksum == "" {
		return nil
	}

	f, err := os.Open(file)
	if err != nil {
		return err
	}

	h := md5.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return err
	}

	if hex.EncodeToString(h.Sum(nil)) != checksum {
		return errors.New("Checksum not match ")
	}
	return nil
}

func getPackageInfo(path string) (*ZipPackageInfo, error) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	for _, file := range reader.File {
		if filepath.Base(file.Name) == ZIP_INFO_FILENAME {
			return func() (*ZipPackageInfo, error) {
				rc, err := file.Open()
				if err != nil {
					return nil, err
				}
				defer rc.Close()
				var buf bytes.Buffer
				buf.Grow(int(file.UncompressedSize64))
				_, err = buf.ReadFrom(rc)
				if err != nil {
					return nil, err
				}
				return readZipInfo(buf.Bytes())
			}()
		}
	}
	return nil, errors.New("pkg.info not found")
}

func (inst *ZipInstaller) Uninstall(pkg Package, del bool) error {
	return pkg.Uninstall(del)
}

func (r *ZipPackageInfo) GetName() string {
	return r.Name
}

// 获得安装包版本号
func (r *ZipPackageInfo) GetVersion() int {
	return r.AppVersion
}

// 获得安装包描述
func (r *ZipPackageInfo) GetDescription() string {
	return r.Description
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

func (pkg *ZipPackage) GetInstallPath() string {
	return pkg.InstallPath
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

func (pkg *ZipPackage) Equal(other Package) bool {
	if pkg.Name != other.GetName() {
		return false
	}
	if pkg.Version != other.GetVersion() {
		return false
	}
	if pkg.InstallPath != other.GetInstallPath() {
		return false
	}
	if pkg.Info != other.GetInfo() {
		return false
	}
	return true
}

func readZipInfo(data []byte) (*ZipPackageInfo, error) {
	ret := &ZipPackageInfo{}
	err := json.Unmarshal(data, ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

type DefaultStrategy struct{}

func NewStrategy() *DefaultStrategy {
	return &DefaultStrategy{}
}

func (s *DefaultStrategy) GenInstallPath(dir string, pkgInfo PackageInfo) (string, error) {
	ret := filepath.Join(dir, pkgInfo.GetName())
	if !io2.IsPathExists(ret) {
		return ret, io2.Mkdir(ret)
	}
	return ret, nil
}
