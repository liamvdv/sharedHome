package config

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/liamvdv/sharedHome/errors"
	"github.com/liamvdv/sharedHome/osx"
	"github.com/liamvdv/sharedHome/util"
	"gopkg.in/yaml.v2"
)

// TESTING checks whether the code is run in a test.
// It cannot detect manual tests with executables.
// TODO(liamvdv): remove testing wheels (|| true)
var TESTING bool = (flag.Lookup("test.v") != nil || true)

// NOTE(liamvdv): Config.RootFilepath must not be synchronized, since different plattforms might have other paths.
type Config struct {
	RootFilepath    string   `json:"RootFilepath" yaml:"RootFilepath"`
	UseBackend      string   `json:"UseBackend" yaml:"UseBackend"`
	IgnoreFilenames []string `json:"IgnoreFilenames" yaml:",flow"` // put default ignores here
}

var SupportedBackends = []string{
	"drive",
}

func init() {
	if TESTING {
		SupportedBackends = append(SupportedBackends, "mock")
	}
}

// LoadConfigFile must be called after InitVars. It reads the config file and
// validates it. If it's invalid, it will prompt the user.
func LoadConfigFile(env Env) (*Config, error) {
	config, err := readConfigFile(env.Fs)
	if err != nil {
		if err != uninitializedConfigFile {
			return nil, err
		}
		config = &Config{} // else nil
	}

	errs := validConfigFile(env.Fs, config)
	for ; len(errs) > 0; errs = validConfigFile(env.Fs, config) {
		escape, err := promptConfigFile(env, config, errs)
		if err != nil {
			log.Panic(err)
		}
		if escape {
			return nil, errors.E("Terminating on user request. Could not read valid config file.")
		}
	}
	return config, nil
}

// StoreConfigFile serialises the Config object and persists it to disk.
// It does not check if the object is valid.
func StoreConfigFile(fs osx.Fs, c *Config) error {
	raw, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return fs.WriteFile(ConfigFile, raw, 0606)
}

var uninitializedConfigFile = errors.E("uninitialised config file error")

func readConfigFile(fs osx.Fs) (*Config, error) {
	raw, err := fs.ReadFile(ConfigFile)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return nil, uninitializedConfigFile
	}
	var config Config
	if err := json.Unmarshal(raw, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// validConfigFile returns a nil slice if all parameters are correct.
// Else it will return a slice of messages explaining the problem.
// It may also manipulate c.UseBackend to lowercase, since that is the expected from.
func validConfigFile(fs osx.Fs, c *Config) (errMsg []string) {
	if !util.Exists(fs, c.RootFilepath) {
		msg := fmt.Sprintf("RootFilepath %q does not exist.", c.RootFilepath)
		errMsg = append(errMsg, msg)
	}

	var validBackend bool
	for _, b := range SupportedBackends {
		if strings.EqualFold(c.UseBackend, b) {
			c.UseBackend = b
			validBackend = true
			break
		}
	}
	if !validBackend {
		msg := fmt.Sprintf("UseBackend %q is not supported. Supported are: %s.",
			c.UseBackend, strings.Join(SupportedBackends, ", "))
		errMsg = append(errMsg, msg)
	}

	return errMsg
}

// promptConfigFile asks the user for new Config data. That data is stored
// to the *Config.
func promptConfigFile(env Env, c *Config, errs []string) (escape bool, err error) {
	fmt.Fprintf(env.Stdout,
		`Your configuration file is not valid:
	%s
Would you like to correct it now? (y/n)`, strings.Join(errs, "\n\t"))
	if !ok(env.Stdin) {
		return true, nil
	}
	if len(c.IgnoreFilenames) == 0 {
		c.IgnoreFilenames = globalNotShared
	}

	// Use yaml since it is easier for non-tech people to work with.
	raw, err := yaml.Marshal(c)
	if err != nil {
		return false, err
	}
	tmpPath := filepath.Join(TempCacheFolder, "input.yaml")
	if err := env.Fs.WriteFile(tmpPath, raw, 0600); err != nil {
		return false, err
	}
	defer env.Fs.Remove(tmpPath)

	// If we are testing, we don't want to use openEditor(). To test different
	// versions of the configuration file, a filepath stirng will be written to
	// Stdin. That filepath is a existing file testing the promptConfigFile function.
	if TESTING {
		if _, err := fmt.Fscan(env.Stdin, &tmpPath); err != nil {
			panic(err)
		}
	} else {
		if err := openEditor(env, tmpPath); err != nil {
			return false, err
		}
	}

	raw, err = env.Fs.ReadFile(tmpPath)
	if err != nil {
		return false, err
	}
	return false, yaml.Unmarshal(raw, c)
}

func ok(from io.Reader) bool {
	var answer string
	_, _ = fmt.Fscan(from, &answer)
	for _, r := range [...]string{"y", "yes", "j", "ja"} {
		if strings.EqualFold(r, answer) {
			return true
		}
	}
	return false
}

// openEditor blocks until the user has closed the editor.
func openEditor(env Env, fp string) error {
	var (
		windowsEditor = "notepad.exe"

		linuxEditor = "nano" // not vim, may be hard for some people to grasp
		linuxShell  = []string{"bash", "-c"}
	)

	var editor []string
	switch runtime.GOOS {
	case "windows":
		exepath, err := exec.LookPath(windowsEditor)
		if err != nil {
			return err
		}
		editor = []string{exepath, fp}
	case "linux", "darwin":
		if name := os.Getenv("EDITOR"); name != "" {
			linuxEditor = name
		}
		editor = append(linuxShell, linuxEditor+" "+fp)
	default:
		panic("unsupported platform")
	}

	cmd := exec.Command(editor[0], editor[1:]...)
	cmd.Stdin = env.Stdin
	cmd.Stdout = env.Stdout
	perr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}
	raw, err := io.ReadAll(perr)
	if err != nil {
		return err
	}
	log.Println(string(raw))

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}

var globalNotShared = []string{
	"node_modules",
	"proc",
	"boot",
	"dev",
	"sys",
	"lib",
	"media",
	"mnt",
	"bin",
	"sbin",
	"srv",
	"tmp",
	"log",
}
