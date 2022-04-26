package exploration

import (
	"log"
	"path/filepath"
	"sync"

	"github.com/liamvdv/sharedHome/vfs"
	"github.com/spf13/afero"
)

type task struct {
	// stores local file
	abspath string
	// stores enriched file of dir.
	dir *vfs.File
}

type linkedTask struct {
	next *linkedTask
	elem task
}

type taskStack struct {
	// number of items
	n   int
	top *linkedTask
}

func (stk *taskStack) push(t task) {
	stk.n++
	stk.top = &linkedTask{
		elem: t,
		next: stk.top,
	}
}

func (stk *taskStack) pop() task {
	if stk.n < 1 {
		panic("poping from empty stack")
	}
	stk.n--
	ret := stk.top.elem
	stk.top = stk.top.next
	return ret
}

func (stk *taskStack) len() int {
	return stk.n
}

/*
Basic explore: Dvfs
func Explore(root string, ignores []string, next chan<-)
stack of dirTask,
Goal: Traverse file system and send every folder - that should not be ignored - to next stage (i. e. index generator) in the form the application can understand.
*/

// Explore insertes root as the first dir to explore, but you must add it to jobs.
// Usage: jobs.Add(1)
//		  Explore()
func Explore(fs afero.Fs, root string, ignores []string, jobs *sync.WaitGroup, next chan<- *vfs.File, errc chan<- error) {
	globalIgnore := getGlobalIgnoreFunc(ignores)
	lRoot := len(root)

	r := vfs.File{Relpath: "/"}
	if err := vfs.Enrich(fs, root, &r); err != nil {
		errc <- err
		return
	}
	stack := taskStack{}
	stack.push(task{root, &r})

	for stack.len() != 0 {
		
		t := stack.pop()
		dp := t.abspath
		d := t.dir
		log.Printf("ran %q", dp)

		if d.State == vfs.Ignored {
			continue
		}

		dir, err := fs.Open(dp)
		if err != nil {
			errc <- err
		}
		names, rErr := dir.Readdirnames(-1)
		if err := dir.Close(); err != nil {
			errc <- err
		}
		if rErr != nil {
			errc <- err
			return
		}

		ignore, err := getIgnoreFunc(fs, dp, names)
		if err != nil {
			errc <- err
			continue
		}
		d.Children = make([]vfs.File, 0, len(names))

		for _, name := range names {
			fp := filepath.Join(dp, name)

			excl := ignore(name) || globalIgnore(name)
			if excl {
				// TODO(liamvdv): what todo with ignored files?
				log.Printf("Ignored: %s\n", fp)
			}

			f := vfs.File{Relpath: vfs.CleanPath(fp[lRoot:])}
			if err := vfs.Enrich(fp, &f); err != nil {
				errc <- err
				continue
			}

			if excl {
				f.State = vfs.Ignored
			}

			d.Children = append(d.Children, f)
			if f.Mode.IsDir() {
				ref := &d.Children[len(d.Children)-1] // important: f is a local variable
				stack.push(task{fp, ref})
				jobs.Add(1)
			}
		}
		next <- d
		jobs.Done()
	}
	close(next)
}