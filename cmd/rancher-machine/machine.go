package main

import (
	"fmt"
	"os"
	"strconv"

	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/rancher/machine/commands"
	"github.com/rancher/machine/commands/mcndirs"
	"github.com/rancher/machine/drivers/amazonec2"
	"github.com/rancher/machine/drivers/azure"
	"github.com/rancher/machine/drivers/digitalocean"
	"github.com/rancher/machine/drivers/exoscale"
	"github.com/rancher/machine/drivers/generic"
	"github.com/rancher/machine/drivers/google"
	"github.com/rancher/machine/drivers/hyperv"
	"github.com/rancher/machine/drivers/none"
	"github.com/rancher/machine/drivers/openstack"
	"github.com/rancher/machine/drivers/rackspace"
	"github.com/rancher/machine/drivers/softlayer"
	"github.com/rancher/machine/drivers/virtualbox"
	"github.com/rancher/machine/drivers/vmwarefusion"
	"github.com/rancher/machine/drivers/vmwarevcloudair"
	"github.com/rancher/machine/drivers/vmwarevsphere"
	"github.com/rancher/machine/libmachine/drivers/plugin"
	"github.com/rancher/machine/libmachine/drivers/plugin/localbinary"
	"github.com/rancher/machine/libmachine/log"
	"github.com/rancher/machine/version"
)

var AppHelpTemplate = `Usage: {{.Name}} {{if .Flags}}[OPTIONS] {{end}}COMMAND [arg...]

{{.Usage}}

Version: {{.Version}}{{if or .Author .Email}}

Author:{{if .Author}}
  {{.Author}}{{if .Email}} - <{{.Email}}>{{end}}{{else}}
  {{.Email}}{{end}}{{end}}
{{if .Flags}}
Options:
  {{range .Flags}}{{.}}
  {{end}}{{end}}
Commands:
  {{range .Commands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}
  {{end}}
Run '{{.Name}} COMMAND --help' for more information on a command.
`

var CommandHelpTemplate = `Usage: docker-machine {{.Name}}{{if .Flags}} [OPTIONS]{{end}} [arg...]

{{.Usage}}{{if .Description}}

Description:
   {{.Description}}{{end}}{{if .Flags}}

Options:
   {{range .Flags}}
   {{.}}{{end}}{{ end }}
`

func setDebugOutputLevel() {
	// TODO: I'm not really a fan of this method and really would rather
	// use -v / --verbose TBQH
	for _, f := range os.Args {
		if f == "-D" || f == "--debug" || f == "-debug" {
			log.SetDebug(true)
		}
	}

	debugEnv := os.Getenv("MACHINE_DEBUG")
	if debugEnv != "" {
		showDebug, err := strconv.ParseBool(debugEnv)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing boolean value from MACHINE_DEBUG: %s\n", err)
			os.Exit(1)
		}
		log.SetDebug(showDebug)
	}
}

func main() {
	if os.Getenv(localbinary.PluginEnvKey) == localbinary.PluginEnvVal {
		driverName := os.Getenv(localbinary.PluginEnvDriverName)
		runDriver(driverName)
		return
	}

	localbinary.CurrentBinaryIsDockerMachine = true

	setDebugOutputLevel()
	cli.AppHelpTemplate = AppHelpTemplate
	cli.CommandHelpTemplate = CommandHelpTemplate
	app := cli.NewApp()
	app.Name = filepath.Base(os.Args[0])
	app.Author = "Docker Machine Contributors"
	app.Email = "https://github.com/docker/machine"

	app.Commands = commands.Commands
	app.CommandNotFound = cmdNotFound
	app.Usage = "Create and manage machines running Docker."
	app.Version = version.FullVersion()

	log.Debug("Docker Machine Version: ", app.Version)

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug, D",
			Usage: "Enable debug mode",
		},
		cli.StringFlag{
			EnvVar: "MACHINE_STORAGE_PATH",
			Name:   "storage-path, s",
			Value:  mcndirs.GetBaseDir(),
			Usage:  "Configures storage path",
		},
		cli.StringFlag{
			EnvVar: "MACHINE_TLS_CA_CERT",
			Name:   "tls-ca-cert",
			Usage:  "CA to verify remotes against",
			Value:  "",
		},
		cli.StringFlag{
			EnvVar: "MACHINE_TLS_CA_KEY",
			Name:   "tls-ca-key",
			Usage:  "Private key to generate certificates",
			Value:  "",
		},
		cli.StringFlag{
			EnvVar: "MACHINE_TLS_CLIENT_CERT",
			Name:   "tls-client-cert",
			Usage:  "Client cert to use for TLS",
			Value:  "",
		},
		cli.StringFlag{
			EnvVar: "MACHINE_TLS_CLIENT_KEY",
			Name:   "tls-client-key",
			Usage:  "Private key used in client TLS auth",
			Value:  "",
		},
		cli.StringFlag{
			EnvVar: "MACHINE_GITHUB_API_TOKEN",
			Name:   "github-api-token",
			Usage:  "Token to use for requests to the Github API",
			Value:  "",
		},
		cli.BoolFlag{
			EnvVar: "MACHINE_NATIVE_SSH",
			Name:   "native-ssh",
			Usage:  "Use the native (Go-based) SSH implementation.",
		},
		cli.StringFlag{
			EnvVar: "MACHINE_BUGSNAG_API_TOKEN",
			Name:   "bugsnag-api-token",
			Usage:  "BugSnag API token for crash reporting",
			Value:  "",
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Error(err)
	}
}

func runDriver(driverName string) {
	switch driverName {
	case "amazonec2":
		plugin.RegisterDriver(amazonec2.NewDriver("", ""))
	case "azure":
		plugin.RegisterDriver(azure.NewDriver("", ""))
	case "digitalocean":
		plugin.RegisterDriver(digitalocean.NewDriver("", ""))
	case "exoscale":
		plugin.RegisterDriver(exoscale.NewDriver("", ""))
	case "generic":
		plugin.RegisterDriver(generic.NewDriver("", ""))
	case "google":
		plugin.RegisterDriver(google.NewDriver("", ""))
	case "hyperv":
		plugin.RegisterDriver(hyperv.NewDriver("", ""))
	case "none":
		plugin.RegisterDriver(none.NewDriver("", ""))
	case "openstack":
		plugin.RegisterDriver(openstack.NewDriver("", ""))
	case "rackspace":
		plugin.RegisterDriver(rackspace.NewDriver("", ""))
	case "softlayer":
		plugin.RegisterDriver(softlayer.NewDriver("", ""))
	case "virtualbox":
		plugin.RegisterDriver(virtualbox.NewDriver("", ""))
	case "vmwarefusion":
		plugin.RegisterDriver(vmwarefusion.NewDriver("", ""))
	case "vmwarevcloudair":
		plugin.RegisterDriver(vmwarevcloudair.NewDriver("", ""))
	case "vmwarevsphere":
		plugin.RegisterDriver(vmwarevsphere.NewDriver("", ""))
	default:
		fmt.Fprintf(os.Stderr, "Unsupported driver: %s\n", driverName)
		os.Exit(1)
	}
}

func cmdNotFound(c *cli.Context, command string) {
	log.Errorf(
		"%s: '%s' is not a %s command. See '%s --help'.",
		c.App.Name,
		command,
		c.App.Name,
		os.Args[0],
	)
	os.Exit(1)
}
