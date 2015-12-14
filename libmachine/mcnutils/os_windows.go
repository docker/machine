package mcnutils

import (
	"os/exec"
	"strings"
)

func LocalOSVersion() string {
	command := exec.Command(`systeminfo`)
	output, err := command.Output()
	if err != nil {
		return ""
	}
	return parseOutput(string(output))
}

func parseOutput(output string) string {
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "OS Version") {
			return strings.TrimSpace(line[len("OS Version:"):])
		}
	}
	return ""
}
