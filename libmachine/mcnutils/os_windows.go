package mcnutils

import (
	"os/exec"
	"strings"
)

func LocalOSVersion() string {
	command := exec.Command(`ver`)
	output, err := command.Output()
	if err != nil {
		return ""
	}
	return parseOutput(string(output))
}

func parseOutput(output string) string {
	return strings.TrimSpace(output)
}
