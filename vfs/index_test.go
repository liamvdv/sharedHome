package vfs_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/liamvdv/sharedHome/osx"
	"github.com/liamvdv/sharedHome/vfs"
)

// global ignores
var ignores = []string{
	// state: 0x4
	"d.pdf",
	"e.img",
}

// DO NOT REORDER
var testVfs = vfs.File{Relpath: "/", CTime: 1626984593612799116, MTime: 1626984593612799116, Mode: 0x800001ed, Inode: 0x2417c8, Size: 4096, Children: []vfs.File{
	{Relpath: "/a.txt", CTime: 1626984593612799116, MTime: 1626984593612799116, Mode: 0x1a4, Inode: 0x2417cd, Size: 0, Children: []vfs.File(nil), State: 0x0},
	{Relpath: "/b.txt", CTime: 1626984593612799116, MTime: 1626984593612799116, Mode: 0x1a4, Inode: 0x2417ce, Size: 0, Children: []vfs.File(nil), State: 0x0},
	{Relpath: "/c.txt", CTime: 1626984593612799116, MTime: 1626984593612799116, Mode: 0x1a4, Inode: 0x2417cf, Size: 0, Children: []vfs.File(nil), State: 0x0},
	{Relpath: "/docs", CTime: 1627029843270291407, MTime: 1627029843270291407, Mode: 0x800001ed, Inode: 0x2417c9, Size: 4096, Children: []vfs.File{
		{Relpath: "/docs/.notshared", CTime: 1627029885196960326, MTime: 1627029885196960326, Mode: 0x1a4, Inode: 0x24160e, Size: 43, Children: []vfs.File(nil), State: 0x0},
		{Relpath: "/docs/d.pdf", CTime: 1626984636142799325, MTime: 1626984636142799325, Mode: 0x1a4, Inode: 0x2417d2, Size: 0, Children: []vfs.File(nil), State: 0x4},
		{Relpath: "/docs/e.img", CTime: 1626984636142799325, MTime: 1626984636142799325, Mode: 0x1a4, Inode: 0x2417d3, Size: 0, Children: []vfs.File(nil), State: 0x4},
		{Relpath: "/docs/hpi", CTime: 1626984567669465653, MTime: 1626984567669465653, Mode: 0x800001ed, Inode: 0x2417ca, Size: 4096, Children: []vfs.File{
			{Relpath: "/docs/hpi/application", CTime: 1626984666426132805, MTime: 1626984666426132805, Mode: 0x800001ed, Inode: 0x2417cb, Size: 4096, Children: []vfs.File{
				{Relpath: "/docs/hpi/application/studierfaehigkeitstest.pdf", CTime: 1626984661722799444, MTime: 1626984661722799444, Mode: 0x1a4, Inode: 0x2417d4, Size: 0, Children: []vfs.File(nil), State: 0x0},
				{Relpath: "/docs/hpi/application/notes.txt", CTime: 1626984666426132805, MTime: 1626984666426132805, Mode: 0x1a4, Inode: 0x2417d5, Size: 0, Children: []vfs.File(nil), State: 0x0},
				{Relpath: "/docs/hpi/application/wise202122", CTime: 1626984567669465653, MTime: 1626984567669465653, Mode: 0x800001ed, Inode: 0x2417cc, Size: 4096, Children: []vfs.File{}, State: 0x0},
			}, State: 0x0},
		}, State: 0x0},
		{Relpath: "/docs/tum", CTime: 1626984620276132579, MTime: 1626984620276132579, Mode: 0x800001ed, Inode: 0x2417d0, Size: 4096, Children: []vfs.File{
			{Relpath: "/docs/tum/application", CTime: 1626984620279465912, MTime: 1626984620279465912, Mode: 0x800001ed, Inode: 0x2417d1, Size: 4096, Children: []vfs.File{
				{Relpath: "/docs/tum/application/wise202122", CTime: 1626984695656132938, MTime: 1626984695656132938, Mode: 0x800001ed, Inode: 0x260ef0, Size: 4096, Children: []vfs.File{
					{Relpath: "/docs/tum/application/wise202122/inform.txt", CTime: 1626984695656132938, MTime: 1626984695656132938, Mode: 0x1a4, Inode: 0x261264, Size: 0, Children: []vfs.File(nil), State: 0x0},
					{Relpath: "/docs/tum/application/wise202122/unkown.txt", CTime: 1626984691062799588, MTime: 1626984691062799588, Mode: 0x1a4, Inode: 0x261049, Size: 0, Children: []vfs.File(nil), State: 0x0},
				}, State: 0x0},
			}, State: 0x0},
		}, State: 0x0},
	}, State: 0x0},
}, State: 0x0}

func TestFileIndexLoadStore(t *testing.T) {
	index := vfs.NewFromMemory(&testVfs)

	var file bytes.Buffer
	if err := index.Store(&file); err != nil {
		t.Error(err)
	}
	index2, err := vfs.Load(&file)
	if err != nil {
		t.Error(err)
	}

	if diffs := index.Equals(index2); len(diffs) != 0 {
		t.Error("stored and loaded index are different.")
		t.Log(strings.Join(diffs, "\n"))
	}
}

// You cannot compare Size, CTime and Inodes. Remember to RemoveAll(rootpath).
func buildTestFs(fs osx.Fs, rootpath string, root *vfs.File, t *testing.T) {
	// need to do this because creating files in a dir will change the modtime, which we want to control.
	type chmodCall struct {
		dp    string
		atime time.Time
		mtime time.Time
	}
	var chmodAfterBuild []chmodCall

	queue := make([]*vfs.File, 0, 8) // guess
	queue = append(queue, root)
	var dir *vfs.File
	for l := len(queue); l > 0; l = len(queue) {
		dir, queue = queue[0], queue[1:]

		dp := filepath.Join(rootpath, filepath.FromSlash(dir.Relpath))
		if err := fs.Mkdir(dp, dir.Mode); err != nil {
			if os.IsExist(err) {
				err = fs.Chmod(dp, dir.Mode)
				if err != nil {
					t.Error(err)
				}
			} else {
				t.Error(err)
			}
		}
		chmodAfterBuild = append(chmodAfterBuild, chmodCall{
			dp:    dp,
			atime: time.Now(),
			mtime: time.Unix(0, dir.MTime),
		})
		var f *vfs.File
		for i := range dir.Children {
			f = &dir.Children[i]
			if f.Mode.IsDir() {
				queue = append(queue, f)
				continue
			}
			if err := buildFile(fs, rootpath, f); err != nil {
				t.Error(err)
			}
		}

		// do it here because adding files/folders changes the modtime.
		for _, c := range chmodAfterBuild {
			if err := fs.Chtimes(c.dp, c.atime, c.mtime); err != nil {
				t.Error(err)
			}
		}
	}
}

func buildFile(fs osx.Fs, root string, file *vfs.File) error {
	fp := filepath.Join(root, filepath.FromSlash(file.Relpath))
	if err := fs.WriteFile(fp, []byte("Whatever!"), file.Mode); err != nil {
		return err
	}
	if err := fs.Chtimes(fp, time.Now(), time.Unix(0, file.MTime)); err != nil {
		return err
	}
	return nil
}

// Does not validate .notshared behavior because we don't have the file content.
func TestFileIndexFromWalk(t *testing.T) {
	wantIndex := vfs.NewFromMemory(&testVfs)

	fs := osx.NewOsFs()
	dirpath := testDir(fs)
	defer removeAllTestFiles(t)
	buildTestFs(fs, dirpath, &testVfs, t)

	exp := vfs.NewFromWalk(fs, dirpath, ignores)

	var errwg sync.WaitGroup
	errwg.Add(1)
	go func(errc <-chan error, wg *sync.WaitGroup) {
		for err := range exp.Errc {
			t.Error(err)
		}
		wg.Done()
	}(exp.Errc, &errwg)

	gotIndex, err := exp.DoAndWait()
	if err != nil {
		t.Error(err)
	}
	errwg.Wait()

	if diffs := wantIndex.Equals(gotIndex); len(diffs) != 0 {
		t.Error(strings.Join(diffs, "\n"))
		if err := wantIndex.Print(os.Stdout); err != nil {
			t.Errorf("failed to print index to stdout because %v", err)
		}
		if err := gotIndex.Print(os.Stdout); err != nil {
			t.Errorf("failed to print index to stdout because %v", err)
		}
	}
}
