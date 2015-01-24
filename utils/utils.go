package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

func GetHomeDir() string {
	if runtime.GOOS == "windows" {
		return os.Getenv("USERPROFILE")
	}
	return os.Getenv("HOME")
}

func GetDockerDir() string {
	return fmt.Sprintf(filepath.Join(GetHomeDir(), ".docker"))
}

func GetMachineDir() string {
	return fmt.Sprintf(filepath.Join(GetDockerDir(), "machines"))
}

func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}

	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	if _, err = io.Copy(out, in); err != nil {
		return err
	}

	return nil
}
