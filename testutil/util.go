package testutil

import (
	"log"
	"testing"

	"github.com/liamvdv/sharedHome/osx"
)

var testRegistry map[osx.Fs][]string = make(map[osx.Fs][]string)

func TestDir(fs osx.Fs) string {
	name, err := osx.TempDir(fs, "", "sharedHome")
	if err != nil {
		log.Panicf("Unable to create temp dir %s", err)
	}
	testRegistry[fs] = append(testRegistry[fs], name)
	return name
}

func RemoveAllTestFiles(t *testing.T) {
	for fs, list := range testRegistry {
		for _, path := range list {
			if err := fs.RemoveAll(path); err != nil {
				t.Error(fs.Name(), err)
			}
		}
	}
	testRegistry = make(map[osx.Fs][]string)
}
