package mcndirs

import (
	"os"
	"path/filepath"

	"github.com/docker/machine/libmachine/mcnutils"
)

var (
	BaseDir = os.Getenv("MACHINE_STORAGE_PATH")
)

func GetBaseDir() string {
	if BaseDir == "" {
		BaseDir = filepath.Join(mcnutils.GetHomeDir(), ".docker", "machine")
	}
	return BaseDir
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
