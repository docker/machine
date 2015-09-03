/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package vmwarefusion

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/docker/machine/log"
)

var vmfusionapp string

usr, _ := user.Current()
usr_vmfusionapp := filepath.Join(usr.HomeDir, "Applications/VMware Fusion.app")
if _, err := os.Stat(usr_vmfusionapp); err == nil {
	vmfusionapp = usr_vmfusionapp
} else {
	vmfusionapp = "/Applications/VMware Fusion.app"
}

var (
	vmrunbin    = filepath.Join(vmfusionapp, "Contents/Library/vmrun")
	vdiskmanbin = filepath.Join(vmfusionapp, "Contents/Library/vmware-vdiskmanager")
)

var (
	ErrMachineExist    = errors.New("machine already exists")
	ErrMachineNotExist = errors.New("machine does not exist")
	ErrVMRUNNotFound   = errors.New("VMRUN not found")
)

func vmrun(args ...string) (string, string, error) {
	cmd := exec.Command(vmrunbin, args...)
	if os.Getenv("DEBUG") != "" {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	log.Debugf("executing: %v %v", vmrunbin, strings.Join(args, " "))

	err := cmd.Run()
	if err != nil {
		if ee, ok := err.(*exec.Error); ok && ee == exec.ErrNotFound {
			err = ErrVMRUNNotFound
		}
	}

	return stdout.String(), stderr.String(), err
}

// Make a vmdk disk image with the given size (in MB).
func vdiskmanager(dest string, size int) error {
	cmd := exec.Command(vdiskmanbin, "-c", "-t", "0", "-s", fmt.Sprintf("%dMB", size), "-a", "lsilogic", dest)
	if os.Getenv("DEBUG") != "" {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if stdout := cmd.Run(); stdout != nil {
		if ee, ok := stdout.(*exec.Error); ok && ee == exec.ErrNotFound {
			return ErrVMRUNNotFound
		}
	}
	return nil
}
