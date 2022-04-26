package exploration_test

import (
	"github.com/spf13/afero"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func createFiles(fs afero.Fs, dp string, names ...string) error {
	for _, name := range names {
		fp := dp + "/" + name
		d, err := fs.Create(fp)
		if err != nil {
			return err
		}
		d.Close()
	}
	return nil
}

func testMemFs() afero.Fs {
	memfs := afero.NewMemMapFs()
	memfs.MkdirAll("/home/test/docs/hpi/application/wise202122", 0755)
	must(createFiles(memfs, "/home/test/", "a.pdf", "b.pdf", "c.txt", "d.img"))
	must(createFiles(memfs, "/home/test/docs", "e.pdf", "f.pdf"))
	must(createFiles(memfs, "/home/test/docs/hpi/application/wise202122", "a.pdf", "b.pdf", "c.txt", "d.img"))
	must(memfs.MkdirAll("/home/test/docs/hpi/courses/abc", 0755))
	must(createFiles(memfs, "/home/test/docs/hpi/courses", "programming.md", "data modeling.txt"))
	memfs.MkdirAll("/home/test/docs/tum/application/wise202122", 0755)
	must(createFiles(memfs, "/home/test/docs/hpi/application/wise202122", "n.pdf", "m.pdf", "y.txt", "z.img"))
	return memfs
}

// func TestSingleExplore(t *testing.T) {
// 	testFs := testMemFs()
// 	dirc := make(chan *vfs.File)
// 	errc := make(chan error)

// 	ignores := []string{"tum"}

// 	var jobs sync.WaitGroup
// 	jobs.Add(1)
// 	go exploration.Explore(testFs, "/home/test", ignores, &jobs, dirc, errc)

// 	idx := index.New()
// 	go idx.BuildFromChannel(dirc)

// 	jobs.Wait()
// 	close(errc)

// 	idx.Mu.RLock()
// 	r, err := idx.GetDir("/")
// 	idx.Mu.RUnlock()
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	var buf bytes.Buffer
// 	if err := vfs.PrettyPrint(&buf, r); err != nil {
// 		t.Error(err)
// 	}
// 	// TODO(liamvdv): automatic test for validating the build index.
// 	t.Log("Needs implementation.")
// }