package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/liamvdv/sharedHome/config"
)

// index file stores a representation of the filesystem
const FILENAME = "index"





// ~/LOG_FILENAME
const LOG_FILENAME = ".sharedHome.log"

func errorPersister(errc <-chan error) {
	fp := filepath.Join(config.Dir, LOG_FILENAME)
	file, err := os.OpenFile(fp, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
	if err != nil {
		panic(err)
	}
	// log to file
	log.SetOutput(file)

	// consume errors and write to log.
	for err := range errc {
		fmt.Println(err)
		log.Println(err.Error())
	}
}