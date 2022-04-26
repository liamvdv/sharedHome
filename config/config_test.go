package config_test

import (
	"bytes"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/liamvdv/sharedHome/config"
	"github.com/liamvdv/sharedHome/osx"
	"github.com/liamvdv/sharedHome/testutil"
	"gopkg.in/yaml.v2"
)

func TestLoadAndStoreConfigFile(t *testing.T) {
	defer testutil.RemoveAllTestFiles(t)
	fs := osx.NewMemMapFs()

	testDir := testutil.TestDir(fs)
	configDir := filepath.Join(testDir, ".config")
	var (
		stdin  = &bytes.Buffer{}
		stdout = &bytes.Buffer{}
		stderr = &bytes.Buffer{}
	)
	env := config.Env{
		Fs:     fs,
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	}

	config.InitVars(fs, configDir)
	defer func() {
		err := config.Delete(fs, config.D_ConfigFolder)
		if err != nil {
			t.Error(err)
		}
	}()

	// read promptConfigFile() implementation to understand the following part
	root := filepath.Join(testDir, "home", "liam")
	if err := fs.MkdirAll(root, 0700); err != nil {
		t.Error(err)
	}
	cfg := config.Config{
		RootFilepath:    root,
		UseBackend:      "Mock",
		IgnoreFilenames: []string{"node_modules"},
	}
	path, err := mkValidYamlInputFile(fs, testDir, &cfg)
	if err != nil {
		t.Error(err)
	}

	fmt.Fprintln(stdin, "y")  // Want to change now? y
	fmt.Fprintln(stdin, path) // TESTING special case wants the path.

	c, err := config.LoadConfigFile(env)
	if err != nil {
		t.Error(err)
	}

	if config.StoreConfigFile(fs, c) != nil {
		t.Error(err)
	}

	c, err = config.LoadConfigFile(env)
	if err != nil {
		t.Error(err)
	}
	if config.StoreConfigFile(fs, c) != nil {
		t.Error(err)
	}
}

func mkValidYamlInputFile(fs osx.Fs, testDir string, cfg *config.Config) (path string, err error) {
	tmp, err := fs.CreateTemp(testDir, "inputReplacement*.yaml")
	if err != nil {
		return "", err
	}
	path = tmp.Name()
	raw, err := yaml.Marshal(&cfg)
	if err != nil {
		return "", err
	}
	if err := fs.WriteFile(path, raw, 0600); err != nil {
		return "", err
	}
	return
}
