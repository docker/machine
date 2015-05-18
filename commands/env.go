package commands

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"text/template"

	"github.com/docker/machine/log"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/utils"
)

const (
	envTmpl = `{{ .Prefix }}DOCKER_TLS_VERIFY{{ .Delimiter }}{{ .DockerTLSVerify }}{{ .Suffix }}{{ .Prefix }}DOCKER_HOST{{ .Delimiter }}{{ .DockerHost }}{{ .Suffix }}{{ .Prefix }}DOCKER_CERT_PATH{{ .Delimiter }}{{ .DockerCertPath }}{{ .Suffix }}{{ .UsageHint }}`
)

var (
	improperEnvArgsError = errors.New("Error: Expected either one machine name, or -u flag to unset the variables in the arguments.")
)

type ShellConfig struct {
	Prefix          string
	Delimiter       string
	Suffix          string
	DockerCertPath  string
	DockerHost      string
	DockerTLSVerify string
	UsageHint       string
}

func cmdEnv(c *cli.Context) {
	if len(c.Args()) != 1 && !c.Bool("unset") {
		log.Fatalln(improperEnvArgsError)
	}
	userShell := c.String("shell")
	if userShell == "" {
		shell, err := detectShell()
		if err != nil {
			log.Fatalln(err)
		}
		userShell = shell
	}

	t := template.New("envConfig")

	usageHint := generateUsageHint(c.App.Name, c.Args().First(), userShell)

	shellCfg := ShellConfig{
		DockerCertPath:  "",
		DockerHost:      "",
		DockerTLSVerify: "",
	}

	// unset vars
	if c.Bool("unset") {
		switch userShell {
		case "fish":
			shellCfg.Prefix = "set -e "
			shellCfg.Delimiter = ""
			shellCfg.Suffix = ";\n"
		case "powershell":
			shellCfg.Prefix = "Remove-Item Env:\\\\"
			shellCfg.Delimiter = ""
			shellCfg.Suffix = "\n"
		case "cmd":
			// since there is no way to unset vars in cmd just reset to empty
			shellCfg.DockerCertPath = ""
			shellCfg.DockerHost = ""
			shellCfg.DockerTLSVerify = ""
			shellCfg.Prefix = "set "
			shellCfg.Delimiter = "="
			shellCfg.Suffix = "\n"
		default:
			shellCfg.Prefix = "unset "
			shellCfg.Delimiter = " "
			shellCfg.Suffix = "\n"
		}

		tmpl, err := t.Parse(envTmpl)
		if err != nil {
			log.Fatalln(err)
		}

		if err := tmpl.Execute(os.Stdout, shellCfg); err != nil {
			log.Fatalln(err)
		}
		return
	}

	cfg, err := getMachineConfig(c)
	if err != nil {
		log.Fatalln(err)
	}

	if cfg.machineUrl == "" {
		log.Fatalf("%s is not running. Please start this with %s start %s\n", cfg.machineName, c.App.Name, cfg.machineName)
	}

	dockerHost := cfg.machineUrl
	if c.Bool("swarm") {
		if !cfg.SwarmOptions.Master {
			log.Fatalf("%s is not a swarm master\n", cfg.machineName)
		}
		u, err := url.Parse(cfg.SwarmOptions.Host)
		if err != nil {
			log.Fatalln(err)
		}
		parts := strings.Split(u.Host, ":")
		swarmPort := parts[1]

		// get IP of machine to replace in case swarm host is 0.0.0.0
		mUrl, err := url.Parse(cfg.machineUrl)
		if err != nil {
			log.Fatalln(err)
		}
		mParts := strings.Split(mUrl.Host, ":")
		machineIp := mParts[0]

		dockerHost = fmt.Sprintf("tcp://%s:%s", machineIp, swarmPort)
	}

	u, err := url.Parse(cfg.machineUrl)
	if err != nil {
		log.Fatalln(err)
	}

	if u.Scheme != "unix" {
		// validate cert and regenerate if needed
		valid, err := utils.ValidateCertificate(
			u.Host,
			cfg.caCertPath,
			cfg.serverCertPath,
			cfg.serverKeyPath,
		)
		if err != nil {
			log.Fatalln(err)
		}

		if !valid {
			log.Debugf("invalid certs detected; regenerating for %s\n", u.Host)

			if err := runActionWithContext("configureAuth", c); err != nil {
				log.Fatalln(err)
			}
		}
	}

	shellCfg = ShellConfig{
		DockerCertPath:  cfg.machineDir,
		DockerHost:      dockerHost,
		DockerTLSVerify: "1",
		UsageHint:       usageHint,
	}

	switch userShell {
	case "fish":
		shellCfg.Prefix = "set -x "
		shellCfg.Suffix = "\";\n"
		shellCfg.Delimiter = " \""
	case "powershell":
		shellCfg.Prefix = "$Env:"
		shellCfg.Suffix = "\"\n"
		shellCfg.Delimiter = " = \""
	case "cmd":
		shellCfg.Prefix = "set "
		shellCfg.Suffix = "\n"
		shellCfg.Delimiter = "="
	default:
		shellCfg.Prefix = "export "
		shellCfg.Suffix = "\"\n"
		shellCfg.Delimiter = "=\""
	}

	tmpl, err := t.Parse(envTmpl)
	if err != nil {
		log.Fatalln(err)
	}

	if err := tmpl.Execute(os.Stdout, shellCfg); err != nil {
		log.Fatalln(err)
	}
}

func generateUsageHint(appName, machineName, userShell string) string {
	cmd := ""
	switch userShell {
	case "fish":
		if machineName != "" {
			cmd = fmt.Sprintf("eval (%s env %s)", appName, machineName)
		} else {
			cmd = fmt.Sprintf("eval (%s env)", appName)
		}
	case "powershell":
		if machineName != "" {
			cmd = fmt.Sprintf("%s env --shell=powershell %s | Invoke-Expression", appName, machineName)
		} else {
			cmd = fmt.Sprintf("%s env --shell=powershell | Invoke-Expression", appName)
		}
	case "cmd":
		cmd = "copy and paste the above values into your command prompt"
	default:
		if machineName != "" {
			cmd = fmt.Sprintf("eval \"$(%s env %s)\"", appName, machineName)
		} else {
			cmd = fmt.Sprintf("eval \"$(%s env)\"", appName)
		}
	}

	return fmt.Sprintf("# Run this command to configure your shell: \n# %s\n", cmd)
}
