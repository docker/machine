package drivers

import (
	"fmt"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
)

func PublicKeyPath() string {
	homeDir, err := homedir.Dir()
	if err != nil {
		homeDir = ""
	}
	return filepath.Join(homeDir, ".docker/public-key.json")
}

func AddPublicKeyToAuthorizedHosts(d Driver, authorizedKeysPath string) error {
	f, err := os.Open(PublicKeyPath())
	if err != nil {
		return err
	}
	defer f.Close()

	cmdString := fmt.Sprintf("mkdir -p %q && cat > %q", authorizedKeysPath, filepath.Join(authorizedKeysPath, "docker-host.json"))
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
