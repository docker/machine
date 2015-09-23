/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package vmwarevsphere

import (
	"archive/tar"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/machine/libmachine/log"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/drivers/vmwarevsphere/errors"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
)

const (
	isoFilename = "boot2docker.iso"
	B2DISOName  = isoFilename
	B2DUser     = "docker"
	B2DPass     = "tcuser"
)

type Driver struct {
	*drivers.BaseDriver
	CPU            int
	Memory         int
	DiskSize       int
	Boot2DockerURL string
	IP             string
	Username       string
	Password       string
	Network        string
	Datastore      string
	Datacenter     string
	Pool           string
	HostIP         string
	ISO            string
}

const (
	defaultCpus     = 2
	defaultMemory   = 2048
	defaultDiskSize = 20000
)

func init() {
	drivers.Register("vmwarevsphere", &drivers.RegisteredDriver{
		GetCreateFlags: GetCreateFlags,
	})
}

// GetCreateFlags registers the flags this driver adds to
// "docker hosts create"
func GetCreateFlags() []cli.Flag {
	return []cli.Flag{
		cli.IntFlag{
			EnvVar: "VSPHERE_CPU_COUNT",
			Name:   "vmwarevsphere-cpu-count",
			Usage:  "vSphere CPU number for docker VM",
			Value:  defaultCpus,
		},
		cli.IntFlag{
			EnvVar: "VSPHERE_MEMORY_SIZE",
			Name:   "vmwarevsphere-memory-size",
			Usage:  "vSphere size of memory for docker VM (in MB)",
			Value:  defaultMemory,
		},
		cli.IntFlag{
			EnvVar: "VSPHERE_DISK_SIZE",
			Name:   "vmwarevsphere-disk-size",
			Usage:  "vSphere size of disk for docker VM (in MB)",
			Value:  defaultDiskSize,
		},
		cli.StringFlag{
			EnvVar: "VSPHERE_BOOT2DOCKER_URL",
			Name:   "vmwarevsphere-boot2docker-url",
			Usage:  "vSphere URL for boot2docker image",
		},
		cli.StringFlag{
			EnvVar: "VSPHERE_VCENTER",
			Name:   "vmwarevsphere-vcenter",
			Usage:  "vSphere IP/hostname for vCenter",
		},
		cli.StringFlag{
			EnvVar: "VSPHERE_USERNAME",
			Name:   "vmwarevsphere-username",
			Usage:  "vSphere username",
		},
		cli.StringFlag{
			EnvVar: "VSPHERE_PASSWORD",
			Name:   "vmwarevsphere-password",
			Usage:  "vSphere password",
		},
		cli.StringFlag{
			EnvVar: "VSPHERE_NETWORK",
			Name:   "vmwarevsphere-network",
			Usage:  "vSphere network where the docker VM will be attached",
		},
		cli.StringFlag{
			EnvVar: "VSPHERE_DATASTORE",
			Name:   "vmwarevsphere-datastore",
			Usage:  "vSphere datastore for docker VM",
		},
		cli.StringFlag{
			EnvVar: "VSPHERE_DATACENTER",
			Name:   "vmwarevsphere-datacenter",
			Usage:  "vSphere datacenter for docker VM",
		},
		cli.StringFlag{
			EnvVar: "VSPHERE_POOL",
			Name:   "vmwarevsphere-pool",
			Usage:  "vSphere resource pool for docker VM",
		},
		cli.StringFlag{
			EnvVar: "VSPHERE_COMPUTE_IP",
			Name:   "vmwarevsphere-compute-ip",
			Usage:  "vSphere compute host IP where the docker VM will be instantiated",
		},
	}
}

func NewDriver(hostName, storePath string) drivers.Driver {
	return &Driver{
		CPU:      defaultCpus,
		Memory:   defaultMemory,
		DiskSize: defaultDiskSize,
		BaseDriver: &drivers.BaseDriver{
			MachineName: hostName,
			StorePath:   storePath,
		},
	}
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) GetSSHUsername() string {
	if d.SSHUser == "" {
		d.SSHUser = "docker"
	}

	return d.SSHUser
}

func (d *Driver) DriverName() string {
	return "vmwarevsphere"
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.SSHUser = "docker"
	d.SSHPort = 22
	d.CPU = flags.Int("vmwarevsphere-cpu-count")
	d.Memory = flags.Int("vmwarevsphere-memory-size")
	d.DiskSize = flags.Int("vmwarevsphere-disk-size")
	d.Boot2DockerURL = flags.String("vmwarevsphere-boot2docker-url")
	d.IP = flags.String("vmwarevsphere-vcenter")
	d.Username = flags.String("vmwarevsphere-username")
	d.Password = flags.String("vmwarevsphere-password")
	d.Network = flags.String("vmwarevsphere-network")
	d.Datastore = flags.String("vmwarevsphere-datastore")
	d.Datacenter = flags.String("vmwarevsphere-datacenter")
	d.Pool = flags.String("vmwarevsphere-pool")
	d.HostIP = flags.String("vmwarevsphere-compute-ip")
	d.SwarmMaster = flags.Bool("swarm-master")
	d.SwarmHost = flags.String("swarm-host")
	d.SwarmDiscovery = flags.String("swarm-discovery")

	d.ISO = filepath.Join(d.StorePath, isoFilename)

	return nil
}

