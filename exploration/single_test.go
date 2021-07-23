package exploration_test

import (
	"bytes"
	"path/filepath"
	"sync"
	"testing"

	"github.com/liamvdv/sharedHome/exploration"
	"github.com/liamvdv/sharedHome/fs"
	"github.com/liamvdv/sharedHome/index"
)

func TestSingleExplore(t *testing.T) {
	root, _ := filepath.Abs("./testdata")
	ignores := []string{
		"notes.txt",
	}
	dirc := make(chan *fs.File)
	errc := make(chan error)

	var jobs sync.WaitGroup
	jobs.Add(1)
	go exploration.Explore(root, ignores, &jobs, dirc, errc)

	idx := index.New()
	go idx.BuildFromChannel(dirc)

	jobs.Wait()
	close(errc)

	idx.Mu.RLock()
	r, err := idx.GetDir("/")
	idx.Mu.RUnlock()
	if err != nil {
		t.Error(err)
	}
	var buf bytes.Buffer
	if err := fs.PrettyPrint(&buf, r); err != nil {
		t.Error(err)
	}
	// TODO(liamvdv): automatic test for validating the build index.
	t.Log("Needs implementation.")
}
