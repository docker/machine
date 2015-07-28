package xhyve

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"os/exec"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/log"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
	"github.com/docker/machine/utils"
)

const (
	isoFilename = "boot2docker.iso"
)

type Driver struct {
	MachineName    string
	IPAddress      string
	Memory         int
	DiskSize       int
	CPU            int
	ISO            string
	TmpISO         string
	UUID           string
	SSHUser        string
	SSHPort        int
	Boot2DockerURL string
	CaCertPath     string
	PrivateKeyPath string
	SwarmMaster    bool
	SwarmHost      string
	SwarmDiscovery string

	storePath string
}

type Stringer interface {
	String() string
}

func init() {
	drivers.Register("xhyve", &drivers.RegisteredDriver{
		New:            NewDriver,
		GetCreateFlags: GetCreateFlags,
	})
}

// RegisterCreateFlags registers the flags this driver adds to
// "docker hosts create"
func GetCreateFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			EnvVar: "XHYVE_BOOT2DOCKER_URL",
			Name:   "xhyve-boot2docker-url",
			Usage:  "The URL of the boot2docker image. Defaults to the latest available version",
			Value:  "",
		},
		cli.IntFlag{
			EnvVar: "XHYVE_CPU_COUNT",
			Name:   "xhyve-cpu-count",
			Usage:  "Number of CPUs for the machine (-1 to use the number of CPUs available)",
			Value:  1,
		},
		cli.IntFlag{
			EnvVar: "XHYVE_MEMORY_SIZE",
			Name:   "xhyve-memory",
			Usage:  "Size of memory for host in MB",
			Value:  1024,
		},
		cli.IntFlag{
			EnvVar: "XHYVE_DISK_SIZE",
			Name:   "xhyve-disk-size",
			Usage:  "Size of disk for host in MB",
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

func (d *Driver) DriverName() string {
	return "xhyve"
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.Boot2DockerURL = flags.String("xhyve-boot2docker-url")
	d.CPU = flags.Int("xhyve-cpu-count")
	d.Memory = flags.Int("xhyve-memory")
	d.DiskSize = flags.Int("xhyve-disk-size")
	d.ISO = path.Join(d.storePath, isoFilename)
	d.SwarmMaster = flags.Bool("swarm-master")
	d.SwarmHost = flags.String("swarm-host")
	d.SwarmDiscovery = flags.String("swarm-discovery")
	d.SSHUser = "docker"
	d.SSHPort = 22

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

func (d *Driver) GetState() (state.State, error) { // TODO
	// VMRUN only tells use if the vm is running or not
	//	if stdout, _, _ := vmrun("list"); strings.Contains(stdout, d.vmxPath()) {
	return state.Running, nil
	//	}
	//	return state.Stopped, nil
}

func (d *Driver) PreCreateCheck() error {
	reVersion := regexp.MustCompile(`^(\d+\.)?$`)
	ver, err := vboxVersionDetect()
	majorVersion := reVersion.FindString(ver)
	if majorVersion != "5" || majorVersion != "" && err != nil {
		return fmt.Errorf("Virtual Box version 4 or lower will cause a kernel panic if xhyve tries to run." +
			"You are running version: " +
			ver +
			" Please upgrade to version 5 at https://www.virtualbox.org/wiki/Downloads")
	}
	return nil
}

func (d *Driver) Create() error {

	b2dutils := utils.NewB2dUtils("", "")
	if err := b2dutils.CopyIsoToMachineDir(d.Boot2DockerURL, d.MachineName); err != nil {
		return err
	}

	log.Infof("Creating SSH key...")
	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return err
	}

	log.Infof("Creating VM...")
	if err := os.MkdirAll(d.storePath, 0755); err != nil {
		return err
	}

	log.Debugf("Extracting vmlinuz64 and initrd.img from %s...", isoFilename)
	if err := d.extractKernelImages(); err != nil {
		return err
	}

	log.Debugf("Writing boot2docker VM disk image...")
	if err := d.writeDiskImage(); err != nil {
		return err
	}

	log.Debugf("Creating Blank disk image...")
	if err := d.generateBlankDiskImage(d.DiskSize); err != nil { // TODO
		return err
	}

	log.Debugf("Generate UUID...")
	d.UUID = uuidgen()
	log.Debugf(d.UUID) // TODO

	log.Debugf("Create UUID file...")
	if err := d.createUUIDFile(); err != nil {
		return err
	}

	log.Debugf("Running xhyve VM...")
	if err := d.Start(); err != nil {
		return err
	}

	// TODO Maybe get MAC address here from host asignment

	var ip string
	var err error

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

	return nil
}

func (d *Driver) Start() error {
	var Password string
	log.Infof("Creating %s xhyve VM...", d.MachineName)
	cmd := exec.Command("sudo", "xhyve", // TODO
		fmt.Sprintf("-m %dM", d.Memory),
		"-s 0:0,hostbridge -s 31,lpc",
		"-l com1,stdio",
		"-s 2:0,virtio-net",
		fmt.Sprintf("-s 2:1,virtio-tap,tap1"),
		fmt.Sprintf("-s 3,ahci-cd,%d", path.Join(d.storePath, isoFilename)),
		fmt.Sprintf("-s 4,virtio-blk,%d", path.Join(d.storePath, d.MachineName+".img")),
		fmt.Sprintf("-U %s", d.UUID),
		fmt.Sprintf("-f kexec,%s,%s,loglevel=3 user=docker console=ttyS0 console=tty0 noembed nomodeset norestore waitusb=10:LABEL=boot2docker-data base", path.Join(d.storePath, "vmlinuz64"), path.Join(d.storePath, "initrd.img")),
	)
	//	cmd := exec.Command("sudo xhyve -m 4G -c 4 -s 0:0,hostbridge -s 31,lpc -l com1,stdio -s 2:0,virtio-net -s 3,ahci-cd,'/Users/zchee/.docker/machine/machines/xhyve-test/boot2docker.iso' -s 4,virtio-blk,'/Users/zchee/.docker/machine/machines/xhyve-test/xhyve-test.img' -U D2B9B60C-2465-4AF7-BCB6-522D795B043E -f 'kexec,vmlinuz64,initrd.img,loglevel=3 user=docker console=ttyS0 console=tty0 noembed nomodeset norestore waitusb=10:LABEL=boot2docker-data base'")
	cmd.Stdin = strings.NewReader(Password)
	log.Debug(cmd)
	err := cmd.Run()
	if err != nil {
		log.Error(err, cmd.Stdout)
	}

	return nil
}

func (d *Driver) Stop() error { // TODO
	xhyve("controlvm", d.MachineName, "acpipowerbutton")
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

	d.IPAddress = ""

	return nil
}

func (d *Driver) Remove() error { // TODO
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
	//return xhyve("unregistervm", "--delete", d.MachineName)
	return nil
}

func (d *Driver) Restart() error { // TODO
	s, err := d.GetState()
	if err != nil {
		return err
	}

	if s == state.Running {
		if err := d.Stop(); err != nil {
			return err
		}
	}
	return d.Start()
}

func (d *Driver) Kill() error { // TODO
	//return xhyve("controlvm", d.MachineName, "poweroff")
	return nil
}

func (d *Driver) setMachineNameIfNotSet() {
	if d.MachineName == "" {
		d.MachineName = fmt.Sprintf("docker-machine-unknown")
	}
}

func (d *Driver) imgPath() string {
	return path.Join(d.storePath, fmt.Sprintf("%s.img", d.MachineName))
}

func (d *Driver) uuidPath() string {
	return path.Join(d.storePath, "uuid")
}

func (d *Driver) createUUIDFile() error {
	var uuidfile *os.File
	var err error

	if uuidfile, err = os.Create(d.uuidPath()); err != nil {
		return err
	}

	uuid, err := io.WriteString(uuidfile, d.UUID)
	if err != nil {
		log.Debug(uuid, err) // TODO
	}

	uuidfile.Close()
	return nil
}

func (d *Driver) getIPfromDHCPLease() (string, error) {
	var dhcpfh *os.File
	var dhcpcontent []byte
	var macaddr string
	var err error
	var lastipmatch string
	var currentip string

	// DHCP lease table for NAT vmnet interface
	var dhcpfile = "/var/db/dhcpd_leases"

	if dhcpfh, err = os.Open(dhcpfile); err != nil {
		return "", err
	}
	defer dhcpfh.Close()

	if dhcpcontent, err = ioutil.ReadAll(dhcpfh); err != nil {
		return "", err
	}

	// Get the IP from the lease table.
	leaseip := regexp.MustCompile(`^\s*ip_address=(.+?)$`)
	log.Debug(leaseip) // TODO
	// Get the MAC address associated.
	leasemac := regexp.MustCompile(`^\s*hw_address=1,(.+?)$`)
	log.Debug(leasemac) // TODO

	for _, line := range strings.Split(string(dhcpcontent), "\n") {

		if matches := leaseip.FindStringSubmatch(line); matches != nil {
			lastipmatch = matches[1]
			log.Debug(lastipmatch)
			continue
		}

		if matches := leasemac.FindStringSubmatch(line); matches != nil {
			currentip = lastipmatch
			macaddr = matches[1]
			log.Debug(macaddr)
			continue
		}
	}

	if currentip == "" {
		return "", fmt.Errorf("IP not found for MAC %s in DHCP leases", leasemac)
	}

	if macaddr == "" {
		return "", fmt.Errorf("couldn't find MAC address in DHCP leases file %s", dhcpfile)
	}

	log.Debugf("IP found in DHCP lease table: %s", currentip)
	return currentip, nil
}

func (d *Driver) publicSSHKeyPath() string {
	return d.GetSSHKeyPath() + ".pub"
}

func (d *Driver) extractKernelImages() error {
	var vmlinuz64 = "/Volumes/Boot2Docker-v1.7/boot/vmlinuz64"
	var initrd = "/Volumes/Boot2Docker-v1.7/boot/initrd.img"

	hdiutil("attach", d.ISO)
	log.Debugf("Mounting %s", isoFilename)

	log.Debugf("Extract vmlinuz64")
	if err := utils.CopyFile(vmlinuz64, filepath.Join(d.storePath, "vmlinuz64")); err != nil {
		return err
	}
	log.Debugf("Extract initrd.img")
	if err := utils.CopyFile(initrd, filepath.Join(d.storePath, "initrd.img")); err != nil {
		return err
	}

	return nil
}

func (d *Driver) generateBlankDiskImage(count int) error {
	cmd := dd
	output := d.ISO
	cmd("/dev/zero", output, "1m", count)

	return nil
}

// Make a boot2docker VM disk image.
func (d *Driver) writeDiskImage() error { // TODO
	log.Debugf("Creating hard disk image...")

	magicString := "boot2docker, this is xhyve speaking"

	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	// magicString first so the automount script knows to format the disk
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
	raw := buf.Bytes()

	if err := ioutil.WriteFile(d.imgPath(), raw, 0644); err != nil {
		return err
	}

	return nil
}
