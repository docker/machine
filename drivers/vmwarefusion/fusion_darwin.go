/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package vmwarefusion

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/codegangsta/cli"
	"github.com/docker/docker/pkg/homedir"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/log"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
	"github.com/docker/machine/utils"
)

const (
	B2DUser     = "docker"
	B2DPass     = "docker"
	isoFilename = "boot2docker-vmware.iso"
)

// Driver for VMware Fusion
type Driver struct {
	*drivers.BaseDriver
	Memory         int
	DiskSize       int
	CPU            int
	ISO            string
	Boot2DockerURL string
	CPUS           int
}

func init() {
	drivers.Register("vmwarefusion", &drivers.RegisteredDriver{
		New:            NewDriver,
		GetCreateFlags: GetCreateFlags,
	})
}

// GetCreateFlags registers the flags this driver adds to
// "docker hosts create"
func GetCreateFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			EnvVar: "FUSION_BOOT2DOCKER_URL",
			Name:   "vmwarefusion-boot2docker-url",
			Usage:  "Fusion URL for boot2docker image",
		},
		cli.IntFlag{
			EnvVar: "FUSION_CPU_COUNT",
			Name:   "vmwarefusion-cpu-count",
			Usage:  "number of CPUs for the machine (-1 to use the number of CPUs available)",
			Value:  1,
		},
		cli.IntFlag{
			EnvVar: "FUSION_MEMORY_SIZE",
			Name:   "vmwarefusion-memory-size",
			Usage:  "Fusion size of memory for host VM (in MB)",
			Value:  1024,
		},
		cli.IntFlag{
			EnvVar: "FUSION_DISK_SIZE",
			Name:   "vmwarefusion-disk-size",
			Usage:  "Fusion size of disk for host VM (in MB)",
			Value:  20000,
		},
	}
}

