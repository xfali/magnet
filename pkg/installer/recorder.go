// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package installer

import (
	"encoding/json"
	"github.com/xfali/goutils/io"
	"io/ioutil"
	"os"
	"sync"
)

type ZipRecorder struct {
	path string
	pkgs map[string]Package

	lock sync.Mutex
}

func CreateRecorder(path string) (*ZipRecorder, error) {
	ret := &ZipRecorder{
		path: path,
		pkgs: map[string]Package{},
	}
	if io.IsPathExists(path) {
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
	r.lock.Lock()
	defer r.lock.Unlock()

	r.pkgs[pkg.GetName()] = pkg
	return r.flush()
}

func (r *ZipRecorder) Remove(pkg Package) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	delete(r.pkgs, pkg.GetName())
	return r.flush()
}

func (r *ZipRecorder) ListPackage() []Package {
	r.lock.Lock()
	defer r.lock.Unlock()

	ret := make([]Package, 0, len(r.pkgs))
	for _, v := range r.pkgs {
		ret = append(ret, v)
	}
	return ret
}

func (r *ZipRecorder) GetPackage(name string) []Package {
	r.lock.Lock()
	defer r.lock.Unlock()

	return []Package{r.pkgs[name]}
}

type JsonRecorder struct {
	path string
	pkgs map[string][]Package

	lock sync.Mutex
}

func CreateJsonRecorder(path string) (*JsonRecorder, error) {
	ret := &JsonRecorder{
		path: path,
		pkgs: map[string][]Package{},
	}
	if io.IsPathExists(path) {
		d, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}
		tmp := map[string][]*ZipPackage{}
		err = json.Unmarshal(d, &tmp)
		if err != nil {
			return nil, err
		}
		for k, v := range tmp {
			pkgs := make([]Package, len(v))
			for i := range v {
				pkgs[i] = v[i]
			}
			ret.pkgs[k] = pkgs
		}
	}

	return ret, nil
}

func (r *JsonRecorder) flush() error {
	d, err := json.Marshal(r.pkgs)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(r.path, d, os.ModePerm)
}

func (r *JsonRecorder) Save(pkg Package) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	pkgs, ok := r.pkgs[pkg.GetName()]
	for _, v := range pkgs {
		if v.Equal(pkg) {
			return nil
		}
	}
	if ok {
		r.pkgs[pkg.GetName()] = append(pkgs, pkg)
	} else {
		r.pkgs[pkg.GetName()] = []Package{pkg}
	}
	return r.flush()
}

func (r *JsonRecorder) Remove(pkg Package) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	pkgs, ok := r.pkgs[pkg.GetName()]
	if !ok {
		return nil
	}
	for i := range pkgs {
		if pkgs[i].Equal(pkg) {
			pkgs = append(pkgs[:i], pkgs[i+1:]...)
			if len(pkgs) == 0 {
				delete(r.pkgs, pkg.GetName())
			} else {
				r.pkgs[pkg.GetName()] = pkgs
			}
		}
	}

	return r.flush()
}

func (r *JsonRecorder) ListPackage() []Package {
	r.lock.Lock()
	defer r.lock.Unlock()

	ret := make([]Package, 0, len(r.pkgs))
	for _, v := range r.pkgs {
		for _, pkg := range v {
			ret = append(ret, pkg)
		}
	}
	return ret
}

func (r *JsonRecorder) GetPackage(name string) []Package {
	r.lock.Lock()
	defer r.lock.Unlock()

	return r.pkgs[name]
}
