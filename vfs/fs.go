package vfs

import (
	"fmt"
	stdfs "io/fs"
	"path/filepath"
	"time"

	"github.com/liamvdv/sharedHome/osx"
)

// internal paths are handled with stdlib/path
// host specific paths are handled with stdlib/path/filepath
// conversion is handled with stdlib/path/filepath ToSlash() and FromSlash()

type File struct {
	// Relpath stores the relative path to root in the generalised slash form.
	// It always starts with "/" and may not end with "/" except it's root.
	// Convert with:  filepath.ToSlash(abspath[len(root):]))
	// and back with: filepath.Join(root, filepath.FromSlash(relpath))
	Relpath string
	// CTime is the creation datetime in unixnano.
	CTime int64
	// MTime is the last modification datetime in unixnano.
	MTime int64
	// Mode is the uint32 file mode storing file permission and file type.
	Mode stdfs.FileMode
	// Inode is the system dependent inode. On Linux and Mac that is inode, on windows it is the NTFS file id.
	Inode uint64
	// Size is the size of the original file in bytes. For directories it is the number of children.
	Size int64
	// Children is nil for normal files. For directories, it is a slice of all childrens lexiographically ordered by name.
	Children []File
	// State stores the state of the File needed by synchronization.
	State State
}

func (f *File) Base() string {
	return filepath.Base(f.Relpath)
}

func (f *File) Dir() string {
	return filepath.Dir(f.Relpath)
}

func (f *File) DirBase() (dirpath string, base string) {
	dirpath = filepath.Dir(f.Relpath)
	base = f.Relpath[len(dirpath)+1:] // len('/') = 1
	return
}

// Compare returns true if the files are the same. The state and children are not compared.
func (a *File) Equals(b *File) bool {
	return a.Inode == b.Inode &&
		a.Relpath == b.Relpath &&
		a.CTime == b.CTime &&
		a.MTime == b.MTime &&
		a.Mode == b.Mode &&
		a.Size == b.Size
}

func (f File) String() string {
	return fmt.Sprintf("{rp: %q %s s: %s mtime: %s}", f.Relpath, f.Mode, f.State, time.Unix(0, f.MTime))
}

type State uint16

const (
	Unchecked State = iota
	Unmodified
	Modified
	Deleted

	Ignored
)

var toString = []string{
	Unchecked:  "Unchecked",
	Unmodified: "Unmodified",
	Modified:   "Modified",
	Deleted:    "Deleted",
	Ignored:    "Ignored",
}

func (s State) String() string {
	if i := int(s); !(0 <= i && i < len(toString)) {
		panic("invalid state")
	}
	return toString[s]
}

func enrichMock(fs osx.Fs, abspath string, f *File) error {
	fi, err := fs.Stat(abspath)
	if err != nil {
		return err
	}
	f.CTime = fi.ModTime().UnixNano()
	f.MTime = fi.ModTime().UnixNano()
	f.Mode = fi.Mode()
	f.Size = fi.ModTime().UnixNano()
	// f.Inode = ?
	return nil
}
