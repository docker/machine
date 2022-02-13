package provision

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/provision/pkgaction"
	"github.com/docker/machine/libmachine/provision/serviceaction"
	"github.com/docker/machine/libmachine/swarm"
)

func init() {
	Register("GoogleContainerOptimizedOS", &RegisteredProvisioner{
		New: NewGoogleCOSProvisioner,
	})
}

func NewGoogleCOSProvisioner(d drivers.Driver) Provisioner {
	return &GoogleCOSProvisioner{
		NewSystemdProvisioner("cos", d),
	}
}

type GoogleCOSProvisioner struct {
	SystemdProvisioner
}

func (p *GoogleCOSProvisioner) String() string {
	return "cos"
}

func (p *GoogleCOSProvisioner) Package(_ string, _ pkgaction.PackageAction) error {
	return nil
}

func (p *GoogleCOSProvisioner) GenerateDockerOptions(dockerPort int) (*DockerOptions, error) {
	var (
		engineCfg bytes.Buffer
	)

	driverNameLabel := fmt.Sprintf("provider=%s", p.Driver.DriverName())
	p.EngineOptions.Labels = append(p.EngineOptions.Labels, driverNameLabel)

	engineConfigTmpl := `[Service]
ExecStart=
ExecStart=/usr/bin/dockerd -H tcp://0.0.0.0:{{.DockerPort}} -H unix:///var/run/docker.sock --tlsverify --tlscacert {{.AuthOptions.CaCertRemotePath}} --tlscert {{.AuthOptions.ServerCertRemotePath}} --tlskey {{.AuthOptions.ServerKeyRemotePath}} {{ range .EngineOptions.Labels }}--label {{.}} {{ end }}{{ range .EngineOptions.InsecureRegistry }}--insecure-registry {{.}} {{ end }}{{ range .EngineOptions.RegistryMirror }}--registry-mirror {{.}} {{ end }}{{ range .EngineOptions.ArbitraryFlags }}--{{.}} {{ end }}
Environment={{range .EngineOptions.Env}}{{ printf "%q" . }} {{end}}
`
	t, err := template.New("engineConfig").Parse(engineConfigTmpl)
	if err != nil {
		return nil, err
	}

	engineConfigContext := EngineConfigContext{
		DockerPort:    dockerPort,
		AuthOptions:   p.AuthOptions,
		EngineOptions: p.EngineOptions,
	}

	_ = t.Execute(&engineCfg, engineConfigContext)

	return &DockerOptions{
		EngineOptions:     engineCfg.String(),
		EngineOptionsPath: p.DaemonOptionsFile,
	}, nil
}

func (p *GoogleCOSProvisioner) Provision(swarmOptions swarm.Options, authOptions auth.Options, engineOptions engine.Options) error {
	p.SwarmOptions = swarmOptions
	p.AuthOptions = authOptions
	p.EngineOptions = engineOptions
	swarmOptions.Env = engineOptions.Env

	log.Debug("Setting hostname")
	err := p.SetHostname(p.Driver.GetMachineName())
	if err != nil {
		return err
	}

	log.Debug("Waiting for Docker Daemon")
	err = mcnutils.WaitFor(p.dockerDaemonResponding)
	if err != nil {
		return err
	}

	log.Debugf("Preparing remote auth options")
	p.AuthOptions = setRemoteAuthOptions(p)

	log.Debug("Configuring auth")
	err = ConfigureAuth(p)
	if err != nil {
		return err
	}

	log.Debug("Configuring local firewall")
	err = p.configureLocalFirewall()
	if err != nil {
		return err
	}

	log.Debug("Configuring swarm")
	err = configureSwarm(p, swarmOptions, p.AuthOptions)
	if err != nil {
		return err
	}

	log.Debug("Enabling Docker in systemd")
	return p.Service("docker", serviceaction.Enable)
}

func (p *GoogleCOSProvisioner) dockerDaemonResponding() bool {
	log.Debug("Checking Docker Daemon")

	out, err := p.SSHCommand("sudo docker version")
	if err != nil {
		log.Warnf("Error getting SSH command to check if the daemon is up: %s", err)
		log.Debugf("'sudo docker version' output:\n%s", out)
		return false
	}

	return true
}

func (p *GoogleCOSProvisioner) configureLocalFirewall() error {
	out, err := p.SSHCommand("sudo iptables -A INPUT -p tcp -m tcp --dport 2376 -j ACCEPT")
	if err != nil {
		log.Debugf("Iptables output:\n%s", out)
		return err
	}

	return nil
}
