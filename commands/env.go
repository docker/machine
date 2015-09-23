package commands

import (
	"errors"
	"fmt"
	"os"
	"text/template"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/libmachine/log"
)

const (
	envTmpl = `{{ .Prefix }}DOCKER_TLS_VERIFY{{ .Delimiter }}{{ .DockerTLSVerify }}{{ .Suffix }}{{ .Prefix }}DOCKER_HOST{{ .Delimiter }}{{ .DockerHost }}{{ .Suffix }}{{ .Prefix }}DOCKER_CERT_PATH{{ .Delimiter }}{{ .DockerCertPath }}{{ .Suffix }}{{ .Prefix }}DOCKER_MACHINE_NAME{{ .Delimiter }}{{ .MachineName }}{{ .Suffix }}{{ .UsageHint }}`
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
	MachineName     string
}

func cmdEnv(c *cli.Context) {
	if len(c.Args()) != 1 && !c.Bool("unset") {
		log.Fatal(improperEnvArgsError)
	}

	h := getFirstArgHost(c)

	dockerHost, authOptions, err := runConnectionBoilerplate(h, c)
	if err != nil {
		log.Fatalf("Error running connection boilerplate: %s", err)
	}

	userShell := c.String("shell")
	if userShell == "" {
		shell, err := detectShell()
		if err != nil {
			log.Fatal(err)
		}
		userShell = shell
	}

	t := template.New("envConfig")

	usageHint := generateUsageHint(c.App.Name, c.Args().First(), userShell)

	shellCfg := ShellConfig{
		DockerCertPath:  "",
		DockerHost:      "",
		DockerTLSVerify: "",
		MachineName:     "",
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
			log.Fatal(err)
		}

		if err := tmpl.Execute(os.Stdout, shellCfg); err != nil {
			log.Fatal(err)
		}
		return
	}

	shellCfg = ShellConfig{
		DockerCertPath:  authOptions.CertDir,
		DockerHost:      dockerHost,
		DockerTLSVerify: "1",
		UsageHint:       usageHint,
		MachineName:     h.Name,
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
		log.Fatal(err)
	}

	if err := tmpl.Execute(os.Stdout, shellCfg); err != nil {
		log.Fatal(err)
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
