package xhyve

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/docker/machine/log"
)

var (
	reVMNameUUID      = regexp.MustCompile(`"(.+)" {([0-9a-f-]+)}`)
	reVMInfoLine      = regexp.MustCompile(`(?:"(.+)"|(.+))=(?:"(.*)"|(.*))`)
	reColonLine       = regexp.MustCompile(`(.+):\s+(.*)`)
	reEqualLine       = regexp.MustCompile(`(.+)=(.*)`)
	reEqualQuoteLine  = regexp.MustCompile(`"(.+)"="(.*)"`)
	reMachineNotFound = regexp.MustCompile(`Could not find a registered machine named '(.+)'`)
)

var (
	ErrMachineExist    = errors.New("machine already exists")
	ErrMachineNotExist = errors.New("machine does not exist")
	ErrXhyveNotFound   = errors.New("xhyve not found")
	XhyveCmd           = setXhyveCmd()
)

// Detect the xhyve cmd's path if needed
func setXhyveCmd() string { // TODO
	cmd := "xhyve"
	if path, err := exec.LookPath(cmd); err == nil {
		return path
	}
	return cmd
}

func xhyve(args ...string) (string, error) { // TODO
	var Password string
	cmd := exec.Command("sudo", "-S", XhyveCmd)
	cmd.Stdin = strings.NewReader(Password)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	log.Debugf("executing: %v %v %v", cmd, args, stdout.String())

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	log.Debug(stdout.String())
	return stdout.String(), err
}

func xhyveOut(args ...string) (string, error) { // TODO
	stdout, _, err := xhyveOutErr(args...)
	return stdout, err
}

func xhyveOutErr(args ...string) (string, string, error) { // TODO
	cmd := exec.Command(XhyveCmd, args...)
	log.Debugf("executing: %v %v", XhyveCmd, strings.Join(args, " "))
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	stderrStr := stderr.String()
	log.Debugf("STDOUT: %v", stdout.String())
	log.Debugf("STDERR: %v", stderrStr)
	if err != nil {
		if ee, ok := err.(*exec.Error); ok && ee == exec.ErrNotFound {
			err = ErrXhyveNotFound
		}
	} else {
		if strings.Contains(stderrStr, "error:") {
			err = fmt.Errorf("%v %v failed: %v", XhyveCmd, strings.Join(args, " "), stderrStr)
		}
	}
	return stdout.String(), stderrStr, err
}
