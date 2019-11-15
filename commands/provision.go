package commands

import "github.com/rancher/machine/libmachine"

func cmdProvision(c CommandLine, api libmachine.API) error {
	return runAction("provision", c, api)
}
