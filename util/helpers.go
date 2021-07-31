package util

import (
	"errors"
	"os"

	"github.com/liamvdv/sharedHome/osx"
)

// TODO(liamvdv): super difficult to get 100% right.

func Exists(fs osx.Fs, path string) bool {
	_, err := fs.Stat(path)
	return err == nil || !errors.Is(err, os.ErrNotExist)
}

func SaveClose(f osx.File) {
	if err := f.Close(); err != nil {
		panic(err)
	}
}

// PrepareTestConfigInit makes the working directory the --/sharedHome project root.
// func PrepareTestConfigInit() {
// 	fp, err := os.Getwd()
// 	if err != nil {
// 		panic(err)
// 	}
// 	for {
// 		name := filepath.Base(fp)
// 		if name == "sharedHome" { // src root dir
// 			break
// 		}
// 		if len(fp) < 3 { // no valid known path.
// 			panic("cannot find sharedHome root directory.")
// 		}

// 		fp = fp[:len(fp)-1-len(name)] // -1 for '/' or '\'
// 	}

// 	if err := os.Chdir(fp); err != nil {
// 		panic(err)
// 	}
// }
