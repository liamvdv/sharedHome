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

func Run() (e error) {
	// TODO(liamvdv): make config.InitVars() not panic?
	config.InitVars()
	cfg, err := config.LoadConfigFile()
	if err != nil {
		return err
	}
	defer func() {
		err := config.StoreConfigFile(cfg)
		if err != nil {
			log.Printf("Failed to store configuration: %v\n", err)
		}
	}()
	// go look at local index file. list index files remote, check for lock and check names, i. e. sun. -> maybe: download/use local copy -> decrypt and unmarshal
	// else: just unmarshal.
	// defer UploadIndex: just upload if changes are found (store some var that will change later).
	//
	// go explorer to build index file of local state
	// simultaniously fetch index
	filec := make(chan *vfs.File)
	exp := exploration.New()
	idx := index.New()
	go idx.BuildFromChannel(filec)
	go exp.Start()

	return nil
}
