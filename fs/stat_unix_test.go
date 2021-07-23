package fs_test

import (
	"os"
	"testing"

	"github.com/liamvdv/sharedHome/fs"
)

func TestEnrichUnix(t *testing.T) {
	f := fs.File{
		Relpath: "/testdata/sth.txt",
	}
	fp := "./testdata/sth.txt"

	if err := fs.Enrich(fp, &f); err != nil {
		t.Error(err)
	}
	fi, err := os.Stat(fp)
	if err != nil {
		t.Errorf("os.Stat should calculate want, instead failed: %v\n", err)
	}
	if fiMtime := fi.ModTime().UnixNano(); fiMtime != f.MTime {
		t.Errorf("ModTime wrong. want: %d  got: %d\n", fiMtime, f.MTime)
	}
	// 0x1FF = 0001 1111 1111
	// perm = ... r wxrw xrwx
	if fi.Mode() != f.Mode&0x1FF {
		t.Errorf("File mode wrong. want: %o  got: %o\n", fi.Mode(), f.Mode)
	}
	if fi.Size() != f.Size {
		t.Errorf("Size wrong. want: %d  got: %d\n", fi.Size(), f.Size)
	}
	if f.Relpath != "/testdata/sth.txt" {
		t.Errorf("Relpath wrong. Enrich must not change the relpath.")
	}
}

func TestFileName(t *testing.T) {
	f := fs.File{
		Relpath: "/testdata/sth.txt",
	}
	fp := "./testdata/sth.txt"

	fi, err := os.Stat(fp)
	if err != nil {
		t.Errorf("os.Stat should calculate want, instead failed: %v\n", err)
	}
	fiName := fi.Name()
	fName := f.Base()
	if fiName != fName {
		t.Errorf("want: %q  got %q\n", fiName, fName)
	}
}
