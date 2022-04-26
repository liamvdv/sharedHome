package config

import (
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	"github.com/liamvdv/sharedHome/errors"
	"github.com/liamvdv/sharedHome/osx"
	"github.com/liamvdv/sharedHome/util"
)

const (
	IgnoreFile = ".notshared"

	// IndexFileTemplate = {sun}.bin
	// sun is the sequential update number
	IndexFileTemplate = "%d.bin"
	// LockIndexFileTemplate = lock-{sun}.bin
	// lock is used to prevent other clients form accessing the index file.
	LockIndexFileTemplate = "lock-%d.bin"
)

var (
	// ConfigFolder = CONFIG_DIR/sharedHome
	ConfigFolder string
	// BackendConfigFolder = CONFIG_DIR/sharedHome/backend
	// Stores the configuration files for each backend service
	BackendConfigFolder string

	// IndexCacheFolder = CONFIG_DIR/sharedHome/index
	// Stores zero or more Cache files identified by their sequential update number.
	IndexCacheFolder string

	// TempCacheFolder = CONFIG_DIR/sharedHome/temp
	TempCacheFolder string

	// ConfigFile = CONFIG_DIR/sharedHome/configuration.json
	ConfigFile string

	// LogFolder = CONFIG_DIR/sharedHome/log
	LogFolder string
)

// InitVars ensures that all named paths and folders exist, else it panics.
// Pass "" to configDirpath to use the user config directory.
func InitVars(fs osx.Fs, configDirpath string) {
	if configDirpath == "" {
		var err error
		configDirpath, err = userConfigDir()
		if err != nil {
			log.Panic(err)
		}
	}

	ConfigFolder = filepath.Join(configDirpath, "sharedHome")
	if err := existOrCreate(fs, ConfigFolder, true); err != nil {
		log.Panic(err)
	}
	ConfigFile = filepath.Join(ConfigFolder, "configuration.json")
	if err := existOrCreate(fs, ConfigFile, false); err != nil {
		log.Panic(err)
	}
	BackendConfigFolder = filepath.Join(ConfigFolder, "backend")
	if err := existOrCreate(fs, BackendConfigFolder, true); err != nil {
		log.Panic(err)
	}
	IndexCacheFolder = filepath.Join(ConfigFolder, "index")
	if err := existOrCreate(fs, IndexCacheFolder, true); err != nil {
		log.Panic(err)
	}
	TempCacheFolder = filepath.Join(ConfigFolder, "temp")
	if err := existOrCreate(fs, TempCacheFolder, true); err != nil {
		log.Panic(err)
	}
	LogFolder = filepath.Join(ConfigFolder, "log")
	if err := existOrCreate(fs, LogFolder, true); err != nil {
		log.Panic(err)
	}
}

// userConfigDir is a drop in replacement for os.UserConfigDir that takes care of
// people to call this as sudo. It works for windows, darwin, ios and unix.
func userConfigDir() (string, error) {
	if runtime.GOOS == "windows" {
		dp := os.Getenv("AppData")
		if dp == "" {
			return "", errors.E("%AppData% is not defined")
		}
		return dp, nil
	}

	home := func() string {
		uname := os.Getenv("SUDO_USER")
		if uname != "" {
			u, err := user.Lookup(uname)
			if err == nil {
				return u.HomeDir
			}
		}
		return os.Getenv("HOME")
	}()

	if runtime.GOOS == "darwin" || runtime.GOOS == "ios" {
		return home + "/Library/Application Support", nil
	}

	// unix
	if dp := os.Getenv("XDG_CONFIG_HOME"); dp != "" {
		return dp, nil
	}
	return home + "/.config", nil
}

func existOrCreate(fs osx.Fs, fp string, isDir bool) error {
	if util.Exists(fs, fp) {
		return nil
	}
	if isDir {
		return fs.MkdirAll(fp, 0700)
	}

	dp := filepath.Dir(fp)
	if err := fs.MkdirAll(dp, 0700); err != nil {
		return err
	}
	f, err := fs.Create(fp)
	if err != nil {
		return err
	}
	return f.Close()
}

type deleteTargets int

const (
	// do not reorder
	D_TempCacheFolder deleteTargets = iota
	D_ConfigFolder
	D_ConfigFile
	D_BackendConfigFolder
	D_IndexCacheFolder
	D_LogFolder
	nTargets
)

var deleteTargetMapping = []*string{
	// do not reorder
	D_TempCacheFolder:     &TempCacheFolder,
	D_ConfigFolder:        &ConfigFile,
	D_ConfigFile:          &ConfigFile,
	D_BackendConfigFolder: &BackendConfigFolder,
	D_IndexCacheFolder:    &IndexCacheFolder,
	D_LogFolder:           &LogFolder,
}

// TODO(liamvdv): Or just rewrite as general Delete(targets ...string) error,
// but then inputs are not validated and os.RemoveAll is dangerous....

// Delete deletes temporary directories including their content.
func Delete(fs osx.Fs, targets ...deleteTargets) error {
	for _, target := range targets {
		if !(0 <= target && target < nTargets) {
			return errors.E("target does not exist")
		}
		fp := *deleteTargetMapping[target]

		err := fs.RemoveAll(fp)
		if err != nil {
			return err
		}
	}
	return nil
}
