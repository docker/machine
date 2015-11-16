/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package vmwareworkstation

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"github.com/docker/machine/libmachine/log"
)

var (
	vmrunbin    = setVmwareCmd("vmrun.exe")
	vdiskmanbin = setVmwareCmd("vmware-vdiskmanager.exe")
)

var (
	ErrMachineExist    = errors.New("machine already exists")
	ErrMachineNotExist = errors.New("machine does not exist")
	ErrVMRUNNotFound   = errors.New("VMRUN not found")
)

// This reads the VMware installation path from the Windows registry.
func workstationVMwareRoot() (s string, err error) {
	key := `SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\vmware.exe`
	subkey := "Path"
	s, err = readRegString(syscall.HKEY_LOCAL_MACHINE, key, subkey)
	if err != nil {
		log.Printf(`Unable to read registry key %s\%s`, key, subkey)
		return
	}

	return normalizePath(s), nil
}

// // This reads the VMware DHCP leases path from the Windows registry.
// func workstationDhcpLeasesPathRegistry() (s string, err error) {
// 	key := "SYSTEM\\CurrentControlSet\\services\\VMnetDHCP\\Parameters"
// 	subkey := "LeaseFile"
// 	s, err = readRegString(syscall.HKEY_LOCAL_MACHINE, key, subkey)
// 	if err != nil {
// 		log.Printf(`Unable to read registry key %s\%s`, key, subkey)
// 		return
// 	}

// 	return normalizePath(s), nil
// }

// func workstationDhcpLeasesPath(device string) string {
// 	path, err := workstationDhcpLeasesPathRegistry()
// 	if err != nil {
// 		log.Printf("Error finding leases in registry: %s", err)
// 	} else if _, err := os.Stat(path); err == nil {
// 		return path
// 	}

// 	return findFile("vmnetdhcp.leases", workstationDataFilePaths())
// }

// workstationProgramFilesPaths returns a list of paths that are eligible
// to contain program files we may want just as vmware.exe.
func workstationProgramFilePaths() []string {
	path, err := workstationVMwareRoot()
	if err != nil {
		log.Printf("Error finding VMware root: %s", err)
	}

	paths := make([]string, 0, 5)
	if os.Getenv("VMWARE_HOME") != "" {
		paths = append(paths, os.Getenv("VMWARE_HOME"))
	}

	if path != "" {
		paths = append(paths, path)
	}

	if os.Getenv("ProgramFiles(x86)") != "" {
		paths = append(paths,
			filepath.Join(os.Getenv("ProgramFiles(x86)"), "/VMware/VMware Workstation"))
	}

	if os.Getenv("ProgramFiles") != "" {
		paths = append(paths,
			filepath.Join(os.Getenv("ProgramFiles"), "/VMware/VMware Workstation"))
	}

	return paths
}

// detect the vmrun and vmware-vdiskmanager cmds' path if needed
func setVmwareCmd(cmd string) string {
	if path, err := exec.LookPath(cmd); err == nil {
		return path
	}
	return findFile(cmd, workstationProgramFilePaths())
}

func vmrun(args ...string) (string, string, error) {
	cmd := exec.Command(vmrunbin, args...)
	if os.Getenv("MACHINE_DEBUG") != "" {
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
	if os.Getenv("MACHINE_DEBUG") != "" {
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

func normalizePath(path string) string {
	path = strings.Replace(path, "\\", "/", -1)
	path = strings.Replace(path, "//", "/", -1)
	path = strings.TrimRight(path, "/")
	return path
}

func findFile(file string, paths []string) string {
	for _, path := range paths {
		path = filepath.Join(path, file)
		path = normalizePath(path)
		log.Debugf("Searching for file '%s'", path)

		if _, err := os.Stat(path); err == nil {
			log.Debugf("Found file '%s'", path)
			return path
		}
	}

	log.Printf("File not found: '%s'", file)
	return ""
}

// See http://blog.natefinch.com/2012/11/go-win-stuff.html
//
func readRegString(hive syscall.Handle, subKeyPath, valueName string) (value string, err error) {
	var h syscall.Handle
	err = syscall.RegOpenKeyEx(hive, syscall.StringToUTF16Ptr(subKeyPath), 0, syscall.KEY_READ, &h)
	if err != nil {
		return
	}
	defer syscall.RegCloseKey(h)

	var typ uint32
	var bufSize uint32
	err = syscall.RegQueryValueEx(
		h,
		syscall.StringToUTF16Ptr(valueName),
		nil,
		&typ,
		nil,
		&bufSize)
	if err != nil {
		return
	}

	data := make([]uint16, bufSize/2+1)
	err = syscall.RegQueryValueEx(
		h,
		syscall.StringToUTF16Ptr(valueName),
		nil,
		&typ,
		(*byte)(unsafe.Pointer(&data[0])),
		&bufSize)
	if err != nil {
		return
	}

	return syscall.UTF16ToString(data), nil
}
