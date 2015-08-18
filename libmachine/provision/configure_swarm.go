package provision

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"
	"text/template"

	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/swarm"
)

type SwarmCommandContext struct {
	ContainerName string
	DockerDir     string
	DockerPort    int
	Ip            string
	Port          string
	AuthOptions   auth.AuthOptions
	SwarmOptions  swarm.SwarmOptions
	SwarmImage    string
}

// Wrapper function to generate a docker run swarm command (manage or join)
// from a template/context and execute it.
func runSwarmCommandFromTemplate(p Provisioner, cmdTmpl string, swarmCmdContext SwarmCommandContext) error {
	var (
		executedCmdTmpl bytes.Buffer
	)

	parsedMasterCmdTemplate, err := template.New("swarmMasterCmd").Parse(cmdTmpl)
	if err != nil {
		return err
	}

	parsedMasterCmdTemplate.Execute(&executedCmdTmpl, swarmCmdContext)

	log.Debugf("The swarm command being run is: %s", executedCmdTmpl.String())

	if _, err := p.SSHCommand(executedCmdTmpl.String()); err != nil {
		return err
	}

	return nil
}

func configureSwarm(p Provisioner, swarmOptions swarm.SwarmOptions, authOptions auth.AuthOptions) error {
	if !swarmOptions.IsSwarm {
		return nil
	}

	log.Info("Configuring swarm...")

	ip, err := p.GetDriver().GetIP()
	if err != nil {
		return err
	}

	u, err := url.Parse(swarmOptions.Host)
	if err != nil {
		return err
	}

	parts := strings.Split(u.Host, ":")
	port := parts[1]

	dockerDir := p.GetDockerOptionsDir()

	swarmCmdContext := SwarmCommandContext{
		ContainerName: "",
		DockerDir:     dockerDir,
		DockerPort:    2376,
		Ip:            ip,
		Port:          port,
		AuthOptions:   authOptions,
		SwarmOptions:  swarmOptions,
		SwarmImage:    swarmOptions.Image,
	}

	// First things first, get the swarm image.
	if _, err := p.SSHCommand(fmt.Sprintf("sudo docker pull %s", swarmOptions.Image)); err != nil {
		return err
	}

	swarmMasterCmdTemplate := `sudo docker run -d \
--restart=always \
--name swarm-agent-master \
-p {{.Port}}:{{.Port}} \
-v {{.DockerDir}}:{{.DockerDir}} \
{{.SwarmImage}} \
manage \
--tlsverify \
--tlscacert={{.AuthOptions.CaCertRemotePath}} \
--tlscert={{.AuthOptions.ServerCertRemotePath}} \
--tlskey={{.AuthOptions.ServerKeyRemotePath}} \
-H {{.SwarmOptions.Host}} \
--strategy {{.SwarmOptions.Strategy}} {{range .SwarmOptions.ArbitraryFlags}} --{{.}}{{end}} {{.SwarmOptions.Discovery}}
`

	swarmWorkerCmdTemplate := `sudo docker run -d \
--restart=always \
--name swarm-agent \
{{.SwarmImage}} \
join --advertise {{.Ip}}:{{.DockerPort}} {{.SwarmOptions.Discovery}}
`

	if swarmOptions.Master {
		log.Debug("Launching swarm master")
		if err := runSwarmCommandFromTemplate(p, swarmMasterCmdTemplate, swarmCmdContext); err != nil {
			return err
		}
	}

	log.Debug("Launch swarm worker")
	if err := runSwarmCommandFromTemplate(p, swarmWorkerCmdTemplate, swarmCmdContext); err != nil {
		return err
	}

	return nil
}
