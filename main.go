package main

import (
	"log"

	"github.com/liamvdv/sharedHome/config"
	"github.com/liamvdv/sharedHome/exploration"
	"github.com/liamvdv/sharedHome/index"
)

func main() {
	// util.PrepareTestConfigInit()
	// config.Init()

	// drive.TestRemoteIdList()
	// can add own context: List().Context(ctx)...
	// r, err := srv.Files.Create(&drive.File{Name: "sharedHome", MimeType: "application/vnd.google-apps.folder"}).Fields("*").Do()
	// if err != nil {
	// 	log.Fatalf("Unable to retireve files: %v", err)
	// }
	// fmt.Printf("%+v\n", *r)

	// fmt.Println("Name:", r.Name, "ID:", r.Id, "DriveID", r.DriveId)
	// r, err := srv.Files.List().Q("mimeType = 'application/vnd.google-apps.folder'").Fields("nextPageToken, files(name, id, modifiedTime)").Do()
	// if err != nil {
	// 	log.Fatalf("Unable to retrieve files: %v", err)
	// }
	// for _, f := range r.Files {
	// 	fmt.Printf("%q %q %s\n", f.Name, f.Id, f.ModifiedTime)
	// }
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
