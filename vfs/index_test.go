package vfs_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/liamvdv/sharedHome/vfs"
)

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
	index := vfs.NewFrom(&testVfs)

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
