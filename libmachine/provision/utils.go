package provision

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/cert"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/provision/serviceaction"
)

type DockerOptions struct {
	EngineOptions     string
	EngineOptionsPath string
}

func installDockerGeneric(p Provisioner, baseURL string) error {
	// install docker - until cloudinit we use ubuntu everywhere so we
	// just install it using the docker repos
	if output, err := p.SSHCommand(fmt.Sprintf("if ! type docker; then curl -sSL %s | sh -; fi", baseURL)); err != nil {
		return fmt.Errorf("error installing docker: %s\n", output)
	}

	return nil
}

func makeDockerOptionsDir(p Provisioner) error {
	dockerDir := p.GetDockerOptionsDir()
	if _, err := p.SSHCommand(fmt.Sprintf("sudo mkdir -p %s", dockerDir)); err != nil {
		return err
	}

	return nil
}

func setRemoteAuthOptions(p Provisioner) auth.Options {
	dockerDir := p.GetDockerOptionsDir()
	authOptions := p.GetAuthOptions()

	// due to windows clients, we cannot use filepath.Join as the paths
	// will be mucked on the linux hosts
	authOptions.CaCertRemotePath = path.Join(dockerDir, "ca.pem")
	authOptions.ServerCertRemotePath = path.Join(dockerDir, "server.pem")
	authOptions.ServerKeyRemotePath = path.Join(dockerDir, "server-key.pem")

	return authOptions
}

func ConfigureAuth(p Provisioner) error {
	var (
		err error
	)

	driver := p.GetDriver()
	machineName := driver.GetMachineName()
	authOptions := p.GetAuthOptions()
	org := mcnutils.GetUsername() + "." + machineName
	bits := 2048

	ip, err := driver.GetIP()
	if err != nil {
		return err
	}

	log.Info("Copying certs to the local machine directory...")

	if err := mcnutils.CopyFile(authOptions.CaCertPath, filepath.Join(authOptions.StorePath, "ca.pem")); err != nil {
		return fmt.Errorf("Copying ca.pem to machine dir failed: %s", err)
	}

	if err := mcnutils.CopyFile(authOptions.ClientCertPath, filepath.Join(authOptions.StorePath, "cert.pem")); err != nil {
		return fmt.Errorf("Copying cert.pem to machine dir failed: %s", err)
	}

	if err := mcnutils.CopyFile(authOptions.ClientKeyPath, filepath.Join(authOptions.StorePath, "key.pem")); err != nil {
		return fmt.Errorf("Copying key.pem to machine dir failed: %s", err)
	}

	log.Debugf("generating server cert: %s ca-key=%s private-key=%s org=%s",
		authOptions.ServerCertPath,
		authOptions.CaCertPath,
		authOptions.CaPrivateKeyPath,
		org,
	)

	// TODO: Switch to passing just authOptions to this func
	// instead of all these individual fields
	err = cert.GenerateCert(
		[]string{ip, "localhost"},
		authOptions.ServerCertPath,
		authOptions.ServerKeyPath,
		authOptions.CaCertPath,
		authOptions.CaPrivateKeyPath,
		org,
		bits,
	)

	if err != nil {
		return fmt.Errorf("error generating server cert: %s", err)
	}

	if err := p.Service("docker", serviceaction.Stop); err != nil {
		return err
	}

	// upload certs and configure TLS auth
	caCert, err := ioutil.ReadFile(authOptions.CaCertPath)
	if err != nil {
		return err
	}

	serverCert, err := ioutil.ReadFile(authOptions.ServerCertPath)
	if err != nil {
		return err
	}
	serverKey, err := ioutil.ReadFile(authOptions.ServerKeyPath)
	if err != nil {
		return err
	}

	log.Info("Copying certs to the remote machine...")

	// printf will choke if we don't pass a format string because of the
	// dashes, so that's the reason for the '%%s'
	certTransferCmdFmt := "printf '%%s' '%s' | sudo tee %s"

	// These ones are for Jessie and Mike <3 <3 <3
	if _, err := p.SSHCommand(fmt.Sprintf(certTransferCmdFmt, string(caCert), authOptions.CaCertRemotePath)); err != nil {
		return err
	}

	if _, err := p.SSHCommand(fmt.Sprintf(certTransferCmdFmt, string(serverCert), authOptions.ServerCertRemotePath)); err != nil {
		return err
	}

	if _, err := p.SSHCommand(fmt.Sprintf(certTransferCmdFmt, string(serverKey), authOptions.ServerKeyRemotePath)); err != nil {
		return err
	}

	dockerURL, err := driver.GetURL()
	if err != nil {
		return err
	}
	u, err := url.Parse(dockerURL)
	if err != nil {
		return err
	}
	dockerPort := 2376
	parts := strings.Split(u.Host, ":")
	if len(parts) == 2 {
		dPort, err := strconv.Atoi(parts[1])
		if err != nil {
			return err
		}
		dockerPort = dPort
	}

	dkrcfg, err := p.GenerateDockerOptions(dockerPort)
	if err != nil {
		return err
	}

	log.Info("Setting Docker configuration on the remote daemon...")

	if _, err = p.SSHCommand(fmt.Sprintf("printf %%s \"%s\" | sudo tee %s", dkrcfg.EngineOptions, dkrcfg.EngineOptionsPath)); err != nil {
		return err
	}

	if err := p.Service("docker", serviceaction.Start); err != nil {
		return err
	}

	return waitForDocker(p, dockerPort)
}

func matchNetstatOut(reDaemonListening, netstatOut string) bool {
	// TODO: I would really prefer this be a Scanner directly on
	// the STDOUT of the executed command than to do all the string
	// manipulation hokey-pokey.
	//
	// TODO: Unit test this matching.
	for _, line := range strings.Split(netstatOut, "\n") {
		match, err := regexp.MatchString(reDaemonListening, line)
		if err != nil {
			log.Warnf("Regex warning: %s", err)
		}
		if match && line != "" {
			return true
		}
	}

	return false
}

func checkDaemonUp(p Provisioner, dockerPort int) func() bool {
	reDaemonListening := fmt.Sprintf(":%d.*LISTEN", dockerPort)
	return func() bool {
		// HACK: Check netstat's output to see if anyone's listening on the Docker API port.
		netstatOut, err := p.SSHCommand("netstat -a")
		if err != nil {
			log.Warnf("Error running SSH command: %s", err)
			return false
		}

		return matchNetstatOut(reDaemonListening, netstatOut)
	}
}

func waitForDocker(p Provisioner, dockerPort int) error {
	if err := mcnutils.WaitForSpecific(checkDaemonUp(p, dockerPort), 5, 3*time.Second); err != nil {
		return NewErrDaemonAvailable(err)
	}

	return nil
}
