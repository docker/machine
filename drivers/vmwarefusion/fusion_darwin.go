/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package vmwarefusion

import (
	"archive/tar"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/log"
	"github.com/docker/machine/provider"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
	"github.com/docker/machine/utils"
)

const (
	B2DUser     = "docker"
	B2DPass     = "tcuser"
	isoFilename = "boot2docker-1.6.0-vmw.iso"
)

// Driver for VMware Fusion
type Driver struct {
	MachineName    string
	IPAddress      string
	Memory         int
	DiskSize       int
	CPU            int
	ISO            string
	Boot2DockerURL string
	CaCertPath     string
	PrivateKeyPath string
	SwarmMaster    bool
	SwarmHost      string
	SwarmDiscovery string
	CPUS           int
	SSHUser        string
	SSHPort        int

	storePath string
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
	return &Driver{MachineName: machineName, storePath: storePath, CaCertPath: caCert, PrivateKeyPath: privateKey}, nil
}

func (d *Driver) AuthorizePort(ports []*drivers.Port) error {
	return nil
}

func (d *Driver) DeauthorizePort(ports []*drivers.Port) error {
	return nil
}

func (d *Driver) GetMachineName() string {
	return d.MachineName
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) GetSSHKeyPath() string {
	return filepath.Join(d.storePath, "id_rsa")
}

func (d *Driver) GetSSHPort() (int, error) {
	if d.SSHPort == 0 {
		d.SSHPort = 22
	}

	return d.SSHPort, nil
}

func (d *Driver) GetSSHUsername() string {
	if d.SSHUser == "" {
		d.SSHUser = "docker"
	}

	return d.SSHUser
}

func (d *Driver) GetProviderType() provider.ProviderType {
	return provider.Local
}

func (d *Driver) DriverName() string {
	return "vmwarefusion"
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.Memory = flags.Int("vmwarefusion-memory-size")
	d.CPU = flags.Int("vmwarefusion-cpu-count")
	d.DiskSize = flags.Int("vmwarefusion-disk-size")
	d.Boot2DockerURL = flags.String("vmwarefusion-boot2docker-url")
	d.ISO = path.Join(d.storePath, isoFilename)
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

	var (
		isoURL string
		err    error
	)

	b2dutils := utils.NewB2dUtils("", "")

	imgPath := utils.GetMachineCacheDir()
	commonIsoPath := filepath.Join(imgPath, isoFilename)
	// just in case boot2docker.iso has been manually deleted
	if _, err := os.Stat(imgPath); os.IsNotExist(err) {
		if err := os.Mkdir(imgPath, 0700); err != nil {
			return err
		}
	}

	if d.Boot2DockerURL != "" {
		isoURL = d.Boot2DockerURL
		log.Infof("Downloading boot2docker.iso from %s...", isoURL)
		if err := b2dutils.DownloadISO(d.storePath, isoFilename, isoURL); err != nil {
			return err
		}
	} else {
		// TODO: until vmw tools are merged into b2d master
		// we will use the iso from the vmware team.
		//// todo: check latest release URL, download if it's new
		//// until then always use "latest"
		//isoURL, err = b2dutils.GetLatestBoot2DockerReleaseURL()
		//if err != nil {
		//	log.Warnf("Unable to check for the latest release: %s", err)
		//}

		// see https://github.com/boot2docker/boot2docker/pull/747
		isoURL := "https://github.com/cloudnativeapps/boot2docker/releases/download/v1.6.0-vmw/boot2docker-1.6.0-vmw.iso"

		if _, err := os.Stat(commonIsoPath); os.IsNotExist(err) {
			log.Infof("Downloading boot2docker.iso to %s...", commonIsoPath)
			// just in case boot2docker.iso has been manually deleted
			if _, err := os.Stat(imgPath); os.IsNotExist(err) {
				if err := os.Mkdir(imgPath, 0700); err != nil {
					return err
				}
			}
			if err := b2dutils.DownloadISO(imgPath, isoFilename, isoURL); err != nil {
				return err
			}
		}

		isoDest := filepath.Join(d.storePath, isoFilename)
		if err := utils.CopyFile(commonIsoPath, isoDest); err != nil {
			return err
		}
	}

	log.Infof("Creating SSH key...")
	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return err
	}

	log.Infof("Creating VM...")
	if err := os.MkdirAll(d.storePath, 0755); err != nil {
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
	diskImg := filepath.Join(d.storePath, fmt.Sprintf("%s.vmdk", d.MachineName))
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
			break
		}
	}

	if ip == "" {
		return fmt.Errorf("Machine didn't return an IP after 120 seconds, aborting")
	}

	// we got an IP, let's copy ssh keys over
	d.IPAddress = ip

	// Generate a tar keys bundle
	if err := d.generateKeyBundle(); err != nil {
		return err
	}

	// Test if /var/lib/boot2docker exists
	vmrun("-gu", B2DUser, "-gp", B2DPass, "directoryExistsInGuest", d.vmxPath(), "/var/lib/boot2docker")

	// Copy SSH keys bundle
	vmrun("-gu", B2DUser, "-gp", B2DPass, "CopyFileFromHostToGuest", d.vmxPath(), path.Join(d.storePath, "userdata.tar"), "/home/docker/userdata.tar")

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
			vmrun("-gu", B2DUser, "-gp", B2DPass, "runScriptInGuest", d.vmxPath(), "/bin/sh", "sudo mkdir "+shareDir+" && sudo mount -t vmhgfs .host:/"+shareName+" "+shareDir)
		}
	}
	return nil
}

func (d *Driver) Start() error {
	log.Infof("Starting %s...", d.MachineName)
	vmrun("start", d.vmxPath(), "nogui")

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
			vmrun("-gu", B2DUser, "-gp", B2DPass, "runScriptInGuest", d.vmxPath(), "/bin/sh", "sudo mkdir "+shareDir+" && sudo mount -t vmhgfs .host:/"+shareName+" "+shareDir)
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
	return path.Join(d.storePath, fmt.Sprintf("%s.vmx", d.MachineName))
}

func (d *Driver) vmdkPath() string {
	return path.Join(d.storePath, fmt.Sprintf("%s.vmdk", d.MachineName))
}

// Download boot2docker ISO image for the given tag and save it at dest.
func downloadISO(dir, file, url string) error {
	rsp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	// Download to a temp file first then rename it to avoid partial download.
	f, err := ioutil.TempFile(dir, file+".tmp")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err := io.Copy(f, rsp.Body); err != nil {
		// TODO: display download progress?
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(f.Name(), path.Join(dir, file)); err != nil {
		return err
	}
	return nil
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

	tf, err := os.Create(path.Join(d.storePath, "userdata.tar"))
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
