package parallels

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/log"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
	"github.com/docker/machine/utils"
)

const (
	isoFilename = "boot2docker.iso"
	minDiskSize = 32
)

type Driver struct {
	CPU            int
	MachineName    string
	SSHUser        string
	SSHPort        int
	IPAddress      string
	Memory         int
	DiskSize       int
	Boot2DockerURL string
	CaCertPath     string
	PrivateKeyPath string
	SwarmMaster    bool
	SwarmHost      string
	SwarmDiscovery string
	storePath      string
}

func init() {
	drivers.Register("parallels", &drivers.RegisteredDriver{
		New:            NewDriver,
		GetCreateFlags: GetCreateFlags,
	})
}

// GetCreateFlags registers the flags this driver adds to
// "docker hosts create"
func GetCreateFlags() []cli.Flag {
	return []cli.Flag{
		cli.IntFlag{
			EnvVar: "PARALLELS_MEMORY_SIZE",
			Name:   "parallels-memory",
			Usage:  "Size of memory for host in MB",
			Value:  1024,
		},
		cli.IntFlag{
			EnvVar: "PARALLELS_CPU_COUNT",
			Name:   "parallels-cpu-count",
			Usage:  "number of CPUs for the machine (-1 to use the number of CPUs available)",
			Value:  1,
		},
		cli.IntFlag{
			EnvVar: "PARALLELS_DISK_SIZE",
			Name:   "parallels-disk-size",
			Usage:  "Size of disk for host in MB",
			Value:  20000,
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
	return d.SSHPort, nil
}

func (d *Driver) GetSSHUsername() string {
	if d.SSHUser == "" {
		d.SSHUser = "docker"
	}

	return d.SSHUser
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
	d.CPU = flags.Int("parallels-cpu-count")
	d.Memory = flags.Int("parallels-memory")
	d.DiskSize = flags.Int("parallels-disk-size")
	d.Boot2DockerURL = flags.String("parallels-boot2docker-url")
	d.SwarmMaster = flags.Bool("swarm-master")
	d.SwarmHost = flags.String("swarm-host")
	d.SwarmDiscovery = flags.String("swarm-discovery")
	d.SSHUser = "docker"
	d.SSHPort = 22

	return nil
}

func (d *Driver) PreCreateCheck() error {
	return nil
}

func (d *Driver) Create() error {
	var (
		err error
	)

	// Check that prctl exists and works
	if err = prlctl("--version"); err != nil {
		return err
	}

	b2dutils := utils.NewB2dUtils("", "")
	if err := b2dutils.CopyIsoToMachineDir(d.Boot2DockerURL, d.MachineName); err != nil {
		return err
	}

	log.Infof("Creating SSH key...")
	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return err
	}

	log.Infof("Creating Parallels Desktop VM...")

	if err := prlctl("create", d.MachineName,
		"--distribution", "linux-2.6",
		"--dst", d.storePath,
		"--no-hdd"); err != nil {
		return err
	}

	cpus := d.CPU
	if cpus < 1 {
		cpus = int(runtime.NumCPU())
	}
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
		"--device-set", "cdrom0",
		"--iface", "sata",
		"--position", "0",
		"--image", filepath.Join(d.storePath, isoFilename)); err != nil {
		return err
	}

	// Create a small plain disk. It will be converted and expanded later
	if err := prlctl("set", d.MachineName,
		"--device-add", "hdd",
		"--iface", "sata",
		"--position", "1",
		"--image", d.diskPath(),
		"--type", "plain",
		"--size", fmt.Sprintf("%d", minDiskSize)); err != nil {
		return err
	}

	if err := d.generateDiskImage(d.DiskSize); err != nil {
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

	log.Infof("Starting Parallels Desktop VM...")

	if err := d.Start(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) Start() error {
	s, err := d.GetState()
	if err != nil {
		return err
	}

	switch s {
	case state.Stopped, state.Saved, state.Paused:
		if err := prlctl("start", d.MachineName); err != nil {
			return err
		}
		log.Infof("Waiting for VM to start...")
	case state.Running:
		break
	default:
		log.Infof("VM not in restartable state")
	}

	if err := drivers.WaitForSSH(d); err != nil {
		return err
	}

	d.IPAddress, err = d.GetIP()
	return err
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
		if err := d.Kill(); err != nil {
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
	// Assume that Parallels Desktop hosts don't have IPs unless they are running
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

func (d *Driver) publicSSHKeyPath() string {
	return d.GetSSHKeyPath() + ".pub"
}

func (d *Driver) diskPath() string {
	return filepath.Join(d.storePath, "disk.hdd")
}

// Make a boot2docker VM disk image.
func (d *Driver) generateDiskImage(size int) error {
	tarBuf, err := d.generateTar()
	if err != nil {
		return err
	}

	minSizeBytes := int64(minDiskSize) << 20 // usually won't fit in 32-bit int (max 2GB)

	//Expand the initial image if needed
	if bufLen := int64(tarBuf.Len()); bufLen > minSizeBytes {
		bufLenMBytes := bufLen>>20 + 1
		if err := prldisktool("resize",
			"--hdd", d.diskPath(),
			"--size", fmt.Sprintf("%d", bufLenMBytes)); err != nil {
			return err
		}
	}

	// Find hds file
	hdsList, err := filepath.Glob(d.diskPath() + "/*.hds")
	if err != nil {
		return err
	}
	if len(hdsList) == 0 {
		return fmt.Errorf("Could not find *.hds image in %s", d.diskPath())
	}
	hdsPath := hdsList[0]
	log.Debugf("HDS image path: %s", hdsPath)

	// Write tar to the hds file
	hds, err := os.OpenFile(hdsPath, os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer hds.Close()
	hds.Seek(0, os.SEEK_SET)
	_, err = hds.Write(tarBuf.Bytes())
	if err != nil {
		return err
	}
	hds.Close()

	// Convert image to expanding type and resize it
	if err := prldisktool("convert", "--expanding",
		"--hdd", d.diskPath()); err != nil {
		return err
	}

	if err := prldisktool("resize",
		"--hdd", d.diskPath(),
		"--size", fmt.Sprintf("%d", size)); err != nil {
		return err
	}

	return nil
}

// See https://github.com/boot2docker/boot2docker/blob/master/rootfs/rootfs/etc/rc.d/automount
func (d *Driver) generateTar() (*bytes.Buffer, error) {
	magicString := "boot2docker, please format-me"

	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	// magicString first so the automount script knows to format the disk
	file := &tar.Header{Name: magicString, Size: int64(len(magicString))}
	if err := tw.WriteHeader(file); err != nil {
		return nil, err
	}
	if _, err := tw.Write([]byte(magicString)); err != nil {
		return nil, err
	}
	// .ssh/key.pub => authorized_keys
	file = &tar.Header{Name: ".ssh", Typeflag: tar.TypeDir, Mode: 0700}
	if err := tw.WriteHeader(file); err != nil {
		return nil, err
	}
	pubKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return nil, err
	}
	file = &tar.Header{Name: ".ssh/authorized_keys", Size: int64(len(pubKey)), Mode: 0644}
	if err := tw.WriteHeader(file); err != nil {
		return nil, err
	}
	if _, err := tw.Write([]byte(pubKey)); err != nil {
		return nil, err
	}
	file = &tar.Header{Name: ".ssh/authorized_keys2", Size: int64(len(pubKey)), Mode: 0644}
	if err := tw.WriteHeader(file); err != nil {
		return nil, err
	}
	if _, err := tw.Write([]byte(pubKey)); err != nil {
		return nil, err
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}
	return buf, nil
}
