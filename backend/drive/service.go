package drive

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/liamvdv/sharedHome/backend"
	"github.com/liamvdv/sharedHome/errors"
	"github.com/liamvdv/sharedHome/fs"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

var (
	_ backend.Service = (*Drive)(nil)
)

// Drive must implement Service.
type Drive struct {
	srv *drive.Service
	parIDs map[string]string
}

func New() (*Drive, error) {
	const op = errors.Op("backend.drive.New")

	raw, err := os.ReadFile(DriveCredentialsFilepath)
	if err != nil {
		return nil, errors.E(op, err)
	}
	// drive.DriveScope allows full access to google drive of user.
	cfg, err := google.ConfigFromJSON(raw, drive.DriveScope)
	if err != nil {
		return nil, errors.E(op, err)
	}
	client := GetClient(cfg)

	ctx := context.Background()
	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	if err := initGlobals(srv); err != nil {
		return nil, err
	}

	return &Drive{
		srv: srv,
	}, nil
}

func (d *Drive) CreateFile(ctx context.Context, header fs.FileHeader, src io.Reader) error {
	dp := filepath.Dir(header.Name)
	parentID := getParent(filepath.Dir(header.Name))
	name := header.Name[]
	_, err := createFile(ctx, d.srv, parentID, )
}

func (d *Drive) ReadFile(ctx context.Context, header fs.FileHeader, dst io.Writer) error {
	return nil
}

func (d *Drive) UpdateFile(ctx context.Context, header fs.FileHeader, src io.Reader) error {
	return nil
}

func (d *Drive) DeleteFile(ctx context.Context, header fs.FileHeader) error {
	return nil
}

func (d *Drive) RenameFile(ctx context.Context, oldHeader, newHeader fs.FileHeader) error {
	return nil
}

//
func (d *Drive) CreateDir(ctx context.Context, header fs.DirHeader) error {
	return nil
}

func (d *Drive) ReadDir(ctx context.Context, header fs.DirHeader) (fs.DirHeader, error) {
	return fs.DirHeader{}, nil
}

func (d *Drive) DeleteDir(ctx context.Context, header fs.DirHeader) error {
	return nil
}

func (d *Drive) RenameDir(ctx context.Context, oldHeader, newHeader fs.DirHeader) error {
	return nil
}

// useless?
func (d *Drive) AddContext(ctx context.Context) {

}
