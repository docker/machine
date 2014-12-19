package drivers

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/machine/ssh"
	"time"
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

func InstallDocker(driver Driver) error {
	log.Infof("Waiting for SSH...")

	ip, err := driver.GetIP()
	if err != nil {
		return err
	}
	if err := ssh.WaitForTCP(fmt.Sprintf("%s:%d", ip, driver.GetSSHPort())); err != nil {
		return err
	}

	log.Infof("Installing and configuring Docker...")
	installCmd := "curl -s get.docker.io | sudo sh -"

	cmd, err := driver.GetSSHCommand(fmt.Sprintf("if [ ! -e /usr/bin/docker ]; then %s; fi", installCmd))
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	// add host entry for local name to workaround some providers not adding
	cmd, err = driver.GetSSHCommand("echo \"127.0.1.2 $(hostname -s)\" | sudo tee -a /etc/hosts > /dev/null 2>&1")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd, err = driver.GetSSHCommand("sudo stop docker")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	log.Debugf("HACK: Downloading version of Docker with identity auth...")

	cmd, err = driver.GetSSHCommand("sudo curl -sS -o /usr/bin/docker https://ehazlett.s3.amazonaws.com/public/docker/linux/docker-1.4.1-136b351e-identity")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	log.Debugf("Updating /etc/default/docker to use identity auth...")

	cmd, err = driver.GetSSHCommand("echo 'export DOCKER_OPTS=\"--auth=identity --host=tcp://0.0.0.0:2376 --host=unix:///var/run/docker.sock --auth-authorized-dir=/root/.docker/authorized-keys.d\"' | sudo tee -a /etc/default/docker")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	log.Debugf("Adding key to authorized-keys.d...")

	// HACK: temporarily chown to ssh user for providers using non-root accounts
	cmd, err = driver.GetSSHCommand(fmt.Sprintf("sudo mkdir -p /root/.docker && sudo chown -R %s /root/.docker", driver.GetSSHUser()))
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	f, err := os.Open(filepath.Join(os.Getenv("HOME"), ".docker/public-key.json"))
	if err != nil {
		return err
	}
	defer f.Close()

	cmdString := fmt.Sprintf("sudo mkdir -p %q && sudo tee -a %q", "/root/.docker/authorized-keys.d", "/root/.docker/authorized-keys.d/docker-host.json")
	cmd, err = driver.GetSSHCommand(cmdString)
	if err != nil {
		return err
	}
	cmd.Stdin = f
	if err := cmd.Run(); err != nil {
		return err
	}

	// HACK: change back ownership
	cmd, err = driver.GetSSHCommand("sudo mkdir -p /root/.docker && sudo chown -R root /root/.docker")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd, err = driver.GetSSHCommand("sudo start docker")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

// EncodeToBase64 returns the string as a base64 encoded string
func EncodeToBase64(data string) string {
	buf := bytes.NewBufferString(data)
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return encoded
}

// WaitForDocker will retry until either a successful connection or maximum retries is reached
func WaitForDocker(url string, maxRetries int) bool {
	counter := 0
	for {
		conn, err := net.DialTimeout("tcp", url, time.Duration(1)*time.Second)
		if err != nil {
			counter++
			if counter == maxRetries {
				return false
			}
			time.Sleep(1 * time.Second)
			continue
		}
		defer conn.Close()
		break
	}
	return true
}
