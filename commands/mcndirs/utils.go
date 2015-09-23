package mcndirs

import (
	"os"
	"path/filepath"

	"github.com/docker/machine/libmachine/mcnutils"
)

func GetBaseDir() string {
	baseDir := os.Getenv("MACHINE_STORAGE_PATH")
	if baseDir == "" {
		baseDir = filepath.Join(mcnutils.GetHomeDir(), ".docker", "machine")
	}
	return baseDir
}

func GetDockerDir() string {
	return filepath.Join(mcnutils.GetHomeDir(), ".docker")
}

func GetMachineDir() string {
	return filepath.Join(GetBaseDir(), "machines")
}

func GetMachineCertDir() string {
	return filepath.Join(GetBaseDir(), "certs")
}

func GetMachineCacheDir() string {
	return filepath.Join(GetBaseDir(), "cache")
}
