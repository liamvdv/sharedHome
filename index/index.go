package index

import (
	"encoding/gob"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/liamvdv/sharedHome/errors"
	"github.com/liamvdv/sharedHome/fs"
)

// See the consistency.txt file for information on the pathspec.
/*
	To be implemented: 
	// Random Access:
		Get
		GetDir (faster then Get, only for dirs)
	// Sequential Access:
		traverse with Children propertry -> possible
	// Store to file
	// Load from file

	// sun - sequential update number for sync
*/

// DirIndex holds all directories for fast random access.
// Normal files can be accessed by taking the parent path, looking that one up and then sequential search through the parent's children.
type DirIndex struct {
	Mu	 sync.RWMutex
	Dir   map[string]*fs.File
}

var ErrFileNotFound = errors.E("File not found.")
// Get accepts the generalised path, i. e. with forward slashes '/'. Returns ErrFileNotFound if the parent or file itself cannot be located.
func (idx *DirIndex) Get(relpath string) (*fs.File, error) {
	i := strings.LastIndex(relpath, "/")
	if i == -1 {
		return nil, errors.E("malformed relpath")
	}
	idx.Mu.RLock()
	parent, ok := idx.Dir[relpath[:i]]
	idx.Mu.RUnlock()
	if !ok {
		return nil, ErrFileNotFound
	}
	for i := range parent.Children {
		if parent.Children[i].Relpath == relpath {
			return &parent.Children[i], nil
		}
	}
	return nil, ErrFileNotFound
}

func (idx *DirIndex) GetDir(relpath string) (*fs.File, error) {
	idx.Mu.RLock()
	dir, ok := idx.Dir[relpath]
	idx.Mu.RUnlock()
	if !ok {
		return nil, ErrFileNotFound
	}
	return dir, nil
}

// StoreTo will serialise the current state of the map to the index file.
func (idx *DirIndex) StoreTo(w io.Writer) error {
	idx.Mu.Lock()
	defer idx.Mu.Unlock()
	return gob.NewEncoder(w).Encode(&idx.Dir)
}

// LoadFrom will read the map from a gob encoded stream of data and return DirIndex.
func LoadFrom(r io.Reader) (*DirIndex, error) {
	idx := DirIndex{}
	// di.mutex.Lock()
	// defer di.mutex.Unlock()
	err := gob.NewDecoder(r).Decode(&idx.Dir)
	if err != nil {
		return nil, err
	}
	return &idx, nil
}

func LoadFromFile(fp string) (idx *DirIndex, err error) {
	f, err := os.Open(fp)
	if err != nil {
		return nil, err
	}
	defer func() {
		fErr := f.Close()
		if err != nil {
			return
		}
		err = fErr
	}()
	return LoadFrom(f)
}

func (idx *DirIndex) StoreToFile(fp string) (err error) {
	f, err := os.OpenFile(fp, os.O_CREATE|os.O_RDWR|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer func() {
		fErr := f.Close()
		if err != nil {
			return
		}
		err = fErr
	}()
	return idx.StoreTo(f)
}

func New(dirs... *fs.File) (*DirIndex) {
	idx := DirIndex{
		Dir: make(map[string]*fs.File, len(dirs) + 1000),
	}
	// idx.Mu.Lock()
	for _, dir := range dirs {
		idx.Dir[dir.Relpath] = dir
	}
	// idx.Mu.Unlock()
	return &idx
}
// BuildFrom Locks until the channel is closed.
func (idx *DirIndex) BuildFromChannel(c <-chan *fs.File) {
	idx.Mu.Lock()
	for f := range c {
		idx.Dir[f.Relpath] = f
	}
	idx.Mu.Unlock()
}

// Equals checks if all items in a are also in b at the right postion. b must not contain any other items.
func (a *DirIndex) Equals(b *DirIndex) bool {
	a.Mu.RLock()
	defer a.Mu.Unlock()
	b.Mu.RLock()
	defer b.Mu.RUnlock()
	if len(a.Dir) != len(a.Dir) {
		return false
	}
	if len(a.Dir) == 0 { // catch nil and 0 elems
		return true
	}

	aroot, err := a.GetDir("/")
	if err != nil {
		return false
	}
	stk := make([]*fs.File, 0, len(aroot.Children)) // estimate
	stk = append(stk, aroot)
	var cur *fs.File
	for l := len(stk); l != 0; {
		stk, cur = stk[:l-2], stk[l-1]

		bcur, err := b.GetDir(cur.Relpath) 
		if err != nil {
			return false
		}

		if !cur.Equals(bcur) {
			return false
		}
		// check for Equality, without this check this function is euquivalent to Contains.
		if len(cur.Children) != len(bcur.Children) {
			return false
		}

		for _, child := range cur.Children {
			if child.Mode.IsDir() {
				stk = append(stk, &child)
				continue
			}
			for _, bchild := range bcur.Children {
				if child.Equals(&bchild) {
					continue
				}
			}
			return false
		}
	}
	return true
}

// DetailedEquals is used to provide feedback on what the differences are. If there are no, DetailedEquals
// returns an empty slice.
func (a *DirIndex) DetailedEquals(b *DirIndex) (diffs []string) {
	a.Mu.RLock()
	defer a.Mu.RUnlock()
	b.Mu.RLock()
	defer b.Mu.RUnlock()
	if len(a.Dir) != len(a.Dir) {
		diffs = append(diffs, "the number of elements in the indexes are different.")
	}
	if len(a.Dir) == 0 { // catch nil and 0 elems
		return
	}

	aroot, err := a.GetDir("/")
	if err != nil {
		return append(diffs, "cannot find root dir in A index.") // abort
	}
	stk := make([]*fs.File, 0, len(aroot.Children)) // estimate
	stk = append(stk, aroot)
	var cur *fs.File
	for l := len(stk); l > 0; l = len(stk) {
		stk, cur = stk[:l-1], stk[l-1]

		bcur, err := b.GetDir(cur.Relpath) 
		if err != nil {
			return append(diffs, "index B does not contain dir " + cur.Relpath) // abort
		}

		if !cur.Equals(bcur) {
			diffs = append(diffs, "index B has different version of " + cur.Relpath)
		}

		if len(cur.Children) != len(bcur.Children) {
			diffs = append(diffs, "index B does have a different number of children for " + cur.Relpath)
		}

Loop:
		for _, child := range cur.Children {
			if child.Mode.IsDir() {
				stk = append(stk, &child)
				continue
			}
			for _, bchild := range bcur.Children {
				if child.Equals(&bchild) {
					continue Loop
				}
			}
			diffs = append(diffs, "index B does not contain file " + child.Relpath)
		}
	}
	return
}