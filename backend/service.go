package backend

import (
	"context"
	"io"
)

/*
	All backends are required to implement a package level function of type
	NewBackend and returning a struct adhereing to the service interface.
	The compiler can check that if you include the following expression globally.

	T is your implementation of the Service interface.

	var (
		_ backend.Service = (*T)(nil)
		_ backend.New = New
	)

*/

type New func() (Service, error)

type RemoteFileHeader struct {
	// Relpath is the encrypted file path including the Name as the last element.
	Relpath string
	// Name is the encrypted file name
	Name string
	// Local is the local FileHeader with information about the local file.
	Local *fs.FileHeader
}

type RemoteDirHeader struct {
	// Relpath is the encrypted file path including the Name as the last element.
	Relpath string
	// Name is the encrypted file name
	Name string
	// Local is the local DirHeader with information about the local directory.
	Local *fs.DirHeader
}

type FileCreator interface {
	CreateFile(ctx context.Context, h RemoteFileHeader, src io.Reader) error
}

type FileReader interface {
	ReadFile(ctx context.Context, h RemoteFileHeader, dst io.Writer) error
}

type FileUpdater interface {
	UpdateFile(ctx context.Context, h RemoteFileHeader, src io.Reader) error
}

type FileRemover interface {
	DeleteFile(ctx context.Context, h RemoteFileHeader) error
}

type FileRenamer interface {
	RenameFile(ctx context.Context, oldHeader, newHeader RemoteFileHeader) error
}

type DirCreator interface {
	CreateDir(ctx context.Context, h RemoteDirHeader) error
}

type DirReader interface {
	ReadDir(ctx context.Context, h RemoteDirHeader) (fs.DirHeader, error)
}

type DirRemover interface {
	DeleteDir(ctx context.Context, h RemoteDirHeader) error
}

type DirRenamer interface {
	RenameDir(ctx context.Context, oldHeader, newHeader RemoteDirHeader) error
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

var (
	_ backend.Service = (*Drive)(nil)
	_ backend.New = New
)

type Drive struct{ x }

func (d *Drive) CreateFile(ctx context.Context, h RemoteFileHeader, src io.Reader) error {
	return nil
}

func (d *Drive) ReadFile(ctx context.Context, h RemoteFileHeader, dst io.Writer) error {
	return nil
}

func (d *Drive) UpdateFile(ctx context.Context, h RemoteFileHeader, src io.Reader) error {
	return nil
}

func (d *Drive) DeleteFile(ctx context.Context, h RemoteFileHeader) error {
	return nil
}

func (d *Drive) RenameFile(ctx context.Context, oldHeader, newHeader RemoteFileHeader) error {
	return nil
}

//

func (d *Drive) CreateDir(ctx context.Context, h RemoteDirHeader) error {
	return nil
}

func (d *Drive) ReadDir(ctx context.Context, h RemoteDirHeader) (fs.DirHeader, error) {
	return x, nil
}

func (d *Drive) DeleteDir(ctx context.Context, h RemoteDirHeader) error {
	return nil
}

func (d *Drive) RenameDir(ctx context.Context, oldHeader, newHeader RemoteDirHeader) error {
	return nil
}

func (d *Drive) AddContext(ctx context.Context) {

}

func New() (*Drive, error) {
	return x, nil
}

*/
