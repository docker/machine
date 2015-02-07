package drivers

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/docker/machine/utils"
)

func PublicKeyPath() string {
	return filepath.Join(utils.GetDockerDir(), "public-key.json")
}

func AddPublicKeyToAuthorizedHosts(d Driver, authorizedKeysPath string) error {
	f, err := os.Open(PublicKeyPath())
	if err != nil {
		return err
	}
	defer f.Close()

	// Use path.Join here, want to create unix path even when running on Windows.
	cmdString := fmt.Sprintf("mkdir -p %q && cat > %q", authorizedKeysPath,
		path.Join(authorizedKeysPath, "docker-host.json"))
	cmd, err := d.GetSSHCommand(cmdString)
	if err != nil {
		return err
	}
	cmd.Stdin = f
	return cmd.Run()
}

func PublicKeyExists() (bool, error) {
	_, err := os.Stat(PublicKeyPath())
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
