package vfs

import (
	"bufio"
	"path/filepath"
	"strings"

	"github.com/liamvdv/sharedHome/errors"
	"github.com/liamvdv/sharedHome/osx"
)

// MVP, just use O(N) lookup without support for globs, i.e. direct pattern matching only on the name.

const (
	comment = "#"
	// TODO(liamvdv): change when config is compileable
	IgnoreFile = ".notshared"
)

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
func getIgnoreFunc(fs osx.Fs, dp string, names []string) (IgnoreFunc, error) {
	const op = errors.Op("exploration.getIgnoreFunc")

	var there bool
	for _, name := range names {
		// TODO(liamvdv): change when config is compileable
		// if name == config.IgnoreFile {
		if name == IgnoreFile {
			there = true
			break
		}
	}
	if !there {
		return func(name string) bool {
			return false
		}, nil
	}

	// TODO(liamvdv): change when config is compileable
	// fp := filepath.Join(dp, config.IgnoreFile)
	fp := filepath.Join(dp, IgnoreFile)
	patterns, err := readIgnoreFile(fs, fp)
	if err != nil {
		return nil, errors.E(op, errors.Path(fp), err)
	}
	return func(name string) bool {
		for _, ignore := range patterns {
			if name == ignore {
				return true
			}
		}
		return false
	}, nil
}

func readIgnoreFile(fs osx.Fs, abspath string) (patterns []string, err error) {
	file, err := fs.Open(abspath)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(text, comment) || len(text) == 0 {
			continue
		}
		// TODO(liamvdv): test whether scanner includes \r
		if li := len(text) - 1; text[li] == '\r' {
			text = text[:li]
		}
		// TODO(liamvdv): we currently only accept "filename" and "dirname",
		// NOT "/filename" or "/path/filename" or any **regexp** <- should support globs.
		// look at golang.org/pkg/path/filepath/#Match
		patterns = append(patterns, text)
	}
	return patterns, nil
}
