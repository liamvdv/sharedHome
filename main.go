package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/liamvdv/sharedHome/config"
	"github.com/liamvdv/sharedHome/util"
	d "github.com/liamvdv/sharedHome/backend/drive"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func main() {
	config.Init()
	util.PrepareTestConfigInit()

	ctx := context.Background()
	
	b, err := os.ReadFile(d.DriveCredentialsFilepath)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := d.GetClient(config)

	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	// can add own context: List().Context(ctx)...
	r, err := srv.Files.Create(&drive.File{Name: "sharedHome", MimeType: "application/vnd.google-apps.folder"}).Fields("*").Do()
	if err != nil {
		log.Fatalf("Unable to retireve files: %v", err)
	}
	fmt.Printf("%+v\n", *r)

	fmt.Println("Name:", r.Name, "ID:", r.Id, "DriveID", r.DriveId)
	// r, err := srv.Files.List().Q("mimeType = 'application/vnd.google-apps.folder'").Fields("nextPageToken, files(name, id, modifiedTime)").Do()
	// if err != nil {
	// 	log.Fatalf("Unable to retrieve files: %v", err)
	// }
	// for _, f := range r.Files {
	// 	fmt.Printf("%q %q %s\n", f.Name, f.Id, f.ModifiedTime)
	// }
}
