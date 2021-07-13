package fs

import (
	stdfs "io/fs"
	"path/filepath"
)

type File struct {
	Relpath string
	CTime   int64
	MTime   int64
	Mode    stdfs.FileMode
	Inode   uint64
	Size    int64

	Children []File
	State    State
}

func (f *File) Name() string {
	return filepath.Base(f.Relpath)
}

type State uint16

const (
	Unchecked State = iota
	Unmodified
	Modified
	Deleted

	Ignored
)
