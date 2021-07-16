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
	"github.com/liamvdv/sharedHome/util"
	"gopkg.in/yaml.v2"
)

// TESTING checks whether the code is run in a test. It cannot detect manual tests with executables.
var TESTING bool = (flag.Lookup("test.v") != nil || // only checks if registered, true/false doesn't matter
	strings.HasSuffix(os.Args[0], ".test") ||
	strings.Contains(os.Args[0], "/_test/") ||
	true) // TODO(liamvdv): remove testing wheels

// NOTE(liamvdv): configFile.RootFilepath must not be syncronised, since different plattforms might have other paths.
type configFile struct {
	RootFilepath    string   `json:"RootFilepath" yaml:"RootFilepath"`
	UseBackend      string   `json:"UseBackend" yaml:"UseBackend"`
	IgnoreFilenames []string `json:"IgnoreFilenames" yaml:",flow"` // put default ignores here
}

var SupportedBackends = func() []string {
	slice := []string{
		// add more backends here
		"drive",
	}
	if TESTING {
		slice = append(slice, "mock")
	}
	return slice
}()

// LoadConfigFile must be called after InitVars.
// It reads the config file and validates it. If it's invalid, it will prompt the user.
func LoadConfigFile() (*configFile, error) {
	var errs []string
	config, err := readConfigFile()
	if err != nil {
		if err != uninitializedConfigFile {
			return nil, err
		}
	} else {
		errs = validConfigFile(config)
		if len(errs) == 0 {
			return config, nil
		}
	}
	for {
		if TESTING {
			// go/src/github.com/liamvdv/sharedHome/testdata
			config = &configFile{
				RootFilepath:    filepath.Join(filepath.Dir(filepath.Dir(ConfigFolder)), "testdata"),
				UseBackend:      "mock",
				IgnoreFilenames: globalNotShared,
			}
		} else {
			escape, err := promptConfigFile(config, errs)
			if err != nil {
				log.Panic(err)
			}
			if escape {
				return nil, errors.E("Terminating on user request. Could not read valid config file.")
			}
		}

		errs = validConfigFile(config)
		if len(errs) == 0 {
			break
		}
	}
	return config, nil
}

// StoreConfigFile serialises the configFile object passed and persists it to disk.
// It does not check if the object is valid.
func StoreConfigFile(c *configFile) error {
	raw, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(ConfigFile, raw, 0606)
}

var uninitializedConfigFile = errors.E("uninitialisedConfigFile error")

func readConfigFile() (*configFile, error) {
	raw, err := os.ReadFile(ConfigFile)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return nil, uninitializedConfigFile
	}
	var config configFile
	if err := json.Unmarshal(raw, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// validConfigFile returns a nil slice if all parameters are correct. Else it will return a slice of messages explaining the problem.
// It may also manipulate c.UseBackend to lowercase, since that is expected behaviour from the other modules.
func validConfigFile(c *configFile) (errMsg []string) {
	if !util.Exists(c.RootFilepath) {
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
		msg := fmt.Sprintf("UseBackend %q is not supported. Supported are: %s.", c.UseBackend, strings.Join(SupportedBackends, ", "))
		errMsg = append(errMsg, msg)
	}

	// TODO(liamvdv): is there a reason to validate global ignores? If so, how should we do that?

	return errMsg
}

// TODO(liamvdv): needs (manual) testing
func promptConfigFile(c *configFile, errs []string) (escape bool, err error) {
	fmt.Printf(
		`Your configuration file is not valid:
	%s
Would you like to correct it now? (y/n)`, strings.Join(errs, "\n\t"))
	if !ok(os.Stdin) {
		return true, nil
	}
	if len(c.IgnoreFilenames) == 0 {
		c.IgnoreFilenames = globalNotShared
	}

	// Use yaml since it is easier for non-tech people to work with
	raw, err := yaml.Marshal(c)
	if err != nil {
		return false, err
	}
	tmpPath := filepath.Join(TempCacheFolder, "input.yaml")
	if err := os.WriteFile(tmpPath, raw, 0606); err != nil {
		return false, err
	}
	defer os.Remove(tmpPath)
	if err := openEditor(tmpPath); err != nil {
		return false, err
	}

	raw, err = os.ReadFile(tmpPath)
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
func openEditor(fp string) error {
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
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
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
