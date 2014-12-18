package drivers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/machine/ssh"
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

func PublicKeyPath() string {
	return filepath.Join(GetHomeDir(), ".docker", "public-key.json")
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

func InstallDocker(host string, port int, user string, sshKey string) error {
	sshCmd := func(args ...string) *exec.Cmd {
		return ssh.GetSSHCommand(host, port, user, sshKey, args...)
	}

	log.Infof("Waiting for SSH...")

	if err := ssh.WaitForTCP(fmt.Sprintf("%s:%d", host, port)); err != nil {
		return err
	}

	cmd := sshCmd("if [ ! -e /usr/bin/docker ]; then curl get.docker.io | sudo sh -; fi")
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = sshCmd("sudo stop docker")
	if err := cmd.Run(); err != nil {
		return err
	}

	log.Debugf("HACK: Downloading version of Docker with identity auth...")

	cmd = sshCmd("sudo curl -sS -o /usr/bin/docker https://ehazlett.s3.amazonaws.com/public/docker/linux/docker-1.4.1-136b351e-identity")
	if err := cmd.Run(); err != nil {
		return err
	}

	log.Debugf("Updating /etc/default/docker to use identity auth...")

	cmd = sshCmd("echo 'export DOCKER_OPTS=\"--auth=identity --host=tcp://0.0.0.0:2376 --host=unix:///var/run/docker.sock --auth-authorized-dir=/root/.docker/authorized-keys.d\"' | sudo tee -a /etc/default/docker")
	if err := cmd.Run(); err != nil {
		return err
	}

	log.Debugf("Adding key to authorized-keys.d...")

	// HACK: temporarily chown to ssh user for providers using non-root accounts
	cmd = sshCmd(fmt.Sprintf("sudo mkdir -p /root/.docker && sudo chown -R %s /root/.docker", user))
	if err := cmd.Run(); err != nil {
		return err
	}

	f, err := os.Open(filepath.Join(os.Getenv("HOME"), ".docker/public-key.json"))
	if err != nil {
		return err
	}
	defer f.Close()

	cmdString := fmt.Sprintf("sudo mkdir -p %q && sudo tee -a %q", "/root/.docker/authorized-keys.d", "/root/.docker/authorized-keys.d/docker-host.json")
	cmd = sshCmd(cmdString)
	cmd.Stdin = f
	if err := cmd.Run(); err != nil {
		return err
	}

	// HACK: change back ownership
	cmd = sshCmd("sudo mkdir -p /root/.docker && sudo chown -R root /root/.docker")
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = sshCmd("sudo start docker")
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
