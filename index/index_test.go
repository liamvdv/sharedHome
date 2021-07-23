package index_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/liamvdv/sharedHome/fs"
	"github.com/liamvdv/sharedHome/index"
)

// DO NOT REORDER
var testFs = []fs.File{
	{Relpath: "/", CTime: 1626984593612799116, MTime: 1626984593612799116, Mode: 0x800001ed, Inode: 0x2417c8, Size: 4096, Children: []fs.File{
		{Relpath: "/a.txt", CTime: 1626984593612799116, MTime: 1626984593612799116, Mode: 0x1a4, Inode: 0x2417cd, Size: 0, Children: []fs.File(nil), State: 0x0},
		{Relpath: "/b.txt", CTime: 1626984593612799116, MTime: 1626984593612799116, Mode: 0x1a4, Inode: 0x2417ce, Size: 0, Children: []fs.File(nil), State: 0x0},
		{Relpath: "/c.txt", CTime: 1626984593612799116, MTime: 1626984593612799116, Mode: 0x1a4, Inode: 0x2417cf, Size: 0, Children: []fs.File(nil), State: 0x0},
		{Relpath: "/docs", CTime: 1627029843270291407, MTime: 1627029843270291407, Mode: 0x800001ed, Inode: 0x2417c9, Size: 4096, Children: []fs.File{
			{Relpath: "/docs/.notshared", CTime: 1627029885196960326, MTime: 1627029885196960326, Mode: 0x1a4, Inode: 0x24160e, Size: 43, Children: []fs.File(nil), State: 0x0},
			{Relpath: "/docs/d.pdf", CTime: 1626984636142799325, MTime: 1626984636142799325, Mode: 0x1a4, Inode: 0x2417d2, Size: 0, Children: []fs.File(nil), State: 0x4},
			{Relpath: "/docs/e.img", CTime: 1626984636142799325, MTime: 1626984636142799325, Mode: 0x1a4, Inode: 0x2417d3, Size: 0, Children: []fs.File(nil), State: 0x4},
			{Relpath: "/docs/hpi", CTime: 1626984567669465653, MTime: 1626984567669465653, Mode: 0x800001ed, Inode: 0x2417ca, Size: 4096, Children: []fs.File{
				{Relpath: "/docs/hpi/application", CTime: 1626984666426132805, MTime: 1626984666426132805, Mode: 0x800001ed, Inode: 0x2417cb, Size: 4096, Children: []fs.File{
					{Relpath: "/docs/hpi/application/studierfaehigkeitstest.pdf", CTime: 1626984661722799444, MTime: 1626984661722799444, Mode: 0x1a4, Inode: 0x2417d4, Size: 0, Children: []fs.File(nil), State: 0x0},
					{Relpath: "/docs/hpi/application/notes.txt", CTime: 1626984666426132805, MTime: 1626984666426132805, Mode: 0x1a4, Inode: 0x2417d5, Size: 0, Children: []fs.File(nil), State: 0x0},
					{Relpath: "/docs/hpi/application/wise202122", CTime: 1626984567669465653, MTime: 1626984567669465653, Mode: 0x800001ed, Inode: 0x2417cc, Size: 4096, Children: []fs.File{}, State: 0x0},
				}, State: 0x0},
			}, State: 0x0},
			{Relpath: "/docs/tum", CTime: 1626984620276132579, MTime: 1626984620276132579, Mode: 0x800001ed, Inode: 0x2417d0, Size: 4096, Children: []fs.File{
				{Relpath: "/docs/tum/application", CTime: 1626984620279465912, MTime: 1626984620279465912, Mode: 0x800001ed, Inode: 0x2417d1, Size: 4096, Children: []fs.File{
					{Relpath: "/docs/tum/application/wise202122", CTime: 1626984695656132938, MTime: 1626984695656132938, Mode: 0x800001ed, Inode: 0x260ef0, Size: 4096, Children: []fs.File{
						{Relpath: "/docs/tum/application/wise202122/inform.txt", CTime: 1626984695656132938, MTime: 1626984695656132938, Mode: 0x1a4, Inode: 0x261264, Size: 0, Children: []fs.File(nil), State: 0x0},
						{Relpath: "/docs/tum/application/wise202122/unkown.txt", CTime: 1626984691062799588, MTime: 1626984691062799588, Mode: 0x1a4, Inode: 0x261049, Size: 0, Children: []fs.File(nil), State: 0x0},
					}, State: 0x0},
				}, State: 0x0},
			}, State: 0x0},
		}, State: 0x0},
	}, State: 0x0},
}

func fakeExploration(dirc chan<- *fs.File, wg *sync.WaitGroup) {
	// do not reorder, this is real life insertion order. Easier to grasp than loop and checking for IsDir,....
	dirc <- &testFs[0]
	dirc <- &testFs[0].Children[3]
	dirc <- &testFs[0].Children[3].Children[3]                         // hpi
	dirc <- &testFs[0].Children[3].Children[4]                         // tum
	dirc <- &testFs[0].Children[3].Children[3].Children[0]             // hpi/application
	dirc <- &testFs[0].Children[3].Children[3].Children[0].Children[2] // hpi/application/wise202122
	dirc <- &testFs[0].Children[3].Children[4].Children[0]             // tum/application
	dirc <- &testFs[0].Children[3].Children[4].Children[0].Children[0] // tum/application/wise202122
	close(dirc)
	wg.Done()
}

func TestLoadAndStoreIndex(t *testing.T) {
	dirc := make(chan *fs.File)
	var wg sync.WaitGroup
	wg.Add(1)
	go fakeExploration(dirc, &wg)

	idx := index.New()
	go idx.BuildFromChannel(dirc)

	wg.Wait()

	var file bytes.Buffer
	if err := idx.StoreTo(&file); err != nil {
		t.Error(err)
		return
	}
	reIdx, err := index.LoadFrom(&file)
	if err != nil {
		t.Error(err)
		return
	}
	diffs := idx.DetailedEquals(reIdx)
	if len(diffs) > 0 {
		for _, diff := range diffs {
			t.Error(diff)
		}
		var buf bytes.Buffer
		aroot, err := idx.GetDir("/")
		if err == nil {
			fmt.Fprintln(&buf, "Index A:")
			fs.PrettyPrint(&buf, aroot)
		}
		broot, err := reIdx.GetDir("/")
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
