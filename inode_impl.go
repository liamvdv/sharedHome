// +build !windows, !linux

package main

import "io/fs"

// Does not support anything other than windows and linux
func getInode(fi *fs.FileInfo) uint64 {
	return -1
}