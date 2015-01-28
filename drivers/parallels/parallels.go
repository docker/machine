package parallels

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
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
)

type Driver struct {
	MachineName    string
	IPAddress      string
	Memory         int
	DiskSize       int
	ISO            string
	Boot2DockerURL string
	CaCertPath     string
	PrivateKeyPath string
	storePath      string
}

type CreateFlags struct {
	Memory         *int
	DiskSize       *int
	Boot2DockerURL *string
}

func init() {
	drivers.Register("parallels", &drivers.RegisteredDriver{
		New:            NewDriver,
		GetCreateFlags: GetCreateFlags,
	})
}

// RegisterCreateFlags registers the flags this driver adds to
// "docker hosts create"
func GetCreateFlags() []cli.Flag {
	return []cli.Flag{
		cli.IntFlag{
			Name:  "parallels-memory",
			Usage: "Size of memory for host in MB",
			Value: 1024,
		},
		cli.IntFlag{
			Name:  "parallels-disk-size",
			Usage: "Size of disk for host in MB",
			Value: 20000,
		},
		cli.StringFlag{
			EnvVar: "PARALLELS_BOOT2DOCKER_URL",
			Name:   "parallels-boot2docker-url",
			Usage:  "The URL of the boot2docker image. Defaults to the latest available version",
			Value:  "",
		},
	}
}

func NewDriver(machineName string, storePath string, caCert string, privateKey string) (drivers.Driver, error) {
	return &Driver{MachineName: machineName, storePath: storePath, CaCertPath: caCert, PrivateKeyPath: privateKey}, nil
}

func (d *Driver) DriverName() string {
	return "parallels"
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

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.Memory = flags.Int("parallels-memory")
	d.DiskSize = flags.Int("parallels-disk-size")
	d.Boot2DockerURL = flags.String("parallels-boot2docker-url")
	d.ISO = path.Join(d.storePath, "boot2docker.iso")

	return nil
}

func cpIso(src, dest string) error {
	buf, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(dest, buf, 0600); err != nil {
		return err
	}
	return nil
}

func (d *Driver) PreCreateCheck() error {
	return nil
}

