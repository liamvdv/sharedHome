package backend

import (
	"context"
	"io"

	"github.com/liamvdv/sharedHome/vfs"
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

type RemoteFile struct {
	// Relpath is the encrypted file path including the Name as the last element.
	HashRelpath string
	// Name is the encrypted file name
	HashName string
	// Local is the local FileHeader with information about the local file.
	Local *vfs.File 
}

type FileCreator interface {
	CreateFile(ctx context.Context, h RemoteFile, src io.Reader) error
}

type FileReader interface {
	ReadFile(ctx context.Context, h RemoteFile, dst io.Writer) error
}

type FileUpdater interface {
	UpdateFile(ctx context.Context, h RemoteFile, src io.Reader) error
}

type FileRemover interface {
	DeleteFile(ctx context.Context, h RemoteFile) error
}

type FileRenamer interface {
	RenameFile(ctx context.Context, old, new RemoteFile) error
}

type DirCreator interface {
	CreateDir(ctx context.Context, h RemoteFile) error
}

type DirReader interface {
	ReadDir(ctx context.Context, h RemoteFile) (*vfs.File, error)
}

type DirRemover interface {
	DeleteDir(ctx context.Context, h RemoteFile) error
}

type DirRenamer interface {
	RenameDir(ctx context.Context, old, new RemoteFile) error
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