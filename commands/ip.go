package commands

func cmdIP(c CommandLine) error {
	return runActionWithContext("ip", c)
}
