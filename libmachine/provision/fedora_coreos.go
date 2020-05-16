package provision

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/provision/pkgaction"
	"github.com/docker/machine/libmachine/swarm"
)

func init() {
	Register("Fedora", &RegisteredProvisioner{
		New: NewFedoraCoreOSProvisioner,
	})
}

// NewFedoraCoreOSProvisioner creates a new provisioner for a driver
func NewFedoraCoreOSProvisioner(d drivers.Driver) Provisioner {
	return &FedoraCoreOSProvisioner{
		NewSystemdProvisioner("fedora", d),
	}
}

// FedoraCoreOSProvisioner is a provisioner based on the CoreOS provisioner
type FedoraCoreOSProvisioner struct {
	SystemdProvisioner
}

// String returns the name of the provisioner
func (provisioner *FedoraCoreOSProvisioner) String() string {
	return "Fedora CoreOS"
}

// SetHostname sets the hostname of the remote machine
func (provisioner *FedoraCoreOSProvisioner) SetHostname(hostname string) error {
	log.Debugf("SetHostname: %s", hostname)

	command := fmt.Sprintf("sudo hostnamectl set-hostname %s", hostname)
	if _, err := provisioner.SSHCommand(command); err != nil {
		return err
	}

	return nil
}

// GenerateDockerOptions formats a systemd drop-in unit which adds support for
// Docker Machine
func (provisioner *FedoraCoreOSProvisioner) GenerateDockerOptions(dockerPort int) (*DockerOptions, error) {
	var (
		engineCfg bytes.Buffer
	)

	driverNameLabel := fmt.Sprintf("provider=%s", provisioner.Driver.DriverName())
	provisioner.EngineOptions.Labels = append(provisioner.EngineOptions.Labels, driverNameLabel)

	engineConfigTmpl := `[Service]
Environment=TMPDIR=/var/tmp
ExecStart=
ExecStart=/usr/bin/dockerd \
				 --exec-opt native.cgroupdriver=systemd \
				 --host=unix:///var/run/docker.sock \
				 --host=tcp://0.0.0.0:{{.DockerPort}} \
				 --tlsverify \
				 --tlscacert {{.AuthOptions.CaCertRemotePath}} \
				 --tlscert {{.AuthOptions.ServerCertRemotePath}} \
				 --tlskey {{.AuthOptions.ServerKeyRemotePath}}{{ range .EngineOptions.Labels }} \
				 --label {{.}}{{ end }}{{ range .EngineOptions.InsecureRegistry }} \
				 --insecure-registry {{.}}{{ end }}{{ range .EngineOptions.RegistryMirror }} \
				 --registry-mirror {{.}}{{ end }}{{ range .EngineOptions.ArbitraryFlags }} \
				 --{{.}}{{ end }} \$DOCKER_OPTS \$DOCKER_OPT_BIP \$DOCKER_OPT_MTU \$DOCKER_OPT_IPMASQ
Environment={{range .EngineOptions.Env}}{{ printf "%q" . }} {{end}}
`

	t, err := template.New("engineConfig").Parse(engineConfigTmpl)
	if err != nil {
		return nil, err
	}

	engineConfigContext := EngineConfigContext{
		DockerPort:    dockerPort,
		AuthOptions:   provisioner.AuthOptions,
		EngineOptions: provisioner.EngineOptions,
	}

	t.Execute(&engineCfg, engineConfigContext)

	return &DockerOptions{
		EngineOptions:     engineCfg.String(),
		EngineOptionsPath: provisioner.DaemonOptionsFile,
	}, nil
}

// CompatibleWithHost returns whether or not this provisoner is compatible
// with the target host
func (provisioner *FedoraCoreOSProvisioner) CompatibleWithHost() bool {
	isFedora := provisioner.OsReleaseInfo.ID == "fedora"
	isCoreOS := provisioner.OsReleaseInfo.VariantID == "coreos"
	return isFedora && isCoreOS
}

// Package installs a package on the remote host. The Fedora CoreOS provisioner
// does not support (or need) any package installation
func (provisioner *FedoraCoreOSProvisioner) Package(name string, action pkgaction.PackageAction) error {
	return nil
}

// Provision provisions the machine
func (provisioner *FedoraCoreOSProvisioner) Provision(swarmOptions swarm.Options, authOptions auth.Options, engineOptions engine.Options) error {
	provisioner.SwarmOptions = swarmOptions
	provisioner.AuthOptions = authOptions
	provisioner.EngineOptions = engineOptions

	if err := provisioner.SetHostname(provisioner.Driver.GetMachineName()); err != nil {
		return err
	}

	if err := makeDockerOptionsDir(provisioner); err != nil {
		return err
	}

	log.Debugf("Preparing certificates")
	provisioner.AuthOptions = setRemoteAuthOptions(provisioner)

	log.Debugf("Setting up certificates")
	if err := ConfigureAuth(provisioner); err != nil {
		return err
	}

	log.Debug("Configuring swarm")
	err := configureSwarm(provisioner, swarmOptions, provisioner.AuthOptions)
	return err
}
