package vfs

import (
	"encoding/gob"
	"fmt"
	"io"
	"path"
	"sync"
	"time"

	"github.com/liamvdv/sharedHome/errors"
)

type FileIndex struct {
	Mu    sync.RWMutex
	Files map[string]*File
}

// Please take a look at exploration.go for NewFromWalk() function implementation.

func NewFromMemory(root *File) *FileIndex {
	index := FileIndex{Files: make(map[string]*File)}
	index.Mu.Lock()
	defer index.Mu.Unlock()

	stack := make([]*File, 0, 10)
	stack = append(stack, root)
	var dir *File
	for l := len(stack); l > 0; l = len(stack) {
		stack, dir = stack[:l-1], stack[l-1]

		index.Files[dir.Relpath] = dir
		for n := range dir.Children {
			if dir.Children[n].Mode.IsDir() {
				stack = append(stack, &dir.Children[n])
			}
		}
	}
	return &index
}

func Load(r io.Reader) (*FileIndex, error) {
	root := File{}
	if err := gob.NewDecoder(r).Decode(&root); err != nil {
		return nil, err
	}
	return NewFromMemory(&root), nil
}

func (i *FileIndex) Store(w io.Writer) error {
	i.Mu.RLock()
	defer i.Mu.RUnlock()
	root, err := i.GetDir("/")
	if err != nil {
		return nil
	}
	// root references every node (most indirectly). gob will traverse all nodes
	return gob.NewEncoder(w).Encode(root)
}

var (
	ErrFileNotFound = errors.E("file not found")
)

// Get can be used to retrieve both a dir file and a normal file.
func (i *FileIndex) Get(relpath string) (*File, error) {
	if relpath == "/" {
		return i.GetDir("/")
	}
	dp := path.Dir(relpath)
	i.Mu.RLock()
	defer i.Mu.RUnlock()
	dir, ok := i.Files[dp]
	if !ok {
		return nil, ErrFileNotFound
	}
	for n := range dir.Children {
		if dir.Children[n].Relpath == relpath {
			return &dir.Children[n], nil
		}
	}
	return nil, ErrFileNotFound
}

// GetDir can only be used to retrieve a dir file. It is faster than Get(string)
func (i *FileIndex) GetDir(relpath string) (*File, error) {
	i.Mu.RLock()
	dir, ok := i.Files[relpath]
	i.Mu.RUnlock()
	if !ok {
		return nil, ErrFileNotFound
	}
	return dir, nil
}

// Equals returns the number of differences as a string array.
// if len(a.Equals(b)) is 0, then they are deep equal.
func (a *FileIndex) Equals(b *FileIndex) (diffs []string) {
	a.Mu.RLock()
	b.Mu.RLock()
	defer a.Mu.RUnlock()
	defer b.Mu.RUnlock()

	if len(a.Files) != len(b.Files) {
		diffs = append(diffs, "number of files in the indexes are different")
	}

	root, err := a.GetDir("/")
	if err != nil {
		return append(diffs, "index A has no root directory")
	}
	stack := make([]*File, 0, len(root.Children))
	stack = append(stack, root)
	var acur, bcur *File
	for l := len(stack); l > 0; l = len(stack) {
		stack, acur = stack[:l-1], stack[l-1]

		bcur, err = b.GetDir(acur.Relpath)
		if err != nil {
			return append(diffs, "index B does not contain dir "+acur.Relpath) // abort
		}

		if !acur.Equals(bcur) {
			diffs = append(diffs, "index B has different version of "+acur.Relpath)
		}

		if len(acur.Children) != len(bcur.Children) {
			diffs = append(diffs, "index B has a different number of children for "+acur.Relpath)
		}

	Loop:
		for _, child := range acur.Children {
			if child.Mode.IsDir() {
				stack = append(stack, &child)
				continue
			}
			for _, bchild := range bcur.Children {
				if child.Equals(&bchild) {
					continue Loop
				}
			}
			diffs = append(diffs, "index B does not contain file "+child.Relpath)
		}
	}
	return

}

func (i *FileIndex) Print(w io.Writer) error {
	root, err := i.GetDir("/")
	if err != nil {
		return err
	}
	return Print(w, root)
}

func Print(w io.Writer, f *File) error {
	// separate calls for indentation
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
	for _, child := range f.Children {
		err := Print(iw, &child)
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
