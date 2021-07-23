// +build linux darwin unix

package fs

import (
	stdfs "io/fs"
	"syscall"

	"golang.org/x/sys/unix"
)

// https://cs.opensource.google/go/go/+/master:src/os/stat_linux.go;drc=master

// Enrich fills all fields in File except Relpath. The first argument must be the absolut path to the file.
func Enrich(fp string, f *File) error {
	// https://pkg.go.dev/golang.org/x/sys@v0.0.0-20210630005230-0f9fa26af87c/unix#Stat_t
	var stat unix.Stat_t
	if err := unix.Stat(fp, &stat); err != nil {
		return err
	}
	f.CTime = timespecToUnixNano(stat.Ctim)
	f.MTime = timespecToUnixNano(stat.Mtim)
	f.Mode = unixModeToFileMode(stat.Mode)
	f.Inode = stat.Ino
	f.Size = stat.Size
	return nil
}

func timespecToUnixNano(t unix.Timespec) int64 {
	// according to https://cs.opensource.google/go/go/+/refs/tags/go1.16.6:src/time/time.go;l=1137
	// adjusted for https://pkg.go.dev/golang.org/x/sys@v0.0.0-20210630005230-0f9fa26af87c/unix#Timespec
	s, ns := t.Unix()
	return s*1e9 + ns
}

// implementation from the the stdlib, adjusted for the use case
// https://cs.opensource.google/go/go/+/master:src/os/stat_linux.go;drc=master;l=12
func unixModeToFileMode(m uint32) stdfs.FileMode {
	mode := stdfs.FileMode(m & 0777)
	switch m & unix.S_IFMT {
	case unix.S_IFBLK:
		mode |= stdfs.ModeDevice
	case unix.S_IFCHR:
		mode |= stdfs.ModeDevice | stdfs.ModeCharDevice
	case unix.S_IFDIR:
		mode |= stdfs.ModeDir
	case unix.S_IFIFO:
		mode |= stdfs.ModeNamedPipe
	case syscall.S_IFLNK:
		mode |= stdfs.ModeSymlink
	case unix.S_IFREG:
		// nothing to do
	case unix.S_IFSOCK:
		mode |= stdfs.ModeSocket
	}
	if m&unix.S_ISGID != 0 {
		mode |= stdfs.ModeSetgid
	}
	if m&unix.S_ISUID != 0 {
		mode |= stdfs.ModeSetuid
	}
	if m&unix.S_ISVTX != 0 {
		mode |= stdfs.ModeSticky
	}
	return mode
}
