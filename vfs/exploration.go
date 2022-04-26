package vfs

import (
	"path/filepath"
	"sync"

	"github.com/liamvdv/sharedHome/osx"
)

/*
Usage:
	walk := vfs.NewFromWalk(fs, config.Root, globalIgnores)
	...
	go ErrorCollector(walk.Errc)
	...
	index, err := walk.DoAndWait()
*/

// exploration is NOT CONCURRENT currently.
type exploration struct {
	// Root is the root filepath of the exploration.
	Root string
	// Errc must be consumed. It will be closed by exploration.
	Errc chan error
	// index is only made public when it was fully build.
	index *FileIndex

	fs           osx.Fs
	globalIgnore IgnoreFunc
	pathswg      sync.WaitGroup

	// This is subject to change for a concurrent implementation.
	stack stack
}

// NewFromWalk returns an exloration struct. The Errc error channel must be consumed.
func NewFromWalk(fs osx.Fs, root string, ignores []string) *exploration {
	return &exploration{
		Root: root,
		Errc: make(chan error, 1),
		index: &FileIndex{
			Files: make(map[string]*File),
		},

		fs:           fs,
		globalIgnore: getGlobalIgnoreFunc(ignores),

		// This is subject to change when a concurrent multiExplorer is implemented.
		stack: stack{},
	}
}

// DoAndWait blocks until the FileIndex is fully built and then return it.
func (x *exploration) DoAndWait() (*FileIndex, error) {
	// ======= Subject to change, err is there for future proof.
	x.pathswg.Add(1)
	go x.noconcurrentExplorer(x.Root)
	// =======

	x.pathswg.Wait()
	close(x.Errc)
	return x.index, nil
}

func (x *exploration) noconcurrentExplorer(root string) {
	var lnRoot = len(x.Root)

	r := File{Relpath: "/"}
	if err := Enrich(x.fs, root, &r); err != nil {
		x.Errc <- err
		return
	}
	x.stack.push(task{root, &r})

	for x.stack.len() > 0 {
		t := x.stack.pop()
		dp := t.abspath
		d := t.dir

		if d.State == Ignored {
			continue
		}

		dir, err := x.fs.Open(dp)
		if err != nil {
			x.Errc <- err
		}
		names, rErr := dir.Readdirnames(-1)
		if err := dir.Close(); err != nil {
			x.Errc <- err
		}
		if rErr != nil {
			x.Errc <- err
			return
		}

		ignore, err := getIgnoreFunc(x.fs, dp, names)
		if err != nil {
			x.Errc <- err
			continue
		}
		d.Children = make([]File, 0, len(names))

		for _, name := range names {
			fp := filepath.Join(dp, name)

			f := File{
				Relpath: filepath.ToSlash(fp[lnRoot:]),
			}
			if err := Enrich(x.fs, fp, &f); err != nil {
				x.Errc <- err
				continue
			}
			// TODO(liamvdv): excludes dirs in test. Don't yet know why.
			// if !f.Mode.IsRegular() {
			// 	log.Println("is---Not---Regular")
			// 	continue
			// }

			if ignore(name) || x.globalIgnore(name) {
				f.State = Ignored
			}
			// copy
			d.Children = append(d.Children, f)
			if f.Mode.IsDir() {
				ref := &d.Children[len(d.Children)-1] // important: f local var.
				x.stack.push(task{fp, ref})
				x.pathswg.Add(1)
			}
		}
		// update index.
		x.index.Mu.Lock()
		x.index.Files[d.Relpath] = d
		x.index.Mu.Unlock()

		x.pathswg.Done()
	}
}

type stack struct {
	// number of items
	n   int
	top *node
}

type node struct {
	next *node
	elem task
}

type task struct {
	// stores local file
	abspath string
	// stores enriched file of dir.
	dir *File
}

func (s *stack) push(t task) {
	s.n++
	s.top = &node{
		elem: t,
		next: s.top,
	}
}

func (s *stack) pop() task {
	if s.n < 1 {
		panic("poping from empty stack")
	}
	s.n--
	ret := s.top.elem
	s.top = s.top.next
	return ret
}

func (s *stack) len() int {
	return s.n
}
