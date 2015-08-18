package provision

import (
	"bytes"
	"fmt"
	"path"
	"text/template"

	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/provision/pkgaction"
	"github.com/docker/machine/libmachine/provision/serviceaction"
	"github.com/docker/machine/libmachine/state"
	"github.com/docker/machine/libmachine/swarm"
)

func init() {
	Register("boot2docker", &RegisteredProvisioner{
		New: NewBoot2DockerProvisioner,
	})
}

func NewBoot2DockerProvisioner(d drivers.Driver) Provisioner {
	return &Boot2DockerProvisioner{
		Driver: d,
	}
}

type Boot2DockerProvisioner struct {
	OsReleaseInfo *OsRelease
	Driver        drivers.Driver
	AuthOptions   auth.AuthOptions
	EngineOptions engine.EngineOptions
	SwarmOptions  swarm.SwarmOptions
}

func (provisioner *Boot2DockerProvisioner) Service(name string, action serviceaction.ServiceAction) error {
	var (
		err error
	)

	if _, err = provisioner.SSHCommand(fmt.Sprintf("sudo /etc/init.d/%s %s", name, action.String())); err != nil {
		return err
	}

	return nil
}

func (provisioner *Boot2DockerProvisioner) upgradeIso() error {
	log.Info("Stopping machine to do the upgrade...")

	if err := provisioner.Driver.Stop(); err != nil {
		return err
	}

	if err := mcnutils.WaitFor(drivers.MachineInState(provisioner.Driver, state.Stopped)); err != nil {
		return err
	}

	machineName := provisioner.GetDriver().GetMachineName()

	log.Infof("Upgrading machine %s...", machineName)

	// TODO: Ideally, we should not read from mcndirs directory at all.
	// The driver should be able to communicate how and where to place the
	// relevant files.
	b2dutils := mcnutils.NewB2dUtils("", "", mcndirs.GetBaseDir())

	// Usually we call this implicitly, but call it here explicitly to get
	// the latest boot2docker ISO.
	if err := b2dutils.DownloadLatestBoot2Docker(); err != nil {
		return err
	}

	// Copy the latest version of boot2docker ISO to the machine's directory
	if err := b2dutils.CopyIsoToMachineDir("", machineName); err != nil {
		return err
	}

	log.Infof("Starting machine back up...")

	if err := provisioner.Driver.Start(); err != nil {
		return err
	}

	return mcnutils.WaitFor(drivers.MachineInState(provisioner.Driver, state.Running))
}

func (provisioner *Boot2DockerProvisioner) Package(name string, action pkgaction.PackageAction) error {
	if name == "docker" && action == pkgaction.Upgrade {
		if err := provisioner.upgradeIso(); err != nil {
			return err
		}
	}
	return nil
}

func (provisioner *Boot2DockerProvisioner) Hostname() (string, error) {
	return provisioner.SSHCommand("hostname")
}

func (provisioner *Boot2DockerProvisioner) SetHostname(hostname string) error {
	if _, err := provisioner.SSHCommand(fmt.Sprintf(
		"sudo /usr/bin/sethostname %s && echo %q | sudo tee /var/lib/boot2docker/etc/hostname",
		hostname,
		hostname,
	)); err != nil {
		return err
	}

	return nil
}

func (provisioner *Boot2DockerProvisioner) GetDockerOptionsDir() string {
	return "/var/lib/boot2docker"
}

func (provisioner *Boot2DockerProvisioner) GetAuthOptions() auth.AuthOptions {
	return provisioner.AuthOptions
}

func (provisioner *Boot2DockerProvisioner) GenerateDockerOptions(dockerPort int) (*DockerOptions, error) {
	var (
		engineCfg bytes.Buffer
	)

	driverNameLabel := fmt.Sprintf("provider=%s", provisioner.Driver.DriverName())
	provisioner.EngineOptions.Labels = append(provisioner.EngineOptions.Labels, driverNameLabel)

	engineConfigTmpl := `
EXTRA_ARGS='
{{ range .EngineOptions.Labels }}--label {{.}}
{{ end }}{{ range .EngineOptions.InsecureRegistry }}--insecure-registry {{.}}
{{ end }}{{ range .EngineOptions.RegistryMirror }}--registry-mirror {{.}}
{{ end }}{{ range .EngineOptions.ArbitraryFlags }}--{{.}}
{{ end }}
'
CACERT={{.AuthOptions.CaCertRemotePath}}
DOCKER_HOST='-H tcp://0.0.0.0:{{.DockerPort}}'
DOCKER_STORAGE={{.EngineOptions.StorageDriver}}
DOCKER_TLS=auto
SERVERKEY={{.AuthOptions.ServerKeyRemotePath}}
SERVERCERT={{.AuthOptions.ServerCertRemotePath}}

{{range .EngineOptions.Env}}export \"{{ printf "%q" . }}\"
{{end}}
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

	daemonOptsDir := path.Join(provisioner.GetDockerOptionsDir(), "profile")
	return &DockerOptions{
		EngineOptions:     engineCfg.String(),
		EngineOptionsPath: daemonOptsDir,
	}, nil
}

func (provisioner *Boot2DockerProvisioner) CompatibleWithHost() bool {
	return provisioner.OsReleaseInfo.Id == "boot2docker"
}

func (provisioner *Boot2DockerProvisioner) SetOsReleaseInfo(info *OsRelease) {
	provisioner.OsReleaseInfo = info
}

func (provisioner *Boot2DockerProvisioner) GetOsReleaseInfo() (*OsRelease, error) {
	return provisioner.OsReleaseInfo, nil
}

func (provisioner *Boot2DockerProvisioner) Provision(swarmOptions swarm.SwarmOptions, authOptions auth.AuthOptions, engineOptions engine.EngineOptions) error {
	provisioner.SwarmOptions = swarmOptions
	provisioner.AuthOptions = authOptions
	provisioner.EngineOptions = engineOptions

	if provisioner.EngineOptions.StorageDriver == "" {
		provisioner.EngineOptions.StorageDriver = "aufs"
	}

	if err := provisioner.SetHostname(provisioner.Driver.GetMachineName()); err != nil {
		return err
	}

	ip, err := provisioner.GetDriver().GetIP()
	if err != nil {
		return err
	}

	// b2d hosts need to wait for the daemon to be up
	// before continuing with provisioning
	if err := mcnutils.WaitForDocker(ip, 2376); err != nil {
		return err
	}

	if err := makeDockerOptionsDir(provisioner); err != nil {
		return err
	}

	provisioner.AuthOptions = setRemoteAuthOptions(provisioner)

	if err := ConfigureAuth(provisioner); err != nil {
		return err
	}

	if err := configureSwarm(provisioner, swarmOptions, provisioner.AuthOptions); err != nil {
		return err
	}

	return nil
}

func (provisioner *Boot2DockerProvisioner) SSHCommand(args string) (string, error) {
	return drivers.RunSSHCommandFromDriver(provisioner.Driver, args)
}

func (provisioner *Boot2DockerProvisioner) GetDriver() drivers.Driver {
	return provisioner.Driver
}
