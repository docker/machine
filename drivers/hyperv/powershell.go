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

func cmdOut(args ...string) (string, error) {
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

func cmd(args ...string) error {
	_, err := cmdOut(args...)
	return err
}

func parseLines(stdout string) []string {
	resp := []string{}

	s := bufio.NewScanner(strings.NewReader(stdout))
	for s.Scan() {
		resp = append(resp, s.Text())
	}

	return resp
}

func hypervAvailable() error {
	stdout, err := cmdOut("@(Get-Command Get-VM).ModuleName")
	if err != nil {
		return err
	}

	resp := parseLines(stdout)
	if resp[0] != "Hyper-V" {
		return fmt.Errorf("Hyper-V PowerShell Module is not available")
	}

	return nil
}
