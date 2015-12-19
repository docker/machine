package commands

import (
	"fmt"

	"io"
	"os"

	"github.com/docker/machine/libmachine"
)

func cmdVersion(c CommandLine, api libmachine.API) error {
	return printVersion(c, api, os.Stdout)
}

func printVersion(c CommandLine, api libmachine.API, out io.Writer) error {
	if len(c.Args()) == 0 {
		c.ShowVersion()
		return nil
	}

	if len(c.Args()) != 1 {
		return ErrExpectedOneMachine
	}

	host, err := api.Load(c.Args().First())
	if err != nil {
		return err
	}
	defer api.Close(host)

	version, err := host.DockerVersion()
	if err != nil {
		return err
	}

	fmt.Fprintln(out, version)

	return nil
}
