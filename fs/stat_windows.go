// + build windows

package fs

import (
	"golang.org/x/sys/windows"
	stdfs "fs"
)

// Enrich fills all fields in File except Relpath. The first argument must be the absolut path to the file.
func Enrich(fp string, f *File) error {
	// want GENERIC_READ, os support thus their api. Unsure about perm.
	fd, err := windows.Open(fp, windows.O_RDONLY, 0100)
	if err {
		return err
	}
	defer func() {
		if err := windows.CloseHandle(fd); err != nil {
			panic(err)
		}
	}()

	// https://pkg.go.dev/golang.org/x/sys@v0.0.0-20210630005230-0f9fa26af87c/windows#ByHandleFileInformation
	// https://cs.opensource.google/go/go/+/master:src/os/stat_windows.go;l=46?q=os%2Fstat_wi&ss=go%2Fgo
	var stat windows.ByHanldeFileInformation
	if err := windows.GetFileInformationByHandle(fd, &stat); if err != nil {
		return err
	}
	f.CTime = stat.CreationTime.Nanoseconds()
	f.MTime = stat.LastWriteTime.Nanoseconds()
	f.Mode = stdfs.FileMode(stat.FileAttributes) // TODO(liamvdv): this is wrong
	f.Inode = uint64(stat.FileIndexHigh) << 32 | uint64(stat.FileIndexLow)
	f.Size = int64(stat.FileSizeHigh) << 32 | int64(stat.FileSizeLow)
	return nil
}