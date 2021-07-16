package config

import (
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/liamvdv/sharedHome/util"
)

const (
	IgnoreFile = ".notshared"
)

var (
	// ConfigFolder = CONFIG_DIR/sharedHome
	ConfigFolder string
	// BackendConfigFolder = CONFIG_DIR/sharedHome/backend
	// Stores the configuration files for each backend service
	BackendConfigFolder string

	// IndexCacheFolder = CONFIG_DIR/sharedHome/index
	IndexCacheFolder string

	// TempCacheFolder = CONFIG_DIR/sharedHome/temp
	TempCacheFolder string

	// ConfigFile = CONFIG_DIR/sharedHome/configuration.json
	ConfigFile string
)

// InitVars ensures that all named paths and folders exist. It does not check the content, i. e.
// configurations files may be invalid.
// When the TESTING is true, the config base directory is $GOPATH/src/github.com/liamvdv/sharedHome/_fakeConfigDir
func InitVars() {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		log.Panic(err)
	}
	if TESTING {
		userConfigDir, err = os.Getwd()
		if err != nil {
			log.Panic(err)
		}
		for ; filepath.Base(userConfigDir) != "sharedHome"; userConfigDir = filepath.Dir(userConfigDir) {
		}
		userConfigDir = filepath.Join(userConfigDir, "_fakeConfigDir")
	}

	ConfigFolder = filepath.Join(userConfigDir, "sharedHome")
	if err := existOrCreate(ConfigFolder, true); err != nil {
		log.Panic(err)
	}
	ConfigFile = filepath.Join(ConfigFolder, "configuration.json")
	if err := existOrCreate(ConfigFile, false); err != nil {
		log.Panic(err)
	}
	BackendConfigFolder = filepath.Join(ConfigFolder, "backend")
	if err := existOrCreate(BackendConfigFolder, true); err != nil {
		log.Panic(err)
	}
	IndexCacheFolder = filepath.Join(ConfigFolder, "index")
	if err := existOrCreate(IndexCacheFolder, true); err != nil {
		log.Panic(err)
	}
	TempCacheFolder = filepath.Join(ConfigFolder, "temp")
	if err := existOrCreate(TempCacheFolder, true); err != nil {
		log.Panic(err)
	}
}

type deleteTargets int

const (
	// do not reorder
	D_TempCacheFolder deleteTargets = iota
	D_ConfigFolder
	D_ConfigFile
	D_BackendConfigFolder
	D_IndexCacheFolder
	nTargets
)

var deleteTargetMapping = []*string{
	// do not reorder
	D_TempCacheFolder:     &TempCacheFolder,
	D_ConfigFolder:        &ConfigFile,
	D_ConfigFile:          &ConfigFile,
	D_BackendConfigFolder: &BackendConfigFolder,
	D_IndexCacheFolder:    &IndexCacheFolder,
}

// TODO(liamvdv): Use in main with
//	config.InitVars()
//	defer config.Delete(config.D_TempCacheFolder)
// Or just rewrite as general Delete(targets ...string) error, but then inputs are not validated and os.RemoveAll is dangerous....

// Delete deletes temporary directories including their content.
func Delete(targets ...deleteTargets) error {
	for _, target := range targets {
		if !(0 <= target && target < nTargets) {
			return errors.New("target does not exist")
		}
		fp := *deleteTargetMapping[target]

		err := os.RemoveAll(fp)
		if err != nil {
			return err
		}
	}
	return nil
}

func existOrCreate(fp string, dir bool) error {
	if util.Exists(fp) {
		return nil
	}
	if dir {
		return os.MkdirAll(fp, 0700)
	}

	dp := filepath.Dir(fp)
	if err := os.MkdirAll(dp, 0700); err != nil {
		return err
	}
	f, err := os.Create(fp)
	if err != nil {
		return err
	}
	return f.Close()
}

// TODO(liamvdv): Implement function for retrieving the backend configuration file that is service dependent.
