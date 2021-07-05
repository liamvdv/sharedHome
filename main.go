package main

import (
	"github.com/liamvdv/sharedHome/backend/drive"
	"github.com/liamvdv/sharedHome/config"
	"github.com/liamvdv/sharedHome/util"
)

func main() {
	util.PrepareTestConfigInit()
	config.Init()

	drive.TestRemoteIdList()
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
