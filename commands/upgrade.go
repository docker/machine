package commands

func cmdUpgrade(c CommandLine) error {
	return runActionWithContext("upgrade", c)
}
