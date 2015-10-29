package commands

func cmdKill(c CommandLine) error {
	return runActionWithContext("kill", c)
}
