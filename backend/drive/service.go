package drive

import (
	"io"
	"log"
	"os"

	"github.com/liamvdv/sharedHome/backend"
	"github.com/liamvdv/sharedHome/errors"
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
	srv    *drive.Service
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
		log.Panicf("could not initGlobals: %v\n", err) // TODO(liamvdv): remove
		return nil, err
	}

	return &Drive{
		srv: srv,
	}, nil
}

func (d *Drive) CreateFile(ctx context.Context, h backend.RemoteFileHeader, src io.Reader) error {
	parentId := "TODO" // TODO(liamvdv): proper parent

	return createFile(ctx, d.srv, parentId, h.Name, h.Local.ModTime, src)
}

func (d *Drive) ReadFile(ctx context.Context, h backend.RemoteFileHeader, dst io.Writer) error {
	return nil
}

func (d *Drive) UpdateFile(ctx context.Context, h backend.RemoteFileHeader, src io.Reader) error {
	return nil
}

func (d *Drive) DeleteFile(ctx context.Context, h backend.RemoteFileHeader) error {
	return nil
}

func (d *Drive) RenameFile(ctx context.Context, oldHeader, newHeader backend.RemoteFileHeader) error {
	return nil
}

//
func (d *Drive) CreateDir(ctx context.Context, h backend.RemoteDirHeader) error {
	return nil
}

func (d *Drive) ReadDir(ctx context.Context, h backend.RemoteDirHeader) (fs.DirHeader, error) {
	return fs.DirHeader{}, nil
}

func (d *Drive) DeleteDir(ctx context.Context, h backend.RemoteDirHeader) error {
	return nil
}

func (d *Drive) RenameDir(ctx context.Context, oldHeader, newHeader backend.RemoteDirHeader) error {
	return nil
}

// useless?
func (d *Drive) AddContext(ctx context.Context) {

}
