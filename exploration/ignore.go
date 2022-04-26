package exploration

import (
	"bufio"
	"path/filepath"
	"strings"

	"github.com/liamvdv/sharedHome/config"
	"github.com/liamvdv/sharedHome/errors"
	"github.com/spf13/afero"
)

// MVP, just use O(N) lookup without support for globs, i.e. direct pattern matching only on the name.

// IgnoreFunc returns true if a string should be skipped and false if should be included.
type IgnoreFunc func(string) bool

func getGlobalIgnoreFunc(patterns []string) IgnoreFunc {
	return func(name string) bool {
		for _, p := range patterns {
			if name == p {
				return true
			}
		}
		return false
	}
}

// getIgnoreFunc returns a function that accepts a name and return whether it
// should be EXCLUDED (true) or INCLUDED (false).
func getIgnoreFunc(fs afero.Fs, dp string, names []string) (IgnoreFunc, error) {
	const op = errors.Op("exploration.getIgnoreFunc")

	there := false
	for _, name := range names {
		if name == config.IgnoreFile {
			there = true
			break
		}
	}

	if !there {
		return func(name string) bool {
			return false
		}, nil
	}

	// exists, read in ignore file
	fp := filepath.Join(dp, config.IgnoreFile)
	file, err := fs.Open(fp)
	if err != nil {
		return nil, errors.E(op, errors.Path(fp), err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var ignoreNames = make([]string, 0, 5) // sensible default

	for scanner.Scan() {
		s := strings.TrimSpace(scanner.Text())
		// empty line or comment
		if strings.HasPrefix(s, comment) || len(s) == 0 {
			continue
		}
		// TODO(liamvdv): test whether scanner includes \r
		if li := len(s) - 1; s[li] == '\r' {
			s = s[:li]
		}
		// TODO(liamvdv): we currently only accept "filename" and "dirname",
		// NOT "/filename" or "/path/filename" or any **regexp** <- should support globs.
		// look at golang.org/pkg/path/filepath/#Match
		ignoreNames = append(ignoreNames, s)
	}

	return func(name string) bool {
		for _, ignore := range ignoreNames {
			if name == ignore {
				return true
			}
		}
		return false
	}, nil
}

const (
	comment = "#"
)
