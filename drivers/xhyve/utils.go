package xhyve

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/machine/log"
)

var (
	//	ErrMachineExist     = errors.New("machine already exists")
	//	ErrMachineNotExist  = errors.New("machine does not exist")
	ErrDdNotFound       = errors.New("xhyve not found")
	ErrUuidgenNotFound  = errors.New("uuidgen not found")
	//	ErrUuid2macNotFound = errors.New("uuid2mac not found")
	ErrHdiutilNotFound  = errors.New("hdiutil not found")
)

func dd(input string, output string, block string, count int) (string, string, error) { // TODO
	cmd := exec.Command("dd", fmt.Sprintf("if=%s", input), fmt.Sprintf("of=%s", output), fmt.Sprintf("bs=%s", block), fmt.Sprintf("count=%d", count))
	if os.Getenv("DEBUG") != "" {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr

	err := cmd.Run()
	if err != nil {
		if ee, ok := err.(*exec.Error); ok && ee == exec.ErrNotFound {
			err = ErrDdNotFound
		}
	}

	fmt.Println(cmd)
	return stdout.String(), stderr.String(), err
}

func uuidgen() string {
	cmd := exec.Command("uuidgen")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	log.Debugf("executing: %v", cmd)

	err := cmd.Run()
	if err != nil {
		if ee, ok := err.(*exec.Error); ok && ee == exec.ErrNotFound {
			err = ErrUuidgenNotFound
		}
	}

	out := stdout.String()
	out = strings.Replace(out, "\n", "", -1)
	return out
}

func hdiutil(args ...string) error {
	cmd := exec.Command("hdiutil", args...)

	log.Debugf("executing: %v %v", cmd, strings.Join(args, " "))

	err := cmd.Run()
	if err != nil {
		log.Error(err)
	}

	return nil
}
