package main

import (
	"log"
	"os"

	"github.com/liamvdv/sharedHome/config"
	"github.com/liamvdv/sharedHome/osx"
)

func main() {
	env := config.Env{
		Fs:     osx.NewOsFs(),
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	config.InitVars(env.Fs, "")
	defer config.Delete(env.Fs, config.D_TempCacheFolder)

	cfg, err := config.LoadConfigFile(env)
	if err != nil {
		log.Panic(err)
	}

	// start signal module -> should panic so that all cleanup functions can run
	// recover here in main for clean exit.

	if len(os.Args) < 2 {
		log.Println("min 2 args")
	}
	switch os.Args[1] {
	case "sync":
		Sync(env, cfg)
	case "init":
	case "config":
	case "show":
	case "unlock":
	}
}

func Sync(env config.Env, cfg *config.Config) {
	// 1. build local index
	// 1. get remote index -> mutex, ... defer index.Upload()

	// 2. use core packge to look at both indexes and then make the needed changes.
}
