/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package vmwarevsphere

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/docker/machine/drivers/vmwarevsphere/errors"
	"github.com/docker/machine/libmachine/log"
)

type VcConn struct {
	driver *Driver
}

func NewVcConn(driver *Driver) VcConn {
	return VcConn{driver: driver}
}

func (conn VcConn) DatastoreLs(path string) (string, error) {
	args := []string{"datastore.ls"}
	args = conn.AppendConnectionString(args)
	args = append(args, fmt.Sprintf("--ds=%s", conn.driver.Datastore))
	args = append(args, fmt.Sprintf("--dc=%s", conn.driver.Datacenter))
	args = append(args, path)
	stdout, stderr, err := govcOutErr(args...)
	if stderr == "" && err == nil {
		return stdout, nil
	}
	return "", errors.NewDatastoreError(conn.driver.Datastore, "ls", stderr)
}

func (conn VcConn) DatastoreMkdir(dirName string) error {
	_, err := conn.DatastoreLs(dirName)
	if err == nil {
		return nil
	}

	log.Infof("Creating directory %s on datastore %s of vCenter %s... ",
		dirName, conn.driver.Datastore, conn.driver.IP)

	args := []string{"datastore.mkdir"}
	args = conn.AppendConnectionString(args)
	args = append(args, fmt.Sprintf("--ds=%s", conn.driver.Datastore))
	args = append(args, fmt.Sprintf("--dc=%s", conn.driver.Datacenter))
	args = append(args, dirName)
	_, stderr, err := govcOutErr(args...)

	if stderr != "" {
		return errors.NewDatastoreError(conn.driver.Datastore, "mkdir", stderr)
	}

	if err != nil {
		return err
	}

	return nil
}

func (conn VcConn) DatastoreUpload(localPath, destination string) error {
	stdout, err := conn.DatastoreLs(destination)
	if err == nil && strings.Contains(stdout, B2DISOName) {
		log.Infof("boot2docker ISO already uploaded, skipping upload... ")
		return nil
	}

	log.Infof("Uploading %s to %s on datastore %s of vCenter %s... ",
		localPath, destination, conn.driver.Datastore, conn.driver.IP)

	dsPath := fmt.Sprintf("%s/%s", destination, B2DISOName)
	args := []string{"datastore.upload"}
	args = conn.AppendConnectionString(args)
	args = append(args, fmt.Sprintf("--ds=%s", conn.driver.Datastore))
	args = append(args, fmt.Sprintf("--dc=%s", conn.driver.Datacenter))
	args = append(args, localPath)
	args = append(args, dsPath)
	_, stderr, err := govcOutErr(args...)
	if stderr == "" && err == nil {
		return nil
	}
	return errors.NewDatastoreError(conn.driver.Datacenter, "upload", stderr)
}

func (conn VcConn) VMInfo() (string, error) {
	args := []string{"vm.info"}
	args = conn.AppendConnectionString(args)
	args = append(args, fmt.Sprintf("--dc=%s", conn.driver.Datacenter))
	args = append(args, conn.driver.MachineName)

	stdout, stderr, err := govcOutErr(args...)
	if strings.Contains(stdout, "Name") && stderr == "" && err == nil {
		return stdout, nil
	}
	return "", errors.NewVMError("find", conn.driver.MachineName, "VM not found")
}

func (conn VcConn) VMCreate(isoPath string) error {
	log.Infof("Creating virtual machine %s of vCenter %s... ",
		conn.driver.MachineName, conn.driver.IP)

	args := []string{"vm.create"}
	args = conn.AppendConnectionString(args)
	args = append(args, fmt.Sprintf("--net=%s", conn.driver.Network))
	args = append(args, fmt.Sprintf("--dc=%s", conn.driver.Datacenter))
	args = append(args, fmt.Sprintf("--ds=%s", conn.driver.Datastore))
	args = append(args, fmt.Sprintf("--iso=%s", isoPath))
	memory := strconv.Itoa(conn.driver.Memory)
	args = append(args, fmt.Sprintf("--m=%s", memory))
	cpu := strconv.Itoa(conn.driver.CPU)
	args = append(args, fmt.Sprintf("--c=%s", cpu))
	args = append(args, "--disk.controller=pvscsi")
	args = append(args, "--net.adapter=vmxnet3")
	args = append(args, "--on=false")
	if conn.driver.Pool != "" {
		args = append(args, fmt.Sprintf("--pool=%s", conn.driver.Pool))
	}
	if conn.driver.HostIP != "" {
		args = append(args, fmt.Sprintf("--host.ip=%s", conn.driver.HostIP))
	}
	args = append(args, conn.driver.MachineName)
	_, stderr, err := govcOutErr(args...)

	if stderr == "" && err == nil {
		return nil
	}
	return errors.NewVMError("create", conn.driver.MachineName, stderr)
}

func (conn VcConn) VMPowerOn() error {
	log.Infof("Powering on virtual machine %s of vCenter %s... ",
		conn.driver.MachineName, conn.driver.IP)

	args := []string{"vm.power"}
	args = conn.AppendConnectionString(args)
	args = append(args, fmt.Sprintf("--dc=%s", conn.driver.Datacenter))
	args = append(args, "-on")
	args = append(args, conn.driver.MachineName)
	_, stderr, err := govcOutErr(args...)

	if stderr == "" && err == nil {
		return nil
	}
	return errors.NewVMError("power on", conn.driver.MachineName, stderr)
}

func (conn VcConn) VMPowerOff() error {
	log.Infof("Powering off virtual machine %s of vCenter %s... ",
		conn.driver.MachineName, conn.driver.IP)

	args := []string{"vm.power"}
	args = conn.AppendConnectionString(args)
	args = append(args, fmt.Sprintf("--dc=%s", conn.driver.Datacenter))
	args = append(args, "-off")
	args = append(args, conn.driver.MachineName)
	_, stderr, err := govcOutErr(args...)

	if stderr == "" && err == nil {
		return nil
	}
	return errors.NewVMError("power on", conn.driver.MachineName, stderr)
}

func (conn VcConn) VMShutdown() error {
	log.Infof("Powering off virtual machine %s of vCenter %s... ",
		conn.driver.MachineName, conn.driver.IP)

	args := []string{"vm.power"}
	args = conn.AppendConnectionString(args)
	args = append(args, fmt.Sprintf("--dc=%s", conn.driver.Datacenter))
	args = append(args, "-s")
	args = append(args, conn.driver.MachineName)
	_, stderr, err := govcOutErr(args...)

	if stderr == "" && err == nil {
		return nil
	}
	return errors.NewVMError("power on", conn.driver.MachineName, stderr)
}

func (conn VcConn) VMDestroy() error {
	log.Infof("Deleting virtual machine %s of vCenter %s... ",
		conn.driver.MachineName, conn.driver.IP)

	args := []string{"vm.destroy"}
	args = conn.AppendConnectionString(args)
	args = append(args, fmt.Sprintf("--dc=%s", conn.driver.Datacenter))
	args = append(args, conn.driver.MachineName)
	_, stderr, err := govcOutErr(args...)

	if stderr == "" && err == nil {
		return nil
	}
	return errors.NewVMError("delete", conn.driver.MachineName, stderr)
}

func (conn VcConn) VMDiskCreate() error {
	args := []string{"vm.disk.create"}
	args = conn.AppendConnectionString(args)
	args = append(args, fmt.Sprintf("--dc=%s", conn.driver.Datacenter))
	args = append(args, fmt.Sprintf("--vm=%s", conn.driver.MachineName))
	args = append(args, fmt.Sprintf("--ds=%s", conn.driver.Datastore))
	args = append(args, fmt.Sprintf("--name=%s/%s", conn.driver.MachineName, conn.driver.MachineName))
	diskSize := strconv.Itoa(conn.driver.DiskSize)
	args = append(args, fmt.Sprintf("--size=%sMiB", diskSize))

	_, stderr, err := govcOutErr(args...)
	if stderr == "" && err == nil {
		return nil
	}
	return errors.NewVMError("add network", conn.driver.MachineName, stderr)
}

func (conn VcConn) VMAttachNetwork() error {
	args := []string{"vm.network.add"}
	args = conn.AppendConnectionString(args)
	args = append(args, fmt.Sprintf("--dc=%s", conn.driver.Datacenter))
	args = append(args, fmt.Sprintf("--vm=%s", conn.driver.MachineName))
	args = append(args, fmt.Sprintf("--net=%s", conn.driver.Network))

	_, stderr, err := govcOutErr(args...)
	if stderr == "" && err == nil {
		return nil
	}
	return errors.NewVMError("add network", conn.driver.MachineName, stderr)
}

func (conn VcConn) VMFetchIP() (string, error) {
	args := []string{"vm.ip"}
	args = conn.AppendConnectionString(args)
	args = append(args, fmt.Sprintf("--dc=%s", conn.driver.Datacenter))
	args = append(args, conn.driver.MachineName)
	stdout, stderr, err := govcOutErr(args...)

	if stderr == "" && err == nil {
		return stdout, nil
	}
	return "", errors.NewVMError("fetching IP", conn.driver.MachineName, stderr)
}

func (conn VcConn) GuestMkdir(guestUser, guestPass, dirName string) error {
	args := []string{"guest.mkdir"}
	args = conn.AppendConnectionString(args)
	args = append(args, fmt.Sprintf("--dc=%s", conn.driver.Datacenter))
	args = append(args, fmt.Sprintf("--l=%s:%s", guestUser, guestPass))
	args = append(args, fmt.Sprintf("--vm=%s", conn.driver.MachineName))
	args = append(args, "-p")
	args = append(args, dirName)
	_, stderr, err := govcOutErr(args...)

	if stderr == "" && err == nil {
		return nil
	}
	return errors.NewGuestError("mkdir", conn.driver.MachineName, stderr)
}

func (conn VcConn) GuestUpload(guestUser, guestPass, localPath, remotePath string) error {
	args := []string{"guest.upload"}
	args = conn.AppendConnectionString(args)
	args = append(args, fmt.Sprintf("--dc=%s", conn.driver.Datacenter))
	args = append(args, fmt.Sprintf("--l=%s:%s", guestUser, guestPass))
	args = append(args, fmt.Sprintf("--vm=%s", conn.driver.MachineName))
	args = append(args, "-f")
	args = append(args, localPath)
	args = append(args, remotePath)
	_, stderr, err := govcOutErr(args...)

	if stderr == "" && err == nil {
		return nil
	}
	return errors.NewGuestError("upload", conn.driver.MachineName, stderr)
}

func (conn VcConn) GuestStart(guestUser, guestPass, remoteBin, remoteArguments string) error {
	args := []string{"guest.start"}
	args = conn.AppendConnectionString(args)
	args = append(args, fmt.Sprintf("--dc=%s", conn.driver.Datacenter))
	args = append(args, fmt.Sprintf("--l=%s:%s", guestUser, guestPass))
	args = append(args, fmt.Sprintf("--vm=%s", conn.driver.MachineName))
	args = append(args, remoteBin)
	args = append(args, remoteArguments)
	_, stderr, err := govcOutErr(args...)

	if stderr == "" && err == nil {
		return nil
	}
	return errors.NewGuestError("start", conn.driver.MachineName, stderr)
}

func (conn VcConn) GuestDownload(guestUser, guestPass, remotePath, localPath string) error {
	args := []string{"guest.download"}
	args = conn.AppendConnectionString(args)
	args = append(args, fmt.Sprintf("--dc=%s", conn.driver.Datacenter))
	args = append(args, fmt.Sprintf("--l=%s:%s", guestUser, guestPass))
	args = append(args, fmt.Sprintf("--vm=%s", conn.driver.MachineName))
	args = append(args, remotePath)
	args = append(args, localPath)
	_, stderr, err := govcOutErr(args...)

	if stderr == "" && err == nil {
		return nil
	}
	return errors.NewGuestError("download", conn.driver.MachineName, stderr)
}

func (conn VcConn) AppendConnectionString(args []string) []string {
	args = append(args, fmt.Sprintf("--u=%s:%s@%s", conn.driver.Username, conn.driver.Password, conn.driver.IP))
	args = append(args, "--k=true")
	return args
}
