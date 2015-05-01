package provision

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/docker/machine/drivers"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/provision/pkgaction"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/docker/machine/log"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/utils"
)

func init() {
	Register("RedHat", &RegisteredProvisioner{
		New: NewRedHatProvisioner,
	})
}

func NewRedHatProvisioner(d drivers.Driver) Provisioner {
	return &RedHatProvisioner{
		packages: []string{
			"curl",
		},
		Driver: d,
	}
}

type RedHatProvisioner struct {
	packages      []string
	OsReleaseInfo *OsRelease
	Driver        drivers.Driver
	AuthOptions   auth.AuthOptions
	EngineOptions engine.EngineOptions
	SwarmOptions  swarm.SwarmOptions
}

func (provisioner *RedHatProvisioner) Service(name string, action pkgaction.ServiceAction) error {
	command := fmt.Sprintf("sudo systemctl %s %s", action.String(), name)

	if _, err := provisioner.SSHCommand(command); err != nil {
		return err
	}

	return nil
}

func (provisioner *RedHatProvisioner) Package(name string, action pkgaction.PackageAction) error {
	var packageAction string

	switch action {
	case pkgaction.Install:
		packageAction = "install"
	case pkgaction.Remove:
		packageAction = "remove"
	case pkgaction.Upgrade:
		packageAction = "upgrade"
	}

	switch name {
	case "docker":
		name = "docker"
	}

	command := fmt.Sprintf("sudo -E yum %s -y %s", packageAction, name)

	if _, err := provisioner.SSHCommand(command); err != nil {
		return err
	}

	return nil
}

func (provisioner *RedHatProvisioner) isAWS() bool {
	if _, err := provisioner.SSHCommand("curl -s http://169.254.169.254/latest/meta-data/ami-id"); err != nil {
		return false
	}

	return true
}

func (provisioner *RedHatProvisioner) configureRepos() error {

	// TODO: should this be handled differently? on aws we need to enable
	// the extras repo different than a standalone rhel box

	repoCmd := "subscription-manager repos --enable=rhel-7-server-extras-rpms"
	if provisioner.isAWS() {
		repoCmd = "yum-config-manager --enable rhui-REGION-rhel-server-extras"
	}

	if _, err := provisioner.SSHCommand(fmt.Sprintf("sudo %s", repoCmd)); err != nil {
		return err
	}

	return nil
}

func (provisioner *RedHatProvisioner) installDocker() error {
	if err := provisioner.Package("docker", pkgaction.Install); err != nil {
		return err
	}

	if err := provisioner.Service("docker", pkgaction.Restart); err != nil {
		return err
	}

	if err := provisioner.Service("docker", pkgaction.Enable); err != nil {
		return err
	}

	return nil
}

func (provisioner *RedHatProvisioner) dockerDaemonResponding() bool {
	if _, err := provisioner.SSHCommand("sudo docker version"); err != nil {
		log.Warn("Error getting SSH command to check if the daemon is up: %s", err)
		return false
	}

	// The daemon is up if the command worked.  Carry on.
	return true
}

func (provisioner *RedHatProvisioner) Provision(swarmOptions swarm.SwarmOptions, authOptions auth.AuthOptions, engineOptions engine.EngineOptions) error {
	provisioner.SwarmOptions = swarmOptions
	provisioner.AuthOptions = authOptions
	provisioner.EngineOptions = engineOptions

	// set default storage driver for redhat
	provisioner.EngineOptions.StorageDriver = "devicemapper"

	if err := provisioner.SetHostname(provisioner.Driver.GetMachineName()); err != nil {
		return err
	}

	// setup extras repo
	if err := provisioner.configureRepos(); err != nil {
		return err
	}

	for _, pkg := range provisioner.packages {
		if err := provisioner.Package(pkg, pkgaction.Install); err != nil {
			return err
		}
	}

	// install docker
	if err := provisioner.installDocker(); err != nil {
		return err
	}

	if err := utils.WaitFor(provisioner.dockerDaemonResponding); err != nil {
		return err
	}

	if err := makeDockerOptionsDir(provisioner); err != nil {
		return err
	}

	provisioner.AuthOptions = setRemoteAuthOptions(provisioner)

	if err := ConfigureAuth(provisioner); err != nil {
		return err
	}

	if err := configureSwarm(provisioner, swarmOptions); err != nil {
		return err
	}

	return nil
}

