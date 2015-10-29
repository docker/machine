package commands

func cmdStop(c CommandLine) error {
	return runActionWithContext("stop", c)
}
