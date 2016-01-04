// +build !windows

package commands

import (
	"fmt"
	"os"
	"path/filepath"
)

func detectShell() (string, error) {
	shell := os.Getenv("SHELL")

	if shell == "" {
		fmt.Printf("The default lines below are for a sh/bash shell, you can specify the shell you're using, with the --shell flag.\n\n")
		return "", ErrUnknownShell
	}

	if os.Getenv("__fish_bin_dir") != "" {
		return "fish", nil
	}

	return filepath.Base(shell), nil
}
