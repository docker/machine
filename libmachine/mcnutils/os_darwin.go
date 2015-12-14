package mcnutils

import "os/exec"

func LocalOSVersion() string {
	command := exec.Command("bash", "-c", `sw_vers | grep ProductVersion | cut -d$'\t' -f2`)
	output, err := command.Output()
	if err != nil {
		return ""
	}
	return string(output)
}
