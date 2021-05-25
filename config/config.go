package config

import (
	_ "embed"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/liamvdv/sharedHome/errors"
	. "github.com/liamvdv/sharedHome/util"
)

/********************** These are the global variables. ***********************/
// Compile time config:
const (
	IgnorefileName = ".notshared"
)

// Runtime config
var TESTING bool = true
var (
	// generally important paths in program
	Home string // Config is stored in user home.
	Wd   string
	Temp string // $TEMP/sharedHome
	Dir  string

	// loaded by user.
	Root string // May differ from user home.
	// backend
	Backend  string
	Username string
	Password string
)

/******************************** End config **********************************/

const (
	dirname   = "sharedHome"
	cfilename = "configuration.json"
)

func Init() {
	const op = errors.Op("config.init")

	if err := initPaths(); err != nil {
		panic(errors.E(op, err))
	}

	if err := initConfig(); err != nil {
		panic(errors.E(op, err).Error())
	}

	if err := initIgnore(); err != nil {
		panic(errors.E(op, err))
	}
}

func initPaths() error {
	const op = errors.Op("config.initPaths")

	var err error
	Home, err = os.UserHomeDir()
	if err != nil {
		return errors.E(op, err)
	}

	Wd, err = os.Getwd()
	if err != nil {
		return errors.E(op, err)
	}

	tempDir := os.TempDir()
	if TESTING {
		Home = Wd
		tempDir = filepath.Join(Wd, "temp")
	}

	Temp = filepath.Join(tempDir, dirname)
	if err := os.MkdirAll(Temp, 0755); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func initConfig() error {
	const op = errors.Op("config.initConfig")

	// find and make right dirpath
	dp, err := os.UserConfigDir()
	if err != nil {
		return errors.E(op, err)
	}

	if TESTING {
		dp = filepath.Join(Wd, "test_")
	}

	Dir = filepath.Join(dp, dirname)

	if err := os.MkdirAll(Dir, 0755); err != nil {
		return errors.E(op, errors.Path(Dir), err)
	}

	// init or create new config
	fp := filepath.Join(Dir, cfilename)
	cfg := config{}
	raw, err := os.ReadFile(fp)
	if err == nil {
		_ = json.Unmarshal(raw, &cfg) // omit error and just continue
	}
	if !validConfig(&cfg) {
		err = promptUser(&cfg, fp)
		if err != nil {
			return errors.E(op, err)
		}
	}

	Root = cfg.Root
	Backend = cfg.Service.Name
	Username = cfg.Service.Username
	Password = cfg.Service.Password
	return nil
}

type serviceConfig struct {
	Name     string `json:"name"` // "drive"
	Username string `json:"username"`
	Password string `json:"password"`
}

type config struct {
	Root    string        `json:"root"`
	Service serviceConfig `json:"service"`
}

// TODO(liamvdv): add service enumeration! Rethink if username and password should be validated
func validConfig(cfg *config) bool {
	return cfg.Service.Name != "" && cfg.Root != "" && Exists(cfg.Root)
}

//go:embed "globalnotshared.txt"
var globalNotShared []byte

func initIgnore() error {
	fp := filepath.Join(Dir, IgnorefileName)
	if Exists(fp) {
		return nil
	}

	if err := os.WriteFile(fp, globalNotShared, 0755); err != nil {
		return err
	}

	return nil
}
