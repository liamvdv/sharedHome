// +build linux

package main

import (
	"io/fs"
	"syscall"
	"log"
)

func getInode(fi *fs.FileInfo) uint64 {
	sys, ok := (*fi).Sys().(*syscall.Stat_t)
	if !ok {
		log.Printf("%v not syscall.Stat_t i. e. not linux", *fi)
	}
	return sys.Ino
}