package main

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/utils"
)

var (
	flVersion = flag.Bool([]string{"v", "-version"}, false, "Print version information and quit")
	flDebug   = flag.Bool([]string{"D", "-debug"}, false, "Enable debug mode")
)

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, "Usage: machine [OPTIONS] COMMAND [arg...]\n\nCreate and manage machines running Docker.\n\nOptions:\n")

		flag.PrintDefaults()

		help := "\nCommands:\n"

		for _, command := range [][]string{
			{"active", "Get or set the active machine"},
			{"create", "Create a machine"},
			{"inspect", "Inspect information about a machine"},
			{"ip", "Get the IP address of a machine"},
			{"kill", "Kill a machine"},
			{"ls", "List machines"},
			{"restart", "Restart a machine"},
			{"rm", "Remove a machine"},
			{"ssh", "Log into or run a command on a machine with SSH"},
			{"start", "Start a machine"},
			{"stop", "Stop a machine"},
			{"upgrade", "Upgrade a machine to the latest version of Docker"},
			{"url", "Get the URL of a machine"},
			{"export", "Export a machine"},
			{"import", "Import a machine"},
		} {
			help += fmt.Sprintf("    %-10.10s%s\n", command[0], command[1])
		}
		help += "\nRun 'machine COMMAND --help' for more information on a command."
		fmt.Fprintf(os.Stderr, "%s\n", help)
	}

	flag.Parse()

	// -D, --debug, -l/--log-level=debug processing
	// When/if -D is removed this block can be deleted
	if *flDebug {
		os.Setenv("DEBUG", "1")
		initLogging(log.DebugLevel)
	}

	cli := &DockerCli{}

	if err := cli.Cmd(flag.Args()...); err != nil {
		if sterr, ok := err.(*utils.StatusError); ok {
			if sterr.Status != "" {
				log.Println(sterr.Status)
			}
			os.Exit(sterr.StatusCode)
		}
		log.Fatal(err)
	}
}
