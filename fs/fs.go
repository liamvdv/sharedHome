package fs

import (
	"fmt"
	"io"
	stdfs "io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type File struct {
	// Relpath stores the relative path to Root of the Exploration. No trailing slashes at the end.
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
	// State stores the state of the File needed by syncronisation.
	State State
}

func (f *File) Base() string {
	return filepath.Base(f.Relpath)
}

func (f *File) Dir() string {
	return filepath.Dir(f.Relpath)
}

func (f *File) DirBase() (dirpath string, name string) {
	dirpath = filepath.Dir(f.Relpath)
	return dirpath, f.Relpath[len(dirpath)+1:] // len('/') = 1
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

// CleanPath turns any path into a path sharedHome use internally, see consistency.txt.
// relpath = CleanPath(abspath[len(root):])
func CleanPath(relpath string) string {
	l := len(relpath)
	if l == 0 {
		panic("relpath empty")
	}
	if l == 1 {
		return "/"
	}
	if r := relpath[l-1]; r == '/' || r == '\\' {
		return strings.ReplaceAll(relpath[:l-1], `\`, `/`)
	}
	return strings.ReplaceAll(relpath, `\`, `/`) // on windows: replace, on linux: nothing happens
}

// returns the local platform specific representation of CleanPath.
func LocalPath(relpath string) string {
	const sep = string(os.PathSeparator)
	return strings.ReplaceAll(relpath, "/", sep)
}

func PrettyPrint(w io.Writer, f *File) error {
	// need to sperate for identWriter to work properly...
	fmt.Fprintln(w, "Relpath:", f.Relpath)
	fmt.Fprintf(w, "CTime: %s\n", time.Unix(0, f.CTime))
	fmt.Fprintf(w, "MTime: %s\n", time.Unix(0, f.MTime))
	fmt.Fprintln(w, "Mode:", f.Mode.String())
	fmt.Fprintf(w, "Inode: %d\n", f.Inode)
	fmt.Fprintf(w, "Size: %d\n", f.Size)
	fmt.Fprintln(w, "State:", f.State.String())

	if !f.Mode.IsDir() {
		return nil
	}
	fmt.Fprintf(w, "Children [\n")
	iw := identWriter{w}
	for _, fc := range f.Children {
		err := PrettyPrint(iw, &fc)
		if err != nil {
			return err
		}
	}
	fmt.Fprintln(w, "]")
	return nil
}

type identWriter struct {
	dst io.Writer
}

func (iw identWriter) Write(buf []byte) (int, error) {
	return iw.dst.Write(append([]byte{'\t'}, buf...))
}
