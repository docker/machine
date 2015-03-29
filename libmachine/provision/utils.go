package provision

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/url"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/provision/pkgaction"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/docker/machine/utils"
)

type DockerOptions struct {
	EngineOptions     string
	EngineOptionsPath string
}

func installDockerGeneric(p Provisioner) error {
	// install docker - until cloudinit we use ubuntu everywhere so we
	// just install it using the docker repos
	if output, err := p.SSHCommand("if ! type docker; then curl -sSL https://get.docker.com | sh -; fi"); err != nil {
		var buf bytes.Buffer
		if _, err := buf.ReadFrom(output.Stderr); err != nil {
			return err
		}

		return fmt.Errorf("error installing docker: %s\n", buf.String())
	}

	return nil
}

func ConfigureAuth(p Provisioner, authOptions auth.AuthOptions) error {
	var (
		err error
	)

	machineName := p.GetDriver().GetMachineName()
	org := machineName
	bits := 2048

	ip, err := p.GetDriver().GetIP()
	if err != nil {
		return err
	}

	// copy certs to client dir for docker client
	machineDir := filepath.Join(utils.GetMachineDir(), machineName)

	if err := utils.CopyFile(authOptions.CaCertPath, filepath.Join(machineDir, "ca.pem")); err != nil {
		log.Fatalf("Error copying ca.pem to machine dir: %s", err)
	}

	if err := utils.CopyFile(authOptions.ClientCertPath, filepath.Join(machineDir, "cert.pem")); err != nil {
		log.Fatalf("Error copying cert.pem to machine dir: %s", err)
	}

	if err := utils.CopyFile(authOptions.ClientKeyPath, filepath.Join(machineDir, "key.pem")); err != nil {
		log.Fatalf("Error copying key.pem to machine dir: %s", err)
	}

	log.Debugf("generating server cert: %s ca-key=%s private-key=%s org=%s",
		authOptions.ServerCertPath,
		authOptions.CaCertPath,
		authOptions.PrivateKeyPath,
		org,
	)

	// TODO: Switch to passing just authOptions to this func
	// instead of all these individual fields
	err = utils.GenerateCert(
		[]string{ip},
		authOptions.ServerCertPath,
		authOptions.ServerKeyPath,
		authOptions.CaCertPath,
		authOptions.PrivateKeyPath,
		org,
		bits,
	)
	if err != nil {
		return fmt.Errorf("error generating server cert: %s", err)
	}

	if err := p.Service("docker", pkgaction.Stop); err != nil {
		return err
	}

	dockerDir := p.GetDockerOptionsDir()

	if _, err := p.SSHCommand(fmt.Sprintf("sudo mkdir -p %s", dockerDir)); err != nil {
		return err
	}

	// upload certs and configure TLS auth
	caCert, err := ioutil.ReadFile(authOptions.CaCertPath)
	if err != nil {
		return err
	}

	// due to windows clients, we cannot use filepath.Join as the paths
	// will be mucked on the linux hosts
	machineCaCertPath := path.Join(dockerDir, "ca.pem")
	authOptions.CaCertRemotePath = machineCaCertPath

	serverCert, err := ioutil.ReadFile(authOptions.ServerCertPath)
	if err != nil {
		return err
	}
	machineServerCertPath := path.Join(dockerDir, "server.pem")
	authOptions.ServerCertRemotePath = machineServerCertPath

	serverKey, err := ioutil.ReadFile(authOptions.ServerKeyPath)
	if err != nil {
		return err
	}
	machineServerKeyPath := path.Join(dockerDir, "server-key.pem")
	authOptions.ServerKeyRemotePath = machineServerKeyPath

	if _, err = p.SSHCommand(fmt.Sprintf("echo \"%s\" | sudo tee %s", string(caCert), machineCaCertPath)); err != nil {
		return err
	}

	if _, err = p.SSHCommand(fmt.Sprintf("echo \"%s\" | sudo tee %s", string(serverKey), machineServerKeyPath)); err != nil {
		return err
	}

	if _, err = p.SSHCommand(fmt.Sprintf("echo \"%s\" | sudo tee %s", string(serverCert), machineServerCertPath)); err != nil {
		return err
	}

	dockerUrl, err := p.GetDriver().GetURL()
	if err != nil {
		return err
	}
	u, err := url.Parse(dockerUrl)
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

	dkrcfg, err := p.GenerateDockerOptions(dockerPort, authOptions)
	if err != nil {
		return err
	}

	if _, err = p.SSHCommand(fmt.Sprintf("echo \"%s\" | sudo tee -a %s", dkrcfg.EngineOptions, dkrcfg.EngineOptionsPath)); err != nil {
		return err
	}

	if err := p.Service("docker", pkgaction.Start); err != nil {
		return err
	}

	return nil
}

func getDefaultDaemonOpts(driverName string, authOptions auth.AuthOptions) string {
	return fmt.Sprintf(`--tlsverify --tlscacert=%s --tlskey=%s --tlscert=%s %s`,
		authOptions.CaCertRemotePath,
		authOptions.ServerKeyRemotePath,
		authOptions.ServerCertRemotePath,
		fmt.Sprintf("--label=provider=%s", driverName),
	)
}

func configureSwarm(p Provisioner, swarmOptions swarm.SwarmOptions) error {
	if !swarmOptions.IsSwarm {
		return nil
	}

	basePath := p.GetDockerOptionsDir()
	ip, err := p.GetDriver().GetIP()
	if err != nil {
		return err
	}

	tlsCaCert := path.Join(basePath, "ca.pem")
	tlsCert := path.Join(basePath, "server.pem")
	tlsKey := path.Join(basePath, "server-key.pem")
	masterArgs := fmt.Sprintf("--tlsverify --tlscacert=%s --tlscert=%s --tlskey=%s -H %s %s",
		tlsCaCert, tlsCert, tlsKey, swarmOptions.Host, swarmOptions.Discovery)
	nodeArgs := fmt.Sprintf("--addr %s:2376 %s", ip, swarmOptions.Discovery)

	u, err := url.Parse(swarmOptions.Host)
	if err != nil {
		return err
	}

	parts := strings.Split(u.Host, ":")
	port := parts[1]

	// TODO: Do not hardcode daemon port, ask the driver
	if err := utils.WaitForDocker(ip, 2376); err != nil {
		return err
	}

	if _, err := p.SSHCommand(fmt.Sprintf("sudo docker pull %s", swarm.DockerImage)); err != nil {
		return err
	}

	dockerDir := p.GetDockerOptionsDir()

	// if master start master agent
	if swarmOptions.Master {
		log.Debug("launching swarm master")
		log.Debugf("master args: %s", masterArgs)
		if _, err = p.SSHCommand(fmt.Sprintf("sudo docker run -d -p %s:%s --restart=always --name swarm-agent-master -v %s:%s %s manage %s",
			port, port, dockerDir, dockerDir, swarm.DockerImage, masterArgs)); err != nil {
			return err
		}
	}

	// start node agent
	log.Debug("launching swarm node")
	log.Debugf("node args: %s", nodeArgs)
	if _, err = p.SSHCommand(fmt.Sprintf("sudo docker run -d --restart=always --name swarm-agent -v %s:%s %s join %s",
		dockerDir, dockerDir, swarm.DockerImage, nodeArgs)); err != nil {
		return err
	}

	return nil
}
