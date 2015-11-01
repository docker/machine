package hyperv

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/docker/machine/libmachine/log"
)

var powershell string

func init() {
	systemPath := strings.Split(os.Getenv("PATH"), ";")
	for _, path := range systemPath {
		if strings.Index(path, "WindowsPowerShell") != -1 {
			powershell = filepath.Join(path, "powershell.exe")
		}
	}
}

func execute(args []string) (string, error) {
	args = append([]string{"-NoProfile"}, args...)
	cmd := exec.Command(powershell, args...)
	log.Debugf("[executing ==>] : %v %v", powershell, strings.Join(args, " "))
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	log.Debugf("[stdout =====>] : %s", stdout.String())
	log.Debugf("[stderr =====>] : %s", stderr.String())
	return stdout.String(), err
}

func parseStdout(stdout string) []string {
	s := bufio.NewScanner(strings.NewReader(stdout))
	resp := []string{}
	for s.Scan() {
		resp = append(resp, s.Text())
	}
	return resp
}

func hypervAvailable() error {
	command := []string{
		"@(Get-Command Get-VM).ModuleName"}
	stdout, err := execute(command)
	if err != nil {
		return err
	}
	resp := parseStdout(stdout)

	if resp[0] == "Hyper-V" {
		return nil
	}
	return fmt.Errorf("Hyper-V PowerShell Module is not available")
}