func (d *Driver) GetURL() (string, error) {
	ip, _ := d.GetIP()
	if ip == "" {
		return "", nil
	}
	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

func (d *Driver) GetIP() (string, error) {
	status, err := d.GetState()
	if status != state.Running {
		return "", errors.NewInvalidStateError(d.MachineName)
	}
	vcConn := NewVcConn(d)
	rawIP, err := vcConn.VMFetchIP()
	if err != nil {
		return "", err
	}
	ip := strings.Trim(strings.Split(rawIP, "\n")[0], " ")
	return ip, nil
}

func (d *Driver) GetState() (state.State, error) {
	vcConn := NewVcConn(d)
	stdout, err := vcConn.VMInfo()
	if err != nil {
		return state.None, err
	}

	if strings.Contains(stdout, "poweredOn") {
		return state.Running, nil
	} else if strings.Contains(stdout, "poweredOff") {
		return state.Stopped, nil
	}
	return state.None, nil
}

func (d *Driver) PreCreateCheck() error {
	return nil
}

// the current implementation does the following:
// 1. check whether the docker directory contains the boot2docker ISO
// 2. generate an SSH keypair and bundle it in a tar.
// 3. create a virtual machine with the boot2docker ISO mounted;
// 4. reconfigure the virtual machine network and disk size;
func (d *Driver) Create() error {
	if err := d.checkVsphereConfig(); err != nil {
		return err
	}

	b2dutils := mcnutils.NewB2dUtils("", "", d.StorePath)
	if err := b2dutils.CopyIsoToMachineDir(d.Boot2DockerURL, d.MachineName); err != nil {
		return err
	}

	log.Infof("Generating SSH Keypair...")
	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return err
	}

	vcConn := NewVcConn(d)
	log.Infof("Uploading Boot2docker ISO ...")
	if err := vcConn.DatastoreMkdir(d.MachineName); err != nil {
		return err
	}

	if _, err := os.Stat(d.ISO); os.IsNotExist(err) {
		log.Errorf("Unable to find boot2docker ISO at %s", d.ISO)
		return errors.NewIncompleteVsphereConfigError(d.ISO)
	}

	if err := vcConn.DatastoreUpload(d.ISO, d.MachineName); err != nil {
		return err
	}

	isoPath := fmt.Sprintf("%s/%s", d.MachineName, isoFilename)
	if err := vcConn.VMCreate(isoPath); err != nil {
		return err
	}

	log.Infof("Configuring the virtual machine %s... ", d.MachineName)
	if err := vcConn.VMDiskCreate(); err != nil {
		return err
	}

	if err := vcConn.VMAttachNetwork(); err != nil {
		return err
	}

	if err := d.Start(); err != nil {
		return err
	}

	// Generate a tar keys bundle
	if err := d.generateKeyBundle(); err != nil {
		return err
	}

	// Copy SSH keys bundle
	if err := vcConn.GuestUpload(B2DUser, B2DPass, d.ResolveStorePath("userdata.tar"), "/home/docker/userdata.tar"); err != nil {
		return err
	}

	// Expand tar file.
	if err := vcConn.GuestStart(B2DUser, B2DPass, "/usr/bin/sudo", "/bin/mv /home/docker/userdata.tar /var/lib/boot2docker/userdata.tar && /usr/bin/sudo tar xf /var/lib/boot2docker/userdata.tar -C /home/docker/ > /var/log/userdata.log 2>&1 && /usr/bin/sudo chown -R docker:staff /home/docker"); err != nil {
		return err
	}

	return nil
}

func (d *Driver) Start() error {
	machineState, err := d.GetState()
	if err != nil {
		return err
	}

	switch machineState {
	case state.Running:
		log.Infof("VM %s has already been started", d.MachineName)
		return nil
	case state.Stopped:
		// TODO add transactional or error handling in the following steps
		vcConn := NewVcConn(d)
		err := vcConn.VMPowerOn()
		if err != nil {
			return err
		}
		// this step waits for the vm to start and fetch its ip address;
		// this guarantees that the opem-vmtools has started working...
		_, err = vcConn.VMFetchIP()
		if err != nil {
			return err
		}

		d.IPAddress, err = d.GetIP()
		return err
	}
	return errors.NewInvalidStateError(d.MachineName)
}

