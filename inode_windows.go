// +build windows

package main

import (
	"io/fs"
	"syscall"
	"log"
)

func getInode(fi *fs.FileInfo) uint64 {
	sys, ok := (*fi).Sys()(*syscall.Win32FileAttributeData)
	if !ok {
		log.Printf("%v not syscall.Stat_t i. e. not windows", *fi)
	}
	return uint64(sys.FileSizeHigh) << 32 | uint64(sys.FileSizeLow)  // turn u32 and u32 to u64 
}


