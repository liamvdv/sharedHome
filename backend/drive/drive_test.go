package drive

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

	"github.com/liamvdv/sharedHome/config"
	"github.com/liamvdv/sharedHome/util"
)

// https://github.com/codyoss/retry

func TestDriveOnline(t *testing.T) {
	util.PrepareTestConfigInit()
	config.Init()

	ctx := context.Background()
	fmt.Println(DriveCredentialsFilepath, DriveTokenFilepath)
	b, err := os.ReadFile(DriveCredentialsFilepath)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := GetClient(config)

	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	// can add own context: List().Context(ctx)...
	r, err := srv.Files.List().PageSize(10).Fields("nextPageToken, files(id, name)").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve files: %v", err)
	}
	fmt.Println("Files:")
	if len(r.Files) == 0 {
		fmt.Println("No files found.")
	} else {
		for _, i := range r.Files {
			fmt.Printf("%s (%s)\n", i.Name, i.Id)
		}
	}
}