func (d *Driver) Stop() error {
	vcConn := NewVcConn(d)
	if err := vcConn.VMShutdown(); err != nil {
		return err
	}

	d.IPAddress = ""

	return nil
}

func (d *Driver) Remove() error {
	machineState, err := d.GetState()
	if err != nil {
		return err
	}
	if machineState == state.Running {
		if err = d.Kill(); err != nil {
			return fmt.Errorf("can't stop VM: %s", err)
		}
	}
	vcConn := NewVcConn(d)
	if err = vcConn.VMDestroy(); err != nil {
		return err
	}
	return nil
}

func (d *Driver) Restart() error {
	if err := d.Stop(); err != nil {
		return err
	}
	// Check for 120 seconds for the machine to stop
	for i := 1; i <= 60; i++ {
		machineState, err := d.GetState()
		if err != nil {
			return err
		}
		if machineState == state.Running {
			log.Debugf("Not there yet %d/%d", i, 60)
			time.Sleep(2 * time.Second)
			continue
		}
		if machineState == state.Stopped {
			break
		}
	}

	machineState, err := d.GetState()
	// If the VM is still running after 120 seconds just kill it.
	if machineState == state.Running {
		if err = d.Kill(); err != nil {
			return fmt.Errorf("can't stop VM: %s", err)
		}
	}

	return d.Start()
}

func (d *Driver) Kill() error {
	vcConn := NewVcConn(d)
	if err := vcConn.VMPowerOff(); err != nil {
		return err
	}

	d.IPAddress = ""

	return nil
}

func (d *Driver) Upgrade() error {
	return fmt.Errorf("upgrade is not supported for vsphere driver at this moment")
}

func (d *Driver) publicSSHKeyPath() string {
	return d.GetSSHKeyPath() + ".pub"
}

func (d *Driver) checkVsphereConfig() error {
	if d.IP == "" {
		return errors.NewIncompleteVsphereConfigError("vSphere IP")
	}
	if d.Username == "" {
		return errors.NewIncompleteVsphereConfigError("vSphere username")
	}
	if d.Password == "" {
		return errors.NewIncompleteVsphereConfigError("vSphere password")
	}
	if d.Network == "" {
		return errors.NewIncompleteVsphereConfigError("vSphere network")
	}
	if d.Datastore == "" {
		return errors.NewIncompleteVsphereConfigError("vSphere datastore")
	}
	if d.Datacenter == "" {
		return errors.NewIncompleteVsphereConfigError("vSphere datacenter")
	}
	return nil
}

// Make a boot2docker userdata.tar key bundle
func (d *Driver) generateKeyBundle() error {
	log.Debugf("Creating Tar key bundle...")

	magicString := "boot2docker, this is vmware speaking"

	tf, err := os.Create(d.ResolveStorePath("userdata.tar"))
	if err != nil {
		return err
	}
	defer tf.Close()
	var fileWriter = tf

	tw := tar.NewWriter(fileWriter)
	defer tw.Close()

	// magicString first so we can figure out who originally wrote the tar.
	file := &tar.Header{Name: magicString, Size: int64(len(magicString))}
	if err := tw.WriteHeader(file); err != nil {
		return err
	}
	if _, err := tw.Write([]byte(magicString)); err != nil {
		return err
	}
	// .ssh/key.pub => authorized_keys
	file = &tar.Header{Name: ".ssh", Typeflag: tar.TypeDir, Mode: 0700}
	if err := tw.WriteHeader(file); err != nil {
		return err
	}
	pubKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return err
	}
	file = &tar.Header{Name: ".ssh/authorized_keys", Size: int64(len(pubKey)), Mode: 0644}
	if err := tw.WriteHeader(file); err != nil {
		return err
	}
	if _, err := tw.Write([]byte(pubKey)); err != nil {
		return err
	}
	file = &tar.Header{Name: ".ssh/authorized_keys2", Size: int64(len(pubKey)), Mode: 0644}
	if err := tw.WriteHeader(file); err != nil {
		return err
	}
	if _, err := tw.Write([]byte(pubKey)); err != nil {
		return err
	}
	if err := tw.Close(); err != nil {
		return err
	}

	return nil

}

func (d *Driver) UpgradeISO() error {

	vcConn := NewVcConn(d)

	if _, err := os.Stat(d.ISO); os.IsNotExist(err) {
		log.Errorf("Unable to find boot2docker ISO at %s", d.ISO)
		return errors.NewIncompleteVsphereConfigError(d.ISO)
	}

	if err := vcConn.DatastoreUpload(d.ISO, d.MachineName); err != nil {
		return err
	}

	return nil

}
