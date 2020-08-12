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
	"io/ioutil"
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

func (inst *ZipInstaller) Install(path string) (Package, error) {
	pkg := &ZipPackage{}
	info, err := getPackageInfo(path)
	if err != nil {
		return nil, err
	}
	pkg.Name = info.Name
	pkg.Version = info.AppVersion
	pkg.Info = info.Info

	saveDir := filepath.Join(inst.installDir, info.Name)
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
			} else {
				_, err = io.Copy(w, rc)
				return err
			}
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
	return Uninstall(del)
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
	r.pkgs[GetName()] = pkg
	return r.flush()
}

func (r *ZipRecorder) Remove(pkg Package) error {
	delete(r.pkgs, GetName())
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

func readZipInfo(data []byte) (*ZipPackageInfo, error) {
	ret := &ZipPackageInfo{}
	err := json.Unmarshal(data, ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}
