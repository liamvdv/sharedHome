package config

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/liamvdv/sharedHome/errors"
	. "github.com/liamvdv/sharedHome/util"
)

var (
	windowsEditor = "notepad.exe"

	linuxEditor = "nano" // not vim, may be hard for some people to grasp
	linuxShell  = []string{"bash", "-c"}
)

const todoMsg = "!!TODO!!"

// promptUser does not set the global variables, but ensures correct values.
func promptUser(cfg *config, fp string) error {
	const op = errors.Op("config.promptUser")

	var editor []string
	switch runtime.GOOS {
	case "windows":
		exepath, err := exec.LookPath(windowsEditor)
		if err != nil {
			return errors.E(op, err)
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

	// reset to default
	cfg.Root = todoMsg
	cfg.Service.Name = "drive" // TODO(liamvdv): not future proof
	if err := reset(fp, cfg); err != nil {
		return err
	}

	fmt.Printf("Please replace the %q fields.\n", todoMsg)
	for !validConfig(cfg) {
		time.Sleep(4 * time.Second)
		// collect input
		cmd := exec.Command(editor[0], editor[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		perr, err := cmd.StderrPipe()
		if err != nil {
			return errors.E(op, err)
		}

		if err := cmd.Start(); err != nil {
			return errors.E(op, err)
		}
		raw, err := io.ReadAll(perr)
		if err != nil {
			return errors.E(op, err)
		}

		log.Println("Editor stderr:", string(raw))

		if err := cmd.Wait(); err != nil {
			return errors.E(op, err)
		}

		// give feedback
		raw, err = os.ReadFile(fp)
		if err != nil {
			return errors.E(op, err)
		}
		if err := json.Unmarshal(raw, cfg); err != nil {
			fmt.Printf("Invalid json. Do only fill out the %q fields.\n", todoMsg)
			if quit() {
				return errors.E(op, "User quitted, unable to obtain information.")
			}
			if err := reset(fp, cfg); err != nil {
				return errors.E(op, err)
			}
			continue
		}
		if !Exists(cfg.Root) {
			fmt.Printf("The path: %q does not exist.\n", cfg.Root)
		}
	}
	return nil
}

func read(fp string, cfg *config) error {
	raw, err := os.ReadFile(fp)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, cfg); err != nil {
		return err
	}
	return nil
}

func reset(fp string, cfg *config) error {
	raw, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(fp, raw, 0755); err != nil {
		return err
	}
	return nil
}

func quit() bool {
	var answer string
	fmt.Println("Do you wan to quit? (y/n)")
	_, err := fmt.Scan(&answer)
	if err != nil {
		log.Println(err)
		return true
	}
	ys := []string{"y", "Y", "yes", "Yes", "YES"}
	ns := []string{"n", "N", "no", "No", "NO"}

	if in(answer, ys) {
		return true
	}
	if in(answer, ns) {
		return false
	}
	return quit()
}

func in(item string, iterable []string) bool {
	for _, i := range iterable {
		if item == i {
			return true
		}
	}
	return false
}
