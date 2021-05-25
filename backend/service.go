package backend

import (
	"context"
	"io"

	"github.com/liamvdv/sharedHome/fs"
)

type FileCreator interface {
	CreateFile(ctx context.Context, header fs.FileHeader, src io.Reader) error
}

type FileReader interface {
	ReadFile(ctx context.Context, header fs.FileHeader, dst io.Writer) error
}

type FileUpdater interface {
	UpdateFile(ctx context.Context, header fs.FileHeader, src io.Reader) error
}

type FileRemover interface {
	DeleteFile(ctx context.Context, header fs.FileHeader) error
}

type FileRenamer interface {
	RenameFile(ctx context.Context, oldHeader, newHeader fs.FileHeader) error
}

//

type DirCreator interface {
	CreateDir(ctx context.Context, header fs.DirHeader) error
}

type DirReader interface {
	ReadDir(ctx context.Context, header fs.DirHeader) (fs.DirHeader, error)
}

type DirRemover interface {
	DeleteDir(ctx context.Context, header fs.DirHeader) error
}

type DirRenamer interface {
	RenameDir(ctx context.Context, oldHeader, newHeader fs.DirHeader) error
}

type Service interface {
	FileReader
	FileCreator
	FileUpdater
	FileRemover
	FileRenamer

	DirReader
	DirCreator
	DirRemover
	DirRenamer

	AddContext(ctx context.Context) 
}


/* Service Template

// validate interface
var _ backend.Service = (*Drive)(nil)

type Drive struct{}

func (d *Drive) CreateFile(ctx context.Context, header fs.FileHeader, src io.Reader) error {
	return nil
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
	return x, nil
}
 
func (d *Drive) DeleteDir(ctx context.Context, header fs.DirHeader) error {
	return nil
}

func (d *Drive) RenameDir(ctx context.Context, oldHeader, newHeader fs.DirHeader) error {
	return nil
}



*/