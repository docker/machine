/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package vmwarefusion

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
	"github.com/docker/machine/utils"
	cssh "golang.org/x/crypto/ssh"
)

const (
	B2D_USER        = "docker"
	B2D_PASS        = "tcuser"
	dockerConfigDir = "/var/lib/boot2docker"
	isoFilename     = "boot2docker-vmw.iso"
)

// Driver for VMware Fusion
type Driver struct {
	MachineName    string
	IPAddress      string
	Memory         int
	DiskSize       int
	ISO            string
	Boot2DockerURL string
	CaCertPath     string
	PrivateKeyPath string

	storePath string
}

type CreateFlags struct {
	Boot2DockerURL *string
	Memory         *int
	DiskSize       *int
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

func (d *Driver) DriverName() string {
	return "vmwarefusion"
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.Memory = flags.Int("vmwarefusion-memory-size")
	d.DiskSize = flags.Int("vmwarefusion-disk-size")
	d.Boot2DockerURL = flags.String("vmwarefusion-boot2docker-url")
	d.ISO = path.Join(d.storePath, isoFilename)

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

	if d.Boot2DockerURL != "" {
		isoURL = d.Boot2DockerURL
		log.Infof("Downloading boot2docker.iso from %s...", isoURL)
		if err := b2dutils.DownloadISO(d.storePath, isoFilename, isoURL); err != nil {
			return err
		}

	} else {
		// TODO: until vmw tools are merged into b2d master
		// we will use the iso from the vmware team
		//// todo: check latest release URL, download if it's new
		//// until then always use "latest"
		//isoURL, err = b2dutils.GetLatestBoot2DockerReleaseURL()
		//if err != nil {
		//	log.Warnf("Unable to check for the latest release: %s", err)
		//}

		isoURL := "https://github.com/cloudnativeapps/boot2docker/releases/download/v1.5.0-vmw/boot2docker-1.5.0-vmw.iso"

		rootPath := filepath.Join(utils.GetDockerDir())
		imgPath := filepath.Join(rootPath, "images")
		commonIsoPath := filepath.Join(imgPath, isoFilename)
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
	if err := ssh.GenerateSSHKey(d.sshKeyPath()); err != nil {
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

	if err := d.Start(); err != nil {
		return err
	}

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

	d.IPAddress = ip

	key, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return err
	}

	// so, vmrun above will not work without vmtools in b2d.  since getting stuff into TCL
	// is much more painful, we simply use the b2d password to get the initial public key
	// onto the machine.  from then on we use the pub key.  meh.
	sshConfig := &cssh.ClientConfig{
		User: B2D_USER,
		Auth: []cssh.AuthMethod{
			cssh.Password(B2D_PASS),
		},
	}
	sshClient, err := cssh.Dial("tcp", fmt.Sprintf("%s:22", ip), sshConfig)
	if err != nil {
		return err
	}
	session, err := sshClient.NewSession()
	if err != nil {
		return err
	}
	if err := session.Run(fmt.Sprintf("mkdir /home/docker/.ssh && echo \"%s\" > /home/docker/.ssh/authorized_keys", string(key))); err != nil {
		return err
	}
	session.Close()

	log.Debugf("Setting hostname: %s", d.MachineName)
	cmd, err := d.GetSSHCommand(fmt.Sprintf(
		"echo \"127.0.0.1 %s\" | sudo tee -a /etc/hosts && sudo hostname %s && echo \"%s\" | sudo tee /etc/hostname",
		d.MachineName,
		d.MachineName,
		d.MachineName,
	))
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) Start() error {
	vmrun("start", d.vmxPath(), "nogui")
	return nil
}

func (d *Driver) Stop() error {
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

	vmrun("deleteVM", d.vmxPath(), "nogui")
	return nil
}

func (d *Driver) Restart() error {
	vmrun("reset", d.vmxPath(), "nogui")
	return nil
}

func (d *Driver) Kill() error {
	vmrun("stop", d.vmxPath(), "nogui")
	return nil
}

func (d *Driver) StartDocker() error {
	log.Debug("Starting Docker...")

	cmd, err := d.GetSSHCommand("sudo /etc/init.d/docker start")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) StopDocker() error {
	log.Debug("Stopping Docker...")

	cmd, err := d.GetSSHCommand("if [ -e /var/run/docker.pid ]; then sudo /etc/init.d/docker stop ; fi")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) GetDockerConfigDir() string {
	return dockerConfigDir
}

func (d *Driver) Upgrade() error {
	return nil
}

func (d *Driver) GetSSHCommand(args ...string) (*exec.Cmd, error) {
	ip, err := d.GetIP()
	if err != nil {
		return nil, err
	}
	return ssh.GetSSHCommand(ip, 22, "docker", d.sshKeyPath(), args...), nil
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

func (d *Driver) sshKeyPath() string {
	return path.Join(d.storePath, "id_rsa")
}

func (d *Driver) publicSSHKeyPath() string {
	return d.sshKeyPath() + ".pub"
}
