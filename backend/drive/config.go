package drive

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/liamvdv/sharedHome/backend"
	"github.com/liamvdv/sharedHome/config"
	. "github.com/liamvdv/sharedHome/util"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

// https://developers.google.com/drive/api/v3/quickstart/go
// https://developers.google.com/workspace/guides/create-credentials
// https://pkg.go.dev/google.golang.org/api/drive/v3#AboutService

var _ backend.Service
var _ = config.Dir



const DriveCredentialsFilename = "drive-credentials.json"

var DriveCredentialsFilepath string

func init() {
	DriveCredentialsFilepath = filepath.Join(config.Dir, DriveCredentialsFilename)
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(cfg *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tok, err := tokenFromFile(DriveCredentialsFilepath)
	if err != nil {
		tok = getTokenFromWeb(cfg)
		saveToken(DriveCredentialsFilepath, tok)
	}
	return cfg.Client(context.Background(), tok)
}

// Retrieves a token from a local file.
func tokenFromFile(fp string) (*oauth2.Token, error) {
	file, err := os.Open(fp)
	if err != nil {
		return nil, err
	}
	defer SaveClose(file)
	tok := &oauth2.Token{}
	err = json.NewDecoder(file).Decode(tok)
	return tok, err
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(cfg *oauth2.Config) *oauth2.Token {
	authURL := cfg.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL) // TODO(liamvdv)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code %v", err) // TODO(liamvdv)
	}

	tok, err := cfg.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err) // TODO(liamvdv)
	}
	return tok
}

// Saves a token to a file path.
func saveToken(fp string, tok *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", fp) // TODO(liamvdv)
	file, err := os.OpenFile(fp, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err) // TODO(liamvdv)
	}
	defer SaveClose(file)
	json.NewEncoder(file).Encode(tok)
}
