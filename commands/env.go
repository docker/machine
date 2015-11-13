package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine/drivers/rpc"
	"github.com/docker/machine/libmachine/log"
)

const (
	envTmpl = `{{ .Prefix }}DOCKER_TLS_VERIFY{{ .Delimiter }}{{ .DockerTLSVerify }}{{ .Suffix }}{{ .Prefix }}DOCKER_HOST{{ .Delimiter }}{{ .DockerHost }}{{ .Suffix }}{{ .Prefix }}DOCKER_CERT_PATH{{ .Delimiter }}{{ .DockerCertPath }}{{ .Suffix }}{{ .Prefix }}DOCKER_MACHINE_NAME{{ .Delimiter }}{{ .MachineName }}{{ .Suffix }}{{ if .NoProxyVar }}{{ .Prefix }}{{ .NoProxyVar }}{{ .Delimiter }}{{ .NoProxyValue }}{{ .Suffix }}{{end}}{{ .UsageHint }}`
)

var (
	errImproperEnvArgs      = errors.New("Error: Expected one machine name")
	errImproperUnsetEnvArgs = errors.New("Error: Expected no machine name when the -u flag is present")
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
	NoProxyVar      string
	NoProxyValue    string
}

func cmdEnv(cli CommandLine, store rpcdriver.Store) error {
	// Ensure that log messages always go to stderr when this command is
	// being run (it is intended to be run in a subshell)
	log.SetOutWriter(os.Stderr)

	if cli.Bool("unset") {
		return unset(cli)
	}

	return set(cli, store)
}

func set(cli CommandLine, store rpcdriver.Store) error {
	if len(cli.Args()) != 1 {
		return errImproperEnvArgs
	}

	host, err := store.Load(cli.Args().First())
	if err != nil {
		return err
	}

	dockerHost, _, err := runConnectionBoilerplate(host, cli)
	if err != nil {
		return fmt.Errorf("Error running connection boilerplate: %s", err)
	}

	userShell, err := getShell(cli)
	if err != nil {
		return err
	}

	shellCfg := &ShellConfig{
		DockerCertPath:  filepath.Join(mcndirs.GetMachineDir(), host.Name),
		DockerHost:      dockerHost,
		DockerTLSVerify: "1",
		UsageHint:       generateUsageHint(userShell, os.Args),
		MachineName:     host.Name,
	}

	if cli.Bool("no-proxy") {
		ip, err := host.Driver.GetIP()
		if err != nil {
			return fmt.Errorf("Error getting host IP: %s", err)
		}

		noProxyVar, noProxyValue := findNoProxyFromEnv()

		// add the docker host to the no_proxy list idempotently
		switch {
		case noProxyValue == "":
			noProxyValue = ip
		case strings.Contains(noProxyValue, ip):
		//ip already in no_proxy list, nothing to do
		default:
			noProxyValue = fmt.Sprintf("%s,%s", noProxyValue, ip)
		}

		shellCfg.NoProxyVar = noProxyVar
		shellCfg.NoProxyValue = noProxyValue
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
		shellCfg.Prefix = "SET "
		shellCfg.Suffix = "\n"
		shellCfg.Delimiter = "="
	default:
		shellCfg.Prefix = "export "
		shellCfg.Suffix = "\"\n"
		shellCfg.Delimiter = "=\""
	}

	return executeTemplateStdout(shellCfg)
}

func unset(cli CommandLine) error {
	if len(cli.Args()) != 0 {
		return errImproperUnsetEnvArgs
	}

	userShell, err := getShell(cli)
	if err != nil {
		return err
	}

	shellCfg := &ShellConfig{
		UsageHint: generateUsageHint(userShell, os.Args),
	}

	if cli.Bool("no-proxy") {
		shellCfg.NoProxyVar, shellCfg.NoProxyValue = findNoProxyFromEnv()
	}

	switch userShell {
	case "fish":
		shellCfg.Prefix = "set -e "
		shellCfg.Suffix = ";\n"
		shellCfg.Delimiter = ""
	case "powershell":
		shellCfg.Prefix = `Remove-Item Env:\\`
		shellCfg.Suffix = "\n"
		shellCfg.Delimiter = ""
	case "cmd":
		shellCfg.Prefix = "SET "
		shellCfg.Suffix = "\n"
		shellCfg.Delimiter = "="
	default:
		shellCfg.Prefix = "unset "
		shellCfg.Suffix = "\n"
		shellCfg.Delimiter = ""
	}

	return executeTemplateStdout(shellCfg)
}

func executeTemplateStdout(shellCfg *ShellConfig) error {
	t := template.New("envConfig")
	tmpl, err := t.Parse(envTmpl)
	if err != nil {
		return err
	}

	return tmpl.Execute(os.Stdout, shellCfg)
}

func getShell(cli CommandLine) (string, error) {
	userShell := cli.String("shell")
	if userShell != "" {
		return userShell, nil
	}
	return detectShell()
}

func findNoProxyFromEnv() (string, string) {
	// first check for an existing lower case no_proxy var
	noProxyVar := "no_proxy"
	noProxyValue := os.Getenv("no_proxy")

	// otherwise default to allcaps HTTP_PROXY
	if noProxyValue == "" {
		noProxyVar = "NO_PROXY"
		noProxyValue = os.Getenv("NO_PROXY")
	}
	return noProxyVar, noProxyValue
}

func generateUsageHint(userShell string, args []string) string {
	cmd := ""
	comment := "#"

	commandLine := strings.Join(args, " ")

	switch userShell {
	case "fish":
		cmd = fmt.Sprintf("eval (%s)", commandLine)
	case "powershell":
		cmd = fmt.Sprintf("%s | Invoke-Expression", commandLine)
	case "cmd":
		cmd = fmt.Sprintf("\tFOR /f \"tokens=*\" %%i IN ('%s') DO %%i", commandLine)
		comment = "REM"
	default:
		cmd = fmt.Sprintf("eval \"$(%s)\"", commandLine)
	}

	return fmt.Sprintf("%s Run this command to configure your shell: \n%s %s\n", comment, comment, cmd)
}

func detectShell() (string, error) {
	// attempt to get the SHELL env var
	shell := filepath.Base(os.Getenv("SHELL"))

	log.Debugf("shell: %s", shell)
	if shell == "" {
		// check for windows env and not bash (i.e. msysgit, etc)
		if runtime.GOOS == "windows" {
			log.Printf("On Windows, please specify either 'cmd' or 'powershell' with the --shell flag.\n\n")
		}

		return "", ErrUnknownShell
	}

	return shell, nil
}
