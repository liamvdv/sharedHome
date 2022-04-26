package vfs

import (
	"os"
	"path"
	"strings"

	"github.com/liamvdv/sharedHome/osx"
)

const sep = string(os.PathSeparator)

// VirtualPath returns the "/" version of the given path. To clean a path use path.Clean() from the stdlib. 
func VirtualPath(relpath string) string {
	return path.Clean(strings.ReplaceAll(relpath, sep, "/"))
}

// returns the local platform specific representation of CleanPath.
func LocalPath(fs osx.Filesystem, root string, relpath string) string {
	return fs.Join(root, strings.ReplaceAll(relpath, "/", sep))
}