func (d *Driver) Create() error {
	var (
		err    error
		isoURL string
	)

	// Check that prctl exists and works
	if err = prlctl("--version"); err != nil {
		return err
	}

	if d.Boot2DockerURL != "" {
		isoURL = d.Boot2DockerURL
		log.Infof("Downloading boot2docker.iso from %s...", isoURL)
		if err := utils.DownloadISO(d.storePath, "boot2docker.iso", isoURL); err != nil {
			return err
		}
	} else {
		// todo: check latest release URL, download if it's new
		// until then always use "latest"
		isoURL, err = utils.GetLatestBoot2DockerReleaseURL()
		if err != nil {
			return err
		}

		// todo: use real constant for .docker
		rootPath := filepath.Join(utils.GetHomeDir(), ".docker")
		imgPath := filepath.Join(rootPath, "images")
		commonIsoPath := filepath.Join(imgPath, "boot2docker.iso")
		if _, err := os.Stat(commonIsoPath); os.IsNotExist(err) {
			log.Infof("Downloading boot2docker.iso to %s...", commonIsoPath)

			// just in case boot2docker.iso has been manually deleted
			if _, err := os.Stat(imgPath); os.IsNotExist(err) {
				if err := os.Mkdir(imgPath, 0700); err != nil {
					return err
				}
			}

			if err := utils.DownloadISO(imgPath, "boot2docker.iso", isoURL); err != nil {
				return err
			}
		}

		isoDest := filepath.Join(d.storePath, "boot2docker.iso")
		if err := utils.CopyFile(commonIsoPath, isoDest); err != nil {
			return err
		}
	}

	log.Infof("Creating SSH key...")
	if err := ssh.GenerateSSHKey(d.sshKeyPath()); err != nil {
		return err
	}

	log.Infof("Creating Parallels Desktop VM...")
	if err := os.MkdirAll(d.storePath, 0755); err != nil {
		return err
	}

	if err := prlctl("create", d.MachineName,
		"--distribution", "linux-2.6",
		"--dst", d.storePath,
		"--no-hdd",
	); err != nil {
		return err
	}

	cpus := uint(runtime.NumCPU())
	if cpus > 32 {
		cpus = 32
	}

	if err := prlctl("set", d.MachineName,
		"--select-boot-device", "off",
		"--cpus", fmt.Sprintf("%d", cpus),
		"--memsize", fmt.Sprintf("%d", d.Memory),
		"--cpu-hotplug", "off",
		"--nested-virt", "on",
		"--pmu-virt", "on",
		"--on-window-close", "keep-running",
		"--longer-battery-life", "on",
		"--3d-accelerate", "off",
		"--device-bootorder", "cdrom0"); err != nil {
		return err
	}

	if err := prlctl("set", d.MachineName,
		"--iface", "net0",
		"--adapter-type", "virtio"); err != nil {
		return err
	}

	if err := prlctl("set", d.MachineName,
		"--device-set", "cdrom0",
		"--iface", "sata",
		"--position", "0",
		"--image", filepath.Join(d.storePath, "boot2docker.iso")); err != nil {
		return err
	}

	if err := prlctl("set", d.MachineName,
		"--device-add", "hdd",
		"--iface", "sata",
		"--position", "1",
		"--type", "expand",
		"--size", fmt.Sprintf("%d", d.DiskSize),
		"--image", d.diskPath(),
	); err != nil {
		return err
	}

	// Don't use Start() since it expects to have a dhcp lease allready
	if err := prlctl("start", d.MachineName); err != nil {
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

	// TODO
	// prlctl exec will not work without parallels tools in b2d. Since getting stuff into TCL
	// is much more painful, we simply use the b2d password to get the initial public key
	// onto the machine. From then on we use the pub key.
	sshConfig := &cssh.ClientConfig{
		User: B2D_USER,
		Auth: []cssh.AuthMethod{
			cssh.Password(B2D_PASS),
		},
	}

	var sshClient *cssh.Client

	for i := 1; i <= 20; i++ {
		sshClient, err = cssh.Dial("tcp", fmt.Sprintf("%s:22", ip), sshConfig)

		if err != nil {
			log.Debugf("Not there yet %d/%d, error: %s", i, 60, err)
			time.Sleep(2 * time.Second)
			continue
		}

		if sshClient != nil {
			log.Debugf("Connected!")
			break
		}
	}

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
	if err := prlctl("start", d.MachineName); err != nil {
		return err
	}

	ip, err := d.GetIP()
	if err != nil {
		return err
	}

	log.Infof("Waiting for VM to start...")
	return ssh.WaitForTCP(fmt.Sprintf("%s:22", ip))
}

func (d *Driver) Stop() error {
	if err := prlctl("stop", d.MachineName); err != nil {
		return err
	}
	for {
		s, err := d.GetState()
		if err != nil {
			return err
		}
		if s == state.Running {
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
	return nil
}

func (d *Driver) Remove() error {
	s, err := d.GetState()
	if err != nil {
		if err == ErrMachineNotExist {
			log.Infof("machine does not exist, assuming it has been removed already")
			return nil
		}
		return err
	}
	if s == state.Running {
		if err := d.Stop(); err != nil {
			return err
		}
	}
	return prlctl("delete", d.MachineName)
}

func (d *Driver) Restart() error {
	if err := d.Stop(); err != nil {
		return err
	}
	return d.Start()
}

func (d *Driver) Kill() error {
	return prlctl("stop", d.MachineName, "--kill")
}

func (d *Driver) Upgrade() error {
	log.Infof("Stopping machine...")
	if err := d.Stop(); err != nil {
		return err
	}

	isoURL, err := utils.GetLatestBoot2DockerReleaseURL()
	if err != nil {
		return err
	}

	log.Infof("Downloading boot2docker...")
	if err := utils.DownloadISO(d.storePath, "boot2docker.iso", isoURL); err != nil {
		return err
	}

	log.Infof("Starting machine...")
	if err := d.Start(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) GetState() (state.State, error) {
	stdout, stderr, err := prlctlOutErr("list", d.MachineName, "--output", "status", "--no-header")
	if err != nil {
		if reMachineNotFound.FindString(stderr) != "" {
			return state.Error, ErrMachineNotExist
		}
		return state.Error, err
	}

	switch stdout {
	// TODO: state.Starting ?!
	case "running\n":
		return state.Running, nil
	case "paused\n":
		return state.Paused, nil
	case "suspended\n":
		return state.Saved, nil
	case "stopping\n":
		return state.Stopping, nil
	case "stopped\n":
		return state.Stopped, nil
	}
	return state.None, nil
}

func (d *Driver) GetIP() (string, error) {
	ip, err := d.getIPfromDHCPLease()
	if err != nil {
		return "", err
	}

	return ip, nil
}

func (d *Driver) getIPfromDHCPLease() (string, error) {

	dhcp_lease_file := "/Library/Preferences/Parallels/parallels_dhcp_leases"

	stdout, err := prlctlOut("list", "-i", d.MachineName)
	macRe := regexp.MustCompile("net0.* mac=([0-9A-F]{12}) card=.*")
	macMatch := macRe.FindAllStringSubmatch(stdout, 1)

	if len(macMatch) != 1 {
		return "", fmt.Errorf("MAC address for NIC: nic0 on Virtual Machine: %s not found!\n", d.MachineName)
	}
	mac := macMatch[0][1]

	if len(mac) != 12 {
		return "", fmt.Errorf("Not a valid MAC address: %s. It should be exactly 12 digits.", mac)
	}

	leases, err := ioutil.ReadFile(dhcp_lease_file)
	if err != nil {
		return "", err
	}

	ipRe := regexp.MustCompile("(.*)=\"(.*),(.*)," + strings.ToLower(mac) + ",.*\"")
	mostRecentIp := ""
	mostRecentLease := uint64(0)
	for _, l := range ipRe.FindAllStringSubmatch(string(leases), -1) {
		ip := l[1]
		expiry, _ := strconv.ParseUint(l[2], 10, 64)
		leaseTime, _ := strconv.ParseUint(l[3], 10, 32)
		log.Debugf("Found lease: %s for MAC: %s, expiring at %d, leased for %d s.\n", ip, mac, expiry, leaseTime)
		if mostRecentLease <= expiry-leaseTime {
			mostRecentIp = ip
			mostRecentLease = expiry - leaseTime
		}
	}

	if len(mostRecentIp) == 0 {
		return "", fmt.Errorf("IP lease not found for MAC address %s in: %s\n", mac, dhcp_lease_file)
	}
	log.Debugf("Found IP lease: %s for MAC address %s\n", mostRecentIp, mac)

	return mostRecentIp, nil
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

func (d *Driver) GetSSHCommand(args ...string) (*exec.Cmd, error) {
	ip, err := d.GetIP()
	if err != nil {
		return nil, err
	}
	return ssh.GetSSHCommand(ip, 22, "docker", d.sshKeyPath(), args...), nil
}

func (d *Driver) sshKeyPath() string {
	return path.Join(d.storePath, "id_rsa")
}

func (d *Driver) publicSSHKeyPath() string {
	return d.sshKeyPath() + ".pub"
}

func (d *Driver) diskPath() string {
	return filepath.Join(d.storePath, "disk.hdd")
}