func (provisioner *RedHatProvisioner) Hostname() (string, error) {
	output, err := provisioner.SSHCommand("hostname")
	if err != nil {
		return "", err
	}

	var so bytes.Buffer
	if _, err := so.ReadFrom(output.Stdout); err != nil {
		return "", err
	}

	return so.String(), nil
}

func (provisioner *RedHatProvisioner) SetHostname(hostname string) error {
	if out, err := provisioner.SSHCommand(fmt.Sprintf(
		"sudo hostname %s && echo %q | sudo tee /etc/hostname",
		hostname,
		hostname,
	)); err != nil {
		log.Info(out)
		log.Errorf("error setting hostname: %s", err)
		return err
	}

	if _, err := provisioner.SSHCommand(fmt.Sprintf(
		"if grep -xq 127.0.1.1.* /etc/hosts; then sudo sed -i 's/^127.0.1.1.*/127.0.1.1 %s/g' /etc/hosts; else echo '127.0.1.1 %s' | sudo tee -a /etc/hosts; fi",
		hostname,
		hostname,
	)); err != nil {
		log.Errorf("error setting /etc/hosts: %s", err)
		return err
	}

	return nil
}

func (provisioner *RedHatProvisioner) GetDockerOptionsDir() string {
	return "/etc/docker"
}

func (provisioner *RedHatProvisioner) SSHCommand(args string) (ssh.Output, error) {
	return drivers.RunSSHCommandFromDriver(provisioner.Driver, args)
}

func (provisioner *RedHatProvisioner) CompatibleWithHost() bool {
	return provisioner.OsReleaseInfo.Id == "rhel"
}

func (provisioner *RedHatProvisioner) GetAuthOptions() auth.AuthOptions {
	return provisioner.AuthOptions
}

func (provisioner *RedHatProvisioner) SetOsReleaseInfo(info *OsRelease) {
	provisioner.OsReleaseInfo = info
}

func (provisioner *RedHatProvisioner) GenerateDockerOptions(dockerPort int) (*DockerOptions, error) {
	var (
		engineCfg  bytes.Buffer
		configPath = "/etc/sysconfig/docker"
	)

	// remove existing
	if _, err := provisioner.SSHCommand(fmt.Sprintf("sudo rm %s", configPath)); err != nil {
		return nil, err
	}

	driverNameLabel := fmt.Sprintf("provider=%s", provisioner.Driver.DriverName())
	provisioner.EngineOptions.Labels = append(provisioner.EngineOptions.Labels, driverNameLabel)

	// systemd / redhat will not load options if they are on newlines
	// instead, it just continues with a different set of options; yeah...
	engineConfigTmpl := `
OPTIONS='--selinux-enabled -H tcp://0.0.0.0:{{.DockerPort}} -H unix:///var/run/docker.sock --storage-driver {{.EngineOptions.StorageDriver}} --tlsverify --tlscacert {{.AuthOptions.CaCertRemotePath}} --tlscert {{.AuthOptions.ServerCertRemotePath}} --tlskey {{.AuthOptions.ServerKeyRemotePath}} {{ range .EngineOptions.Labels }}--label {{.}} {{ end }}{{ range .EngineOptions.InsecureRegistry }}--insecure-registry {{.}} {{ end }}{{ range .EngineOptions.RegistryMirror }}--registry-mirror {{.}} {{ end }}{{ range .EngineOptions.ArbitraryFlags }}--{{.}} {{ end }}'
DOCKER_CERT_PATH=/etc/docker
ADD_REGISTRY='--add-registry registry.access.redhat.com'
GOTRACEBACK='crash'
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

	daemonOptsDir := configPath
	return &DockerOptions{
		EngineOptions:     engineCfg.String(),
		EngineOptionsPath: daemonOptsDir,
	}, nil
}

func (provisioner *RedHatProvisioner) GetDriver() drivers.Driver {
	return provisioner.Driver
}
