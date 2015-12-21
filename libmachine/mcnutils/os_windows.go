package mcnutils

import (
	"os/exec"
	"strings"
)

func LocalOSVersion() string {
	command := exec.Command("ver")
	output, err := command.Output()
	if err == nil {
		return parseVerOutput(string(output))
	}

	command = exec.Command("systeminfo")
	output, err = command.Output()
	if err == nil {
		return parseSystemInfoOutput(string(output))
	}

	return ""
}

func parseSystemInfoOutput(output string) string {
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "OS Version") {
			return strings.TrimSpace(line[len("OS Version:"):])
		}
	}
	return ""
}

func parseVerOutput(output string) string {
	return strings.TrimSpace(output)
}
