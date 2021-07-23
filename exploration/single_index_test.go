package exploration_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/liamvdv/sharedHome/exploration"
	"github.com/liamvdv/sharedHome/fs"
	"github.com/liamvdv/sharedHome/index"
)

func TestSingleExploreAndLoadAndStoreIndex(t *testing.T) {
	root, _ := filepath.Abs("./testdata")
	ignores := []string{
		"notes.txt",
	}
	dirc := make(chan *fs.File)
	errc := make(chan error)

	go func() {
		for err := range errc {
			t.Error(err)
		}
	}()

	var jobs sync.WaitGroup
	jobs.Add(1)
	go exploration.Explore(root, ignores, &jobs, dirc, errc)

	originalIdx := index.New()
	go originalIdx.BuildFromChannel(dirc)

	jobs.Wait()
	close(errc)

	var file bytes.Buffer
	if err := originalIdx.StoreTo(&file); err != nil {
		t.Error(err)
		return
	}

	loadedIdx, err := index.LoadFrom(&file)
	if err != nil {
		t.Error(err)
		return
	}
	diffs := originalIdx.DetailedEquals(loadedIdx)
	if len(diffs) > 0 {
		for _, diff := range diffs {
			t.Error(diff)
		}
		var buf bytes.Buffer
		aroot, err := originalIdx.GetDir("/")
		if err == nil {
			fmt.Fprintln(&buf, "Index A:")
			fs.PrettyPrint(&buf, aroot)
		}
		broot, err := loadedIdx.GetDir("/")
		if err == nil {
			fmt.Fprintln(&buf, "Index B:")
			fs.PrettyPrint(&buf, broot)
		}
		fp, _ := filepath.Abs("./index_test_debug.log")
		if err := os.WriteFile(fp, buf.Bytes(), 0600); err == nil {
			t.Logf("stored index A and B to %q\n", fp)
		}
		return
	}
}

// func TestCreateHardCodedTestcases(t *testing.T) {
// 	root, _ := filepath.Abs("./testdata")
// 	ignores := []string{
// 	}
// 	dirc := make(chan *fs.File)
// 	errc := make(chan error)

// 	go func() {
// 		for err := range errc {
// 			t.Error(err)
// 		}
// 	}()

// 	var jobs sync.WaitGroup
// 	jobs.Add(1)
// 	go exploration.Explore(root, ignores, &jobs, dirc, errc)

// 	go func() {
// 		f, err := os.Create("temp.log.txt")
// 		if err != nil {
// 			panic(err)
// 		}
// 		for item := range dirc {
// 			f.WriteString(fmt.Sprintf("%#v\n", *item))
// 		}
// 		if err := f.Close(); err != nil {
// 			panic(err)
// 		}
// 	}()

// 	jobs.Wait()
// 	close(errc)
// }
