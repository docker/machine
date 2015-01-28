package parallels

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	log "github.com/Sirupsen/logrus"
)

// TODO check these
var (
	reVMNameUUID      = regexp.MustCompile(`"(.+)" {([0-9a-f-]+)}`)
	reVMInfoLine      = regexp.MustCompile(`(?:"(.+)"|(.+))=(?:"(.*)"|(.*))`)
	reColonLine       = regexp.MustCompile(`(.+):\s+(.*)`)
	reMachineNotFound = regexp.MustCompile(`Failed to get VM config: The virtual machine could not be found..*`)
)

var (
	ErrMachineExist    = errors.New("machine already exists")
	ErrMachineNotExist = errors.New("machine does not exist")
	ErrPrlctlNotFound  = errors.New("prlctl not found")
	prlctlCmd          = "prlctl"
)

func prlctl(args ...string) error {
	cmd := exec.Command(prlctlCmd, args...)
	if os.Getenv("DEBUG") != "" {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	log.Debugf("executing: %v %v", prlctlCmd, strings.Join(args, " "))
	if err := cmd.Run(); err != nil {
		if ee, ok := err.(*exec.Error); ok && ee == exec.ErrNotFound {
			return ErrPrlctlNotFound
		}
		return fmt.Errorf("%v %v failed: %v", prlctlCmd, strings.Join(args, " "), err)
	}
	return nil
}

func prlctlOut(args ...string) (string, error) {
	cmd := exec.Command(prlctlCmd, args...)
	if os.Getenv("DEBUG") != "" {
		cmd.Stderr = os.Stderr
	}
	log.Debugf("executing: %v %v", prlctlCmd, strings.Join(args, " "))

	b, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.Error); ok && ee == exec.ErrNotFound {
			err = ErrPrlctlNotFound
		}
	}
	return string(b), err
}

func prlctlOutErr(args ...string) (string, string, error) {
	cmd := exec.Command(prlctlCmd, args...)
	log.Debugf("executing: %v %v", prlctlCmd, strings.Join(args, " "))
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		if ee, ok := err.(*exec.Error); ok && ee == exec.ErrNotFound {
			err = ErrPrlctlNotFound
		}
	}
	return stdout.String(), stderr.String(), err
}
