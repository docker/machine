package commands

func cmdVersion(c CommandLine) error {
	c.ShowVersion()
	return nil
}
