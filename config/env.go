package config

import (
	"io"
	"os"

	"github.com/liamvdv/sharedHome/osx"
)

// The Env struct makes the cli testable.
type Env struct {
	Fs     osx.Fs
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
	// ExitCode int ?
}

func NewOsEnv() *Env {
	return &Env{
		Fs:     osx.NewOsFs(),
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
}
