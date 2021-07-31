package vfs_test

import (
	"io/fs"
	"log"
	"path/filepath"
	"testing"
	"time"

	"github.com/liamvdv/sharedHome/osx"
	"github.com/liamvdv/sharedHome/vfs"
)

var testRegistry map[osx.Fs][]string = make(map[osx.Fs][]string)

func testDir(fs osx.Fs) string {
	name, err := osx.TempDir(fs, "", "sharedHome")
	if err != nil {
		log.Panicf("Unable to create temp dir %s", err)
	}
	testRegistry[fs] = append(testRegistry[fs], name)
	return name
}

func removeAllTestFiles(t *testing.T) {
	for fs, list := range testRegistry {
		for _, path := range list {
			if err := fs.RemoveAll(path); err != nil {
				t.Error(fs.Name(), err)
			}
		}
	}
	testRegistry = make(map[osx.Fs][]string)
}

func TestEnrichHostSpecific(t *testing.T) {
	type test struct {
		dir     bool
		name    string
		content []byte
		mtime   time.Time
		perm    fs.FileMode
	}
	var cases = []test{
		{false, "test0.txt", []byte("gimme gimme gimme"), time.Date(2020, 2, 21, 21, 15, 0, 0, time.UTC), 0644},
		{false, "test1.txt", []byte("a man after"), time.Date(2020, 8, 21, 21, 15, 0, 0, time.UTC), 0504},
		{false, "end.md", []byte("midnight"), time.Date(2021, 2, 21, 21, 15, 0, 0, time.UTC), 0700},
		{true, "subdir1.txt", nil, time.Date(2021, 8, 21, 21, 15, 0, 0, time.UTC), 0755},
	}
	fs := osx.NewOsFs()
	dp := testDir(fs)
	defer removeAllTestFiles(t)
	for _, c := range cases {
		fp := filepath.Join(dp, c.name)
		if err := fs.WriteFile(fp, c.content, c.perm); err != nil {
			t.Errorf("cannot create %s because %v", fp, err)
			continue
		}
		if err := fs.Chtimes(fp, time.Now(), c.mtime); err != nil {
			t.Errorf("cannot overwrite mtime because %v", err)
		}
	}

	for _, c := range cases {
		fp := filepath.Join(dp, c.name)
		// fi, err := fs.Stat(fp)
		// if err != nil {
		// 	t.Error(err)
		// }

		f := vfs.File{
			Relpath: filepath.ToSlash(fp[len(dp):]),
		}
		if err := vfs.Enrich(fs, fp, &f); err != nil {
			t.Error(err)
		}

		if f.Mode != c.perm {
			t.Errorf("unequal file modes want: %s got %s", c.perm, f.Mode)
		}
		if f.MTime != c.mtime.UnixNano() {
			t.Errorf("unequla modification times want: %s got %s", c.mtime, time.Unix(0, f.MTime))
		}
		if f.Size != int64(len(c.content)) {
			t.Errorf("size does not match up want %d got %d", len(c.content), f.Size)
		}
		if f.Relpath != "/"+c.name {
			t.Errorf("fail: enrich must not modify repath.")
		}
	}
}