func NewDriver(machineName string, storePath string, caCert string, privateKey string) (drivers.Driver, error) {
	inner := drivers.NewBaseDriver(machineName, storePath, caCert, privateKey)
	return &Driver{BaseDriver: inner}, nil
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
	return "vmwarefusion"
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.Memory = flags.Int("vmwarefusion-memory-size")
	d.CPU = flags.Int("vmwarefusion-cpu-count")
	d.DiskSize = flags.Int("vmwarefusion-disk-size")
	d.Boot2DockerURL = flags.String("vmwarefusion-boot2docker-url")
	d.ISO = d.ResolveStorePath(isoFilename)
	d.SwarmMaster = flags.Bool("swarm-master")
	d.SwarmHost = flags.String("swarm-host")
	d.SwarmDiscovery = flags.String("swarm-discovery")
	d.SSHUser = "docker"
	d.SSHPort = 22

	// We support a maximum of 16 cpu to be consistent with Virtual Hardware 10
	// specs.
	if d.CPU > 16 {
		d.CPU = 16
	}

	return nil
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	if ip == "" {
		return "", nil
	}
	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

func (d *Driver) GetIP() (string, error) {
	s, err := d.GetState()
	if err != nil {
		return "", err
	}
	if s != state.Running {
		return "", drivers.ErrHostIsNotRunning
	}

	ip, err := d.getIPfromDHCPLease()
	if err != nil {
		return "", err
	}

	return ip, nil
}

func (d *Driver) GetState() (state.State, error) {
	// VMRUN only tells use if the vm is running or not
	if stdout, _, _ := vmrun("list"); strings.Contains(stdout, d.vmxPath()) {
		return state.Running, nil
	}
	return state.Stopped, nil
}

func (d *Driver) PreCreateCheck() error {
	return nil
}

func (d *Driver) Create() error {

	b2dutils := utils.NewB2dUtils("", "", isoFilename)
	if err := b2dutils.CopyIsoToMachineDir(d.Boot2DockerURL, d.MachineName); err != nil {
		return err
	}

	log.Infof("Creating SSH key...")
	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return err
	}

	log.Infof("Creating VM...")
	if err := os.MkdirAll(d.ResolveStorePath("."), 0755); err != nil {
		return err
	}

	if _, err := os.Stat(d.vmxPath()); err == nil {
		return ErrMachineExist
	}

	// Generate vmx config file from template
	vmxt := template.Must(template.New("vmx").Parse(vmx))
	vmxfile, err := os.Create(d.vmxPath())
	if err != nil {
		return err
	}
	vmxt.Execute(vmxfile, d)

	// Generate vmdk file
	diskImg := d.ResolveStorePath(fmt.Sprintf("%s.vmdk", d.MachineName))
	if _, err := os.Stat(diskImg); err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		if err := vdiskmanager(diskImg, d.DiskSize); err != nil {
			return err
		}
	}

	log.Infof("Starting %s...", d.MachineName)
	vmrun("start", d.vmxPath(), "nogui")

	var ip string

	log.Infof("Waiting for VM to come online...")
	for i := 1; i <= 60; i++ {
		ip, err = d.GetIP()
		if err != nil {
			log.Debugf("Not there yet %d/%d, error: %s", i, 60, err)
			time.Sleep(2 * time.Second)
			continue
		}

		if ip != "" {
			log.Debugf("Got an ip: %s", ip)
			break
		}
	}

	if ip == "" {
		return fmt.Errorf("Machine didn't return an IP after 120 seconds, aborting")
	}

	// we got an IP, let's copy ssh keys over
	d.IPAddress = ip

	// use ssh to set keys
	sshClient, err := d.getLocalSSHClient()
	if err != nil {
		return err
	}

	// add pub key for user
	pubKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return err
	}

	if out, err := sshClient.Output(fmt.Sprintf(
		"mkdir -p /home/%s/.ssh",
		d.GetSSHUsername(),
	)); err != nil {
		log.Error(out)
		return err
	}

	if out, err := sshClient.Output(fmt.Sprintf(
		"printf '%%s' '%s' | tee /home/%s/.ssh/authorized_keys",
		string(pubKey),
		d.GetSSHUsername(),
	)); err != nil {
		log.Error(out)
		return err
	}

	// Enable Shared Folders
	vmrun("-gu", B2DUser, "-gp", B2DPass, "enableSharedFolders", d.vmxPath())

	if err := d.setupSharedDirs(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) Start() error {
	log.Infof("Starting %s...", d.MachineName)
	vmrun("start", d.vmxPath(), "nogui")

	log.Debugf("Mounting Shared Folders...")
	if err := d.setupSharedDirs(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) Stop() error {
	log.Infof("Gracefully shutting down %s...", d.MachineName)
	vmrun("stop", d.vmxPath(), "nogui")
	return nil
}

func (d *Driver) Remove() error {

	s, _ := d.GetState()
	if s == state.Running {
		if err := d.Kill(); err != nil {
			return fmt.Errorf("Error stopping VM before deletion")
		}
	}
	log.Infof("Deleting %s...", d.MachineName)
	vmrun("deleteVM", d.vmxPath(), "nogui")
	return nil
}

func (d *Driver) Restart() error {
	log.Infof("Gracefully restarting %s...", d.MachineName)
	vmrun("reset", d.vmxPath(), "nogui")
	return nil
}

func (d *Driver) Kill() error {
	log.Infof("Forcibly halting %s...", d.MachineName)
	vmrun("stop", d.vmxPath(), "hard nogui")
	return nil
}

func (d *Driver) Upgrade() error {
	return fmt.Errorf("VMware Fusion does not currently support the upgrade operation")
}

func (d *Driver) vmxPath() string {
	return d.ResolveStorePath(fmt.Sprintf("%s.vmx", d.MachineName))
}

func (d *Driver) vmdkPath() string {
	return d.ResolveStorePath(fmt.Sprintf("%s.vmdk", d.MachineName))
}

func (d *Driver) getIPfromDHCPLease() (string, error) {
	var vmxfh *os.File
	var dhcpfh *os.File
	var vmxcontent []byte
	var dhcpcontent []byte
	var macaddr string
	var err error
	var lastipmatch string
	var currentip string
	var lastleaseendtime time.Time
	var currentleadeendtime time.Time

	// DHCP lease table for NAT vmnet interface
	var dhcpfile = "/var/db/vmware/vmnet-dhcpd-vmnet8.leases"

	if vmxfh, err = os.Open(d.vmxPath()); err != nil {
		return "", err
	}
	defer vmxfh.Close()

	if vmxcontent, err = ioutil.ReadAll(vmxfh); err != nil {
		return "", err
	}

	// Look for generatedAddress as we're passing a VMX with addressType = "generated".
	vmxparse := regexp.MustCompile(`^ethernet0.generatedAddress\s*=\s*"(.*?)"\s*$`)
	for _, line := range strings.Split(string(vmxcontent), "\n") {
		if matches := vmxparse.FindStringSubmatch(line); matches == nil {
			continue
		} else {
			macaddr = strings.ToLower(matches[1])
		}
	}

	if macaddr == "" {
		return "", fmt.Errorf("couldn't find MAC address in VMX file %s", d.vmxPath())
	}

	log.Debugf("MAC address in VMX: %s", macaddr)
	if dhcpfh, err = os.Open(dhcpfile); err != nil {
		return "", err
	}
	defer dhcpfh.Close()

	if dhcpcontent, err = ioutil.ReadAll(dhcpfh); err != nil {
		return "", err
	}

	// Get the IP from the lease table.
	leaseip := regexp.MustCompile(`^lease (.+?) {$`)
	// Get the lease end date time.
	leaseend := regexp.MustCompile(`^\s*ends \d (.+?);$`)
	// Get the MAC address associated.
	leasemac := regexp.MustCompile(`^\s*hardware ethernet (.+?);$`)

	for _, line := range strings.Split(string(dhcpcontent), "\n") {

		if matches := leaseip.FindStringSubmatch(line); matches != nil {
			lastipmatch = matches[1]
			continue
		}

		if matches := leaseend.FindStringSubmatch(line); matches != nil {
			lastleaseendtime, _ = time.Parse("2006/01/02 15:04:05", matches[1])
			continue
		}

		if matches := leasemac.FindStringSubmatch(line); matches != nil && matches[1] == macaddr && currentleadeendtime.Before(lastleaseendtime) {
			currentip = lastipmatch
			currentleadeendtime = lastleaseendtime
		}
	}

	if currentip == "" {
		return "", fmt.Errorf("IP not found for MAC %s in DHCP leases", macaddr)
	}

	log.Debugf("IP found in DHCP lease table: %s", currentip)
	return currentip, nil

}

func (d *Driver) publicSSHKeyPath() string {
	return d.GetSSHKeyPath() + ".pub"
}

func (d *Driver) setupSharedDirs() error {
	shareDir := homedir.Get()
	shareName := "Home"

	if _, err := os.Stat(shareDir); err != nil && !os.IsNotExist(err) {
		return err
	} else if !os.IsNotExist(err) {
		// add shared folder, create mountpoint and mount it.
		vmrun("-gu", B2DUser, "-gp", B2DPass, "addSharedFolder", d.vmxPath(), shareName, shareDir)
		vmrun("-gu", B2DUser, "-gp", B2DPass, "runScriptInGuest", d.vmxPath(), "/bin/sh", "sudo mkdir -p "+shareDir+" && sudo mount -t vmhgfs .host:/"+shareName+" "+shareDir)
	}

	return nil
}

func (d *Driver) getLocalSSHClient() (ssh.Client, error) {
	ip, err := d.GetIP()
	if err != nil {
		return nil, err
	}

	sshAuth := &ssh.Auth{
		Passwords: []string{"docker"},
		Keys:      []string{d.GetSSHKeyPath()},
	}
	sshClient, err := ssh.NewNativeClient(d.GetSSHUsername(), ip, d.SSHPort, sshAuth)
	if err != nil {
		return nil, err
	}

	return sshClient, nil
}
