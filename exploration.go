package main

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/liamvdv/sharedHome/config"
	"github.com/liamvdv/sharedHome/errors"
	"github.com/liamvdv/sharedHome/fs"
	. "github.com/liamvdv/sharedHome/util"
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

	Next chan<- fs.DirHeader // next stage
}

func NewExploration(root string, nWorkers int, next chan<- fs.DirHeader) *exploration {
	const op = errors.Op("exploration.New")

	gIgnore, err := getIgnoreFunc(config.Dir, []string{config.IgnorefileName})
	if err != nil {
		panic(err)
	}

	exp := exploration{
		root:         root,
		globalIgnore: gIgnore,
		dirpaths:     make(chan string, nWorkers*8),
		errc:         make(chan error, nWorkers),
		Next:         next,
	}

	go errorPersister(exp.errc)

	for i := 1; i <= nWorkers; i++ {
		go exp.explorer()
	}

	return &exp
}

func (exp *exploration) Start() {
	exp.jobwg.Add(1)
	exp.dirpaths <- exp.root

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
	next := exp.Next

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

		dh := fs.DirHeader{
			Name:     filepath.Base(dp),
			Relpath:  dp[len(exp.root):],
			Children: make([]fs.FileHeader, 0, len(names)), // may allocate a few too much, but copying more inefficient.
		}

		for _, name := range names {
			excl := globalIgnore(name) || ignore(name)

			fp := filepath.Join(dp, name)
			if excl {
				// TODO(liamvdv) do more?
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
			fh := fs.FileHeader{
				Name:    name,
				Size:    uint64(fi.Size()),
				ModTime: uint64(fi.ModTime().UnixNano()),
				Inode:   getInode(&fi), // plattform specific
			}

			dh.Children = append(dh.Children, fh)
		}
		next <- dh // works because FileHeaders Array lives on heap.

		// last to prevent wg counter == 0 before really finishing.
		exp.jobwg.Done()
	}
}

type IgnoreFunc func(string) bool

// getIgnoreFunc returns a function that accepts a name and return whether it
// should be EXCLUDED (true) or INCLUDED (false).
func getIgnoreFunc(dp string, names []string) (IgnoreFunc, error) {
	const op = errors.Op("exploration.getIgnoreFunc")

	there := false
	for _, name := range names {
		if name == config.IgnorefileName {
			there = true
			break
		}
	}

	if !there {
		return func(name string) bool {
			return false
		}, nil
	}

	// exists, read in ignore file
	fp := filepath.Join(dp, config.IgnorefileName)
	file, err := os.Open(fp)
	if err != nil {
		return nil, errors.E(op, errors.Path(fp), err)
	}
	defer SaveClose(file)

	scanner := bufio.NewScanner(file)
	var ignoreNames = make([]string, 0, 5) // sensible default

	for scanner.Scan() {
		s := scanner.Text()
		// empty line or comment
		if s == "" || s == "\r" || strings.HasPrefix(s, "#") {
			continue
		}
		if li := len(s) - 1; s[li] == '\r' {
			s = s[:li]
		}
		// TODO(liamvdv): we currently only accept "filename" and "dirname",
		// NOT "/filename" or "/path/filename" or any **regexp** <- should support regexp!
		// look at golang.org/pkg/path/filepath/#Match
		ignoreNames = append(ignoreNames, s)
	}

	return func(name string) bool {
		for _, ignore := range ignoreNames {
			if name == ignore {
				return true
			}
		}
		return false
	}, nil
}
