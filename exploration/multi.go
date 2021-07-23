package exploration

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/liamvdv/sharedHome/config"
	"github.com/liamvdv/sharedHome/errors"
	"github.com/liamvdv/sharedHome/fs"
	"github.com/liamvdv/sharedHome/util"
)

// exploration is the struct holding all the shared resources and information of
// the workers in the exploration stage.
type exploration struct {
	root         string
	globalIgnore IgnoreFunc

	dirpaths chan string
	jobwg    sync.WaitGroup

	errc chan error
	//errwg sync.WaitGroup

	Next chan<- fs.File
}

type dirTask struct {
	abspath string
	data    *fs.File
}

func New(root string, ignores []string, nWorkers int, next chan<- fs.File) *exploration {
	const op = errors.Op("exploration.New")

	exp := exploration{
		root:         root,
		globalIgnore: getGlobalIgnoreFunc(ignores),
		// exploreDirs:  make(chan dirTask, nWorkers),
		errc:         make(chan error, nWorkers),
		Next:         next,
	}

	go errorPersister(exp.errc)

	for i := 1; i <= nWorkers; i++ {
		go exp.explorer()
	}

	return &exp
}

func (exp *exploration) add(fp string) {
	exp.jobwg.Add(1)
	exp.dirpaths <- exp.root
}

func (exp *exploration) Start() {
	exp.add(exp.root)
	// this implementation requires workers to call Done on the dir job when children dirs have been added.
	go func() {
		exp.jobwg.Wait()
		close(exp.dirpaths) //  will stop workers
		close(exp.errc)     // will stop error persister
	}()
}


func (exp *exploration) explorer() {
	const op = errors.Op("exploration.explore")

	errc := exp.errc
	dirpaths := exp.dirpaths
	globalIgnore := exp.globalIgnore
	// next := exp.Next

	for dp := range dirpaths {
		dir, err := os.Open(dp)
		if err != nil {
			errc <- errors.E(op, errors.NotExist, errors.Path(dp), err)
			continue
		}

		names, err := dir.Readdirnames(-1) // make full array, not os.ReadDir because calls stat on all, not wanted for excluded names.
		if err != nil {
			errc <- errors.E(op, errors.Path(dp), err)
			continue // abbort if not fully read
		}
		ignore, err := getIgnoreFunc(dp, names)
		if err != nil {
			errc <- err // already error.E
		}

		// dh := fs.DirHeader{
		// 	Name:     filepath.Base(dp),
		// 	Relpath:  dp[len(exp.root):],
		// 	Children: make([]fs.FileHeader, 0, len(names)), // may allocate a few too much, but copying more inefficient.
		// }

		for _, name := range names {
			excl := globalIgnore(name) || ignore(name)

			fp := filepath.Join(dp, name)
			if excl {
				// TODO(liamvdv): do more?
				log.Printf("Ignored: %s\n", fp)
				continue
			}

			fi, err := os.Stat(fp)
			if err != nil {
				errc <- errors.E(op, errors.Path(fp), "Cannot obtain stat.", err)
				continue
			}

			if fi.IsDir() {
				exp.jobwg.Add(1)
				dirpaths <- dp
				continue
			}
		// 	fh := fs.FileHeader{
		// 		Name:    name,
		// 		Size:    uint64(fi.Size()),
		// 		ModTime: uint64(fi.ModTime().UnixNano()),
		// 		Inode:   getInode(&fi), // plattform specific
		// 	}
		}
		// 	dh.Children = append(dh.Children, fh)
		// }
		// next <- dh // works because FileHeaders Array lives on heap.

		// last to prevent wg counter == 0 before really finishing.
		exp.jobwg.Done()
	}
}

func errorPersister(errc <-chan error) {
	fp := filepath.Join(config.LogFolder, "exploration.log")
	file, err := os.OpenFile(fp, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
	if err != nil {
		panic(err)
	}
	defer util.SaveClose(file)
	// log to file
	log.SetOutput(file)

	// consume errors and write to log.
	for err := range errc {
		if config.TESTING {
			fmt.Println(err)
		}
		log.Println(err.Error())
	}
}
