package osx

import (
	"os"
	"path/filepath"
	"time"
)

const HostPathSeparator = string(filepath.Separator)

// OsFs is a Fs implementation that uses functions provided by the os package.
//
// For details in any method, check the documentation of the os package
// (http://golang.org/pkg/os/).
type OsFs struct{}

func NewOsFs() Fs {
	return &OsFs{}
}

func (OsFs) Name() string { return "OsFs" }

func (OsFs) Create(name string) (File, error) {
	f, e := os.Create(name)
	if f == nil {
		// return nil, not nil value interface
		return nil, e
	}
	return f, e
}

func (OsFs) Mkdir(name string, perm os.FileMode) error {
	return os.Mkdir(name, perm)
}

func (OsFs) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (OsFs) Open(name string) (File, error) {
	f, e := os.Open(name)
	if f == nil {
		// return nil, not nil value interface
		return nil, e
	}
	return f, e
}

func (OsFs) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	f, e := os.OpenFile(name, flag, perm)
	if f == nil {
		// return nil, not nil value interface
		return nil, e
	}
	return f, e
}

func (OsFs) Remove(name string) error {
	return os.Remove(name)
}

func (OsFs) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (OsFs) Rename(oldname, newname string) error {
	return os.Rename(oldname, newname)
}

func (OsFs) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (OsFs) Chmod(name string, mode os.FileMode) error {
	return os.Chmod(name, mode)
}

func (OsFs) Chown(name string, uid, gid int) error {
	return os.Chown(name, uid, gid)
}

func (OsFs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return os.Chtimes(name, atime, mtime)
}

// inserted

func (OsFs) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (OsFs) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}

func (OsFs) ReadDir(name string) ([]os.DirEntry, error) {
	return os.ReadDir(name)
}

func (OsFs) CreateTemp(dir, pattern string) (File, error) {
	return os.CreateTemp(dir, pattern)
}

// end

func (OsFs) LstatIfPossible(name string) (os.FileInfo, bool, error) {
	fi, err := os.Lstat(name)
	return fi, true, err
}

func (OsFs) SymlinkIfPossible(oldname, newname string) error {
	return os.Symlink(oldname, newname)
}

func (OsFs) ReadlinkIfPossible(name string) (string, error) {
	return os.Readlink(name)
}
