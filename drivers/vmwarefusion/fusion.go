/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package vmwarefusion

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"regexp"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
	cryptossh "golang.org/x/crypto/ssh"
)

const (
	B2DUser        = "docker"
	B2DPass        = "tcuser"
	isoFilename    = "boot2docker.iso"
	isoConfigDrive = "configdrive.iso"
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

	SSHPassword    string
	ConfigDriveISO string
	ConfigDriveURL string
}

const (
	defaultSSHUser  = B2DUser
	defaultSSHPass  = B2DPass
	defaultDiskSize = 20000
	defaultCpus     = 1
	defaultMemory   = 1024
)

func init() {
	drivers.Register("vmwarefusion", &drivers.RegisteredDriver{
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
		cli.StringFlag{
			EnvVar: "FUSION_CONFIGDRIVE_URL",
			Name:   "vmwarefusion-configdrive-url",
			Usage:  "Fusion URL for cloud-init configdrive",
		},
		cli.IntFlag{
			EnvVar: "FUSION_CPU_COUNT",
			Name:   "vmwarefusion-cpu-count",
			Usage:  "number of CPUs for the machine (-1 to use the number of CPUs available)",
			Value:  defaultCpus,
		},
		cli.IntFlag{
			EnvVar: "FUSION_MEMORY_SIZE",
			Name:   "vmwarefusion-memory-size",
			Usage:  "Fusion size of memory for host VM (in MB)",
			Value:  defaultMemory,
		},
		cli.IntFlag{
			EnvVar: "FUSION_DISK_SIZE",
			Name:   "vmwarefusion-disk-size",
			Usage:  "Fusion size of disk for host VM (in MB)",
			Value:  defaultDiskSize,
		},
		cli.StringFlag{
			EnvVar: "FUSION_SSH_USER",
			Name:   "vmwarefusion-ssh-user",
			Usage:  "SSH user",
			Value:  defaultSSHUser,
		},
		cli.StringFlag{
			EnvVar: "FUSION_SSH_PASSWORD",
			Name:   "vmwarefusion-ssh-password",
			Usage:  "SSH password",
			Value:  defaultSSHPass,
		},
	}
}

func NewDriver(hostName, storePath string) drivers.Driver {
	return &Driver{
		CPUS:        defaultCpus,
		Memory:      defaultMemory,
		DiskSize:    defaultDiskSize,
		SSHPassword: defaultSSHPass,
		BaseDriver: &drivers.BaseDriver{
			SSHUser:     defaultSSHUser,
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
	return "vmwarefusion"
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.Memory = flags.Int("vmwarefusion-memory-size")
	d.CPU = flags.Int("vmwarefusion-cpu-count")
	d.DiskSize = flags.Int("vmwarefusion-disk-size")
	d.Boot2DockerURL = flags.String("vmwarefusion-boot2docker-url")
	d.ConfigDriveURL = flags.String("vmwarefusion-configdrive-url")
	d.ISO = d.ResolveStorePath(isoFilename)
	d.SwarmMaster = flags.Bool("swarm-master")
	d.SwarmHost = flags.String("swarm-host")
	d.SwarmDiscovery = flags.String("swarm-discovery")
	d.SSHUser = flags.String("vmwarefusion-ssh-user")
	d.SSHPassword = flags.String("vmwarefusion-ssh-password")
	d.SSHPort = 22

	// We support a maximum of 16 cpu to be consistent with Virtual Hardware 10
	// specs.
	if d.CPU < 1 {
		d.CPU = int(runtime.NumCPU())
	}
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
	b2dutils := mcnutils.NewB2dUtils("", "", d.StorePath)
	if err := b2dutils.CopyIsoToMachineDir(d.Boot2DockerURL, d.MachineName); err != nil {
		return err
	}

	// download cloud-init config drive
	if d.ConfigDriveURL != "" {
		log.Infof("Downloading %s from %s", isoConfigDrive, d.ConfigDriveURL)
		if err := b2dutils.DownloadISO(d.ResolveStorePath("."), isoConfigDrive, d.ConfigDriveURL); err != nil {
			return err
		}
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
		ip, err = d.getIPfromDHCPLease()
		if err != nil {
			log.Debugf("Not there yet %d/%d, error: %s", i, 60, err)
			time.Sleep(2 * time.Second)
			continue
		}

		if ip != "" {
			log.Debugf("Got an ip: %s", ip)
			conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, 22), time.Duration(2*time.Second))
			if err != nil {
				log.Debugf("SSH Daemon not responding yet: %s", err)
				time.Sleep(2 * time.Second)
				continue
			}
			conn.Close()
			break
		}
	}

	if ip == "" {
		return fmt.Errorf("Machine didn't return an IP after 120 seconds, aborting")
	}

	// we got an IP, let's copy ssh keys over
	d.IPAddress = ip

	// Do not execute the rest of boot2docker specific configuration
	// The uplaod of the public ssh key uses a ssh connection,
	// this works without installed vmware client tools
	if d.ConfigDriveURL != "" {
		var keyfh *os.File
		var keycontent []byte

		log.Infof("Copy public SSH key to %s [%s]", d.MachineName, d.IPAddress)

		// create .ssh folder in users home
		if err := executeSSHCommand(fmt.Sprintf("mkdir -p /home/%s/.ssh", d.SSHUser), d); err != nil {
			return err
		}

		// read generated public ssh key
		if keyfh, err = os.Open(d.publicSSHKeyPath()); err != nil {
			return err
		}
		defer keyfh.Close()

		if keycontent, err = ioutil.ReadAll(keyfh); err != nil {
			return err
		}

		// add public ssh key to authorized_keys
		if err := executeSSHCommand(fmt.Sprintf("echo '%s' > /home/%s/.ssh/authorized_keys", string(keycontent), d.SSHUser), d); err != nil {
			return err
		}

		// make it secure
		if err := executeSSHCommand(fmt.Sprintf("chmod 600 /home/%s/.ssh/authorized_keys", d.SSHUser), d); err != nil {
			return err
		}

		log.Debugf("Leaving create sequence early, configdrive found")
		return nil
	}

	// Generate a tar keys bundle
	if err := d.generateKeyBundle(); err != nil {
		return err
	}

	// Test if /var/lib/boot2docker exists
	vmrun("-gu", B2DUser, "-gp", B2DPass, "directoryExistsInGuest", d.vmxPath(), "/var/lib/boot2docker")

	// Copy SSH keys bundle
	vmrun("-gu", B2DUser, "-gp", B2DPass, "CopyFileFromHostToGuest", d.vmxPath(), d.ResolveStorePath("userdata.tar"), "/home/docker/userdata.tar")

	// Expand tar file.
	vmrun("-gu", B2DUser, "-gp", B2DPass, "runScriptInGuest", d.vmxPath(), "/bin/sh", "sudo /bin/mv /home/docker/userdata.tar /var/lib/boot2docker/userdata.tar && sudo tar xf /var/lib/boot2docker/userdata.tar -C /home/docker/ > /var/log/userdata.log 2>&1 && sudo chown -R docker:staff /home/docker")

	// Enable Shared Folders
	vmrun("-gu", B2DUser, "-gp", B2DPass, "enableSharedFolders", d.vmxPath())

	var shareName, shareDir string // TODO configurable at some point
	switch runtime.GOOS {
	case "darwin":
		shareName = "Users"
		shareDir = "/Users"
		// TODO "linux" and "windows"
	}

	if shareDir != "" {
		if _, err := os.Stat(shareDir); err != nil && !os.IsNotExist(err) {
			return err
		} else if !os.IsNotExist(err) {
			// add shared folder, create mountpoint and mount it.
			vmrun("-gu", B2DUser, "-gp", B2DPass, "addSharedFolder", d.vmxPath(), shareName, shareDir)
			command := "[ ! -d " + shareDir + " ]&& sudo mkdir " + shareDir + "; [ -f /usr/local/bin/vmhgfs-fuse ]&& sudo /usr/local/bin/vmhgfs-fuse -o allow_other .host:/" + shareName + " " + shareDir + " || sudo mount -t vmhgfs .host:/" + shareName + " " + shareDir
			vmrun("-gu", B2DUser, "-gp", B2DPass, "runScriptInGuest", d.vmxPath(), "/bin/sh", command)
		}
	}
	return nil
}

func (d *Driver) Start() error {
	log.Infof("Starting %s...", d.MachineName)
	vmrun("start", d.vmxPath(), "nogui")

	// Do not execute the rest of boot2docker specific configuration, exit here
	if d.ConfigDriveURL != "" {
		log.Debugf("Leaving start sequence early, configdrive found")
		return nil
	}

	log.Debugf("Mounting Shared Folders...")
	var shareName, shareDir string // TODO configurable at some point
	switch runtime.GOOS {
	case "darwin":
		shareName = "Users"
		shareDir = "/Users"
		// TODO "linux" and "windows"
	}

	if shareDir != "" {
		if _, err := os.Stat(shareDir); err != nil && !os.IsNotExist(err) {
			return err
		} else if !os.IsNotExist(err) {
			// create mountpoint and mount shared folder
			command := "[ ! -d " + shareDir + " ]&& sudo mkdir " + shareDir + "; [ -f /usr/local/bin/vmhgfs-fuse ]&& sudo /usr/local/bin/vmhgfs-fuse -o allow_other .host:/" + shareName + " " + shareDir + " || sudo mount -t vmhgfs .host:/" + shareName + " " + shareDir
			vmrun("-gu", B2DUser, "-gp", B2DPass, "runScriptInGuest", d.vmxPath(), "/bin/sh", command)
		}
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

// execute command over SSH with user / password authentication
func executeSSHCommand(command string, d *Driver) error {
	log.Debugf("Execute executeSSHCommand: %s", command)

	config := &cryptossh.ClientConfig{
		User: d.SSHUser,
		Auth: []cryptossh.AuthMethod{
			cryptossh.Password(d.SSHPassword),
		},
	}

	client, err := cryptossh.Dial("tcp", fmt.Sprintf("%s:%d", d.IPAddress, d.SSHPort), config)
	if err != nil {
		log.Debugf("Failed to dial:", err)
		return err
	}

	session, err := client.NewSession()
	if err != nil {
		log.Debugf("Failed to create session: " + err.Error())
		return err
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b

	if err := session.Run(command); err != nil {
		log.Debugf("Failed to run: " + err.Error())
		return err
	}
	log.Debugf("Stdout from executeSSHCommand: %s", b.String())

	return nil
}
