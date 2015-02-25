package utils

import (
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

func GetBaseDir() string {
	baseDir := os.Getenv("MACHINE_STORAGE_PATH")
	if baseDir == "" {
		baseDir = filepath.Join(GetHomeDir(), ".docker")
	}
	return baseDir
}

func GetDockerDir() string {
	return filepath.Join(GetHomeDir(), ".docker")
}

func GetMachineRoot() string {
	return filepath.Join(GetBaseDir(), "machine")
}

func GetMachineDir() string {
	return filepath.Join(GetMachineRoot(), "machines")
}

func GetMachineCertDir() string {
	return filepath.Join(GetMachineRoot(), "certs")
}

func GetMachineCacheDir() string {
	return filepath.Join(GetMachineRoot(), "cache")
}

func GetUsername() string {
	u := "unknown"
	osUser := ""

	switch runtime.GOOS {
	case "darwin", "linux":
		osUser = os.Getenv("USER")
	case "windows":
		osUser = os.Getenv("USERNAME")
	}

	if osUser != "" {
		u = osUser
	}

	return u
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
