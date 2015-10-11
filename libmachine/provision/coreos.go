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
	"github.com/docker/machine/libmachine/provision/serviceaction"
	"github.com/docker/machine/libmachine/swarm"
)

const (
	hostTmpl = `sudo tee /var/tmp/hostname.yml << EOF
#cloud-config

hostname: %s
EOF
`

	cloudConfigTmpl = `sudo tee /var/tmp/cloudinit.yml << EOF
#cloud-config

coreos:
  etcd:
    discovery: {{.Token}}
    addr: {{.IP}}:4001
    peer-addr: {{.IP}}:7001
    # give etcd more time if it's under heavy load - prevent leader election thrashing
    peer-election-timeout: 4000
    # heartbeat interval should ideally be 1/4 or 1/5 of peer election timeout
    peer-heartbeat-interval: 1000
    # don't keep all day in memory
    snapshot: true
  fleet:
    public-ip: {{.IP}}
    # allow etcd to slow down at times
    etcd-request-timeout: 15
  units:
    - name: etcd.service
      command: start
    - name: fleet.service
      command: start
EOF
`
)

func init() {
	Register("CoreOS", &RegisteredProvisioner{
		New: NewCoreOSProvisioner,
	})
}

func NewCoreOSProvisioner(d drivers.Driver) Provisioner {
	return &CoreOSProvisioner{
		GenericProvisioner{
			DockerOptionsDir:  "/etc/docker",
			DaemonOptionsFile: "/etc/systemd/system/docker.service",
			OsReleaseId:       "coreos",
			Driver:            d,
		},
	}
}

type CoreOSProvisioner struct {
	GenericProvisioner
}

func (provisioner *CoreOSProvisioner) Service(name string, action serviceaction.ServiceAction) error {
	// daemon-reload to catch config updates; systemd -- ugh
	if _, err := provisioner.SSHCommand("sudo systemctl daemon-reload"); err != nil {
		return err
	}

	command := fmt.Sprintf("sudo systemctl %s %s", action.String(), name)

	if _, err := provisioner.SSHCommand(command); err != nil {
		return err
	}

	// wait until docker is running
	if (name == "docker") && (action.String() == "start") {
		if err := waitForDocker(provisioner, 2376); err != nil {
			return err
		}
	}

	return nil
}

func (provisioner *CoreOSProvisioner) SetHostname(hostname string) error {
	log.Debugf("SetHostname: %s", hostname)

	if _, err := provisioner.SSHCommand(fmt.Sprintf(hostTmpl, hostname)); err != nil {
		return err
	}

	if _, err := provisioner.SSHCommand("sudo systemctl start system-cloudinit@var-tmp-hostname.yml.service"); err != nil {
		return err
	}

	return nil
}

func (provisioner *CoreOSProvisioner) SetCloudInit(hostname string) error {
	log.Debugf("SetCloudInit: %s", hostname)

	ip, err := provisioner.Driver.GetIP()
	if err != nil {
		log.Fatalf("Could not get IP address for created machine: %s", err)
	}

	token, err := provisioner.SSHCommand("curl -sSL https://discovery.etcd.io/new?size=3")
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"IP":       ip,
		"HostName": hostname,
		"Token":    token,
	}

	t := template.Must(template.New("cloudinit").Parse(cloudConfigTmpl))
	buf := &bytes.Buffer{}
	if err := t.Execute(buf, data); err != nil {
		panic(err)
	}
	cloudConfigTmpl := buf.String()

	if _, err := provisioner.SSHCommand(cloudConfigTmpl); err != nil {
		return err
	}

	if _, err := provisioner.SSHCommand("sudo systemctl start system-cloudinit@var-tmp-cloudinit.yml.service"); err != nil {
		return err
	}

	return nil
}

func (provisioner *CoreOSProvisioner) GenerateDockerOptions(dockerPort int) (*DockerOptions, error) {
	var (
		engineCfg bytes.Buffer
	)

	driverNameLabel := fmt.Sprintf("provider=%s", provisioner.Driver.DriverName())
	provisioner.EngineOptions.Labels = append(provisioner.EngineOptions.Labels, driverNameLabel)

	engineConfigTmpl := `[Unit]
Description=Docker Socket for the API
After=docker.socket early-docker.target network.target
Requires=docker.socket early-docker.target

[Service]
Environment=TMPDIR=/var/tmp
EnvironmentFile=-/run/flannel_docker_opts.env
MountFlags=slave
LimitNOFILE=1048576
LimitNPROC=1048576
ExecStart=/usr/lib/coreos/dockerd --daemon --host=unix:///var/run/docker.sock --host=tcp://0.0.0.0:{{.DockerPort}} --tlsverify --tlscacert {{.AuthOptions.CaCertRemotePath}} --tlscert {{.AuthOptions.ServerCertRemotePath}} --tlskey {{.AuthOptions.ServerKeyRemotePath}}{{ range .EngineOptions.Labels }} --label {{.}}{{ end }}{{ range .EngineOptions.InsecureRegistry }} --insecure-registry {{.}}{{ end }}{{ range .EngineOptions.RegistryMirror }} --registry-mirror {{.}}{{ end }}{{ range .EngineOptions.ArbitraryFlags }} --{{.}}{{ end }} \$DOCKER_OPTS \$DOCKER_OPT_BIP \$DOCKER_OPT_MTU \$DOCKER_OPT_IPMASQ

[Install]
WantedBy=multi-user.target
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

func (provisioner *CoreOSProvisioner) Package(name string, action pkgaction.PackageAction) error {
	return nil
}

func (provisioner *CoreOSProvisioner) Provision(swarmOptions swarm.SwarmOptions, authOptions auth.AuthOptions, engineOptions engine.EngineOptions) error {
	provisioner.SwarmOptions = swarmOptions
	provisioner.AuthOptions = authOptions
	provisioner.EngineOptions = engineOptions

	if err := provisioner.SetHostname(provisioner.Driver.GetMachineName()); err != nil {
		return err
	}

	if err := provisioner.SetCloudInit(provisioner.Driver.GetMachineName()); err != nil {
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

	return nil
}
