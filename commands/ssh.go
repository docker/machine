package commands

import (
	"fmt"
	"strings"

	"github.com/docker/machine/log"
	"github.com/docker/machine/state"

	"github.com/codegangsta/cli"
)

func cmdSsh(c *cli.Context) {
	args := c.Args()
	name := args.First()
	cmd := ""

	if name == "" {
		log.Fatal("Error: Please specify a machine name.")
	}

	certInfo := getCertPathInfo(c)
	defaultStore, err := getDefaultStore(
		c.GlobalString("storage-path"),
		certInfo.CaCertPath,
		certInfo.CaKeyPath,
	)
	if err != nil {
		log.Fatal(err)
	}

	provider, err := newProvider(defaultStore)
	if err != nil {
		log.Fatal(err)
	}

	host, err := provider.Get(name)
	if err != nil {
		log.Fatal(err)
	}

	currentState, err := host.Driver.GetState()
	if err != nil {
		log.Fatal(err)
	}

	if currentState != state.Running {
		log.Fatalf("Error: Cannot run SSH command: Host %q is not running", host.Name)
	}

	// Loop through the arguments and parse out a command which relies on
	// flags if it exists, for instance an invocation of the form
	// `docker-machine ssh dev -- df -h` would mandate this, otherwise we
	// will accidentally trigger the codegangsta/cli help text because it
	// thinks we are trying to specify codegangsta flags.
	//
	// TODO: I thought codegangsta/cli supported the flag parsing
	// terminator manually, which would mitigate the need for this kind of
	// hack.  We should investigate.
	for i, arg := range args {
		if arg == "--" {
			cmd = strings.Join(args[i+1:], " ")
			break
		}
	}

	// It is possible that the user has specified an appended command which
	// does not rely on the flag parsing terminator, such as
	// `docker-machine ssh dev ls`, so this block accounts for that case.
	if len(cmd) == 0 {
		cmd = strings.Join(args[1:], " ")
	}

	if len(c.Args()) == 1 {
		err := host.CreateSSHShell()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		output, err := host.RunSSHCommand(cmd)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Print(output)
	}

}
