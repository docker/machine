package virtualbox

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
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
	isoFilename         = "boot2docker-virtualbox.iso"
	defaultHostOnlyCIDR = "192.168.99.1/24"
)

var (
	ErrUnableToGenerateRandomIP = errors.New("unable to generate random IP")
)

type Driver struct {
	*drivers.BaseDriver
	CPU                 int
	Memory              int
	DiskSize            int
	Boot2DockerURL      string
	Boot2DockerImportVM string
	HostOnlyCIDR        string
}

func init() {
	drivers.Register("virtualbox", &drivers.RegisteredDriver{
		New:            NewDriver,
		GetCreateFlags: GetCreateFlags,
	})
}

// RegisterCreateFlags registers the flags this driver adds to
// "docker hosts create"
func GetCreateFlags() []cli.Flag {
	return []cli.Flag{
		cli.IntFlag{
			EnvVar: "VIRTUALBOX_MEMORY_SIZE",
			Name:   "virtualbox-memory",
			Usage:  "Size of memory for host in MB",
			Value:  1024,
		},
		cli.IntFlag{
			EnvVar: "VIRTUALBOX_CPU_COUNT",
			Name:   "virtualbox-cpu-count",
			Usage:  "number of CPUs for the machine (-1 to use the number of CPUs available)",
			Value:  1,
		},
		cli.IntFlag{
			EnvVar: "VIRTUALBOX_DISK_SIZE",
			Name:   "virtualbox-disk-size",
			Usage:  "Size of disk for host in MB",
			Value:  20000,
		},
		cli.StringFlag{
			EnvVar: "VIRTUALBOX_BOOT2DOCKER_URL",
			Name:   "virtualbox-boot2docker-url",
			Usage:  "The URL of the boot2docker image. Defaults to the latest available version",
			Value:  "",
		},
		cli.StringFlag{
			Name:  "virtualbox-import-boot2docker-vm",
			Usage: "The name of a Boot2Docker VM to import",
			Value: "",
		},
		cli.StringFlag{
			Name:   "virtualbox-hostonly-cidr",
			Usage:  "Specify the Host Only CIDR",
			Value:  defaultHostOnlyCIDR,
			EnvVar: "VIRTUALBOX_HOSTONLY_CIDR",
		},
	}
}

func NewDriver(machineName string, storePath string, caCert string, privateKey string) (drivers.Driver, error) {
	inner := drivers.NewBaseDriver(machineName, storePath, caCert, privateKey)
	return &Driver{BaseDriver: inner}, nil
}

func (d *Driver) GetSSHHostname() (string, error) {
	return "localhost", nil
}

func (d *Driver) GetSSHUsername() string {
	if d.SSHUser == "" {
		d.SSHUser = "docker"
	}

	return d.SSHUser
}

func (d *Driver) DriverName() string {
	return "virtualbox"
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
	d.CPU = flags.Int("virtualbox-cpu-count")
	d.Memory = flags.Int("virtualbox-memory")
	d.DiskSize = flags.Int("virtualbox-disk-size")
	d.Boot2DockerURL = flags.String("virtualbox-boot2docker-url")
	d.SwarmMaster = flags.Bool("swarm-master")
	d.SwarmHost = flags.String("swarm-host")
	d.SwarmDiscovery = flags.String("swarm-discovery")
	d.SSHUser = "docker"
	d.Boot2DockerImportVM = flags.String("virtualbox-import-boot2docker-vm")
	d.HostOnlyCIDR = flags.String("virtualbox-hostonly-cidr")

	return nil
}

func (d *Driver) PreCreateCheck() error {
	return nil
}

func (d *Driver) Create() error {
	var (
		err error
	)

	// Check that VBoxManage exists and works
	if err = vbm(); err != nil {
		return err
	}

	b2dutils := utils.NewB2dUtils("", "", isoFilename)
	if err := b2dutils.CopyIsoToMachineDir(d.Boot2DockerURL, d.MachineName); err != nil {
		return err
	}

	log.Infof("Creating VirtualBox VM...")

	// import b2d VM if requested
	if d.Boot2DockerImportVM != "" {
		name := d.Boot2DockerImportVM

		// make sure vm is stopped
		_ = vbm("controlvm", name, "poweroff")

		diskInfo, err := getVMDiskInfo(name)
		if err != nil {
			return err
		}

		if _, err := os.Stat(diskInfo.Path); err != nil {
			return err
		}

		if err := vbm("clonehd", diskInfo.Path, d.diskPath()); err != nil {
			return err
		}

		log.Debugf("Importing VM settings...")
		vmInfo, err := getVMInfo(name)
		if err != nil {
			return err
		}

		d.CPU = vmInfo.CPUs
		d.Memory = vmInfo.Memory

		log.Debugf("Importing SSH key...")
		keyPath := filepath.Join(utils.GetHomeDir(), ".ssh", "id_boot2docker")
		if err := utils.CopyFile(keyPath, d.GetSSHKeyPath()); err != nil {
			return err
		}
	} else {
		log.Infof("Creating SSH key...")
		if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
			return err
		}

		log.Debugf("Creating disk image...")
		if err := vbm("createhd", "--size", fmt.Sprintf("%d", d.DiskSize), "--format", "VMDK", "--filename", d.diskPath()); err != nil {
			return err
		}

	}

	if err := vbm("createvm",
		"--basefolder", d.ResolveStorePath("."),
		"--name", d.MachineName,
		"--register"); err != nil {
		return err
	}

	log.Debugf("VM CPUS: %d", d.CPU)
	log.Debugf("VM Memory: %d", d.Memory)

	cpus := d.CPU
	if cpus < 1 {
		cpus = int(runtime.NumCPU())
	}
	if cpus > 32 {
		cpus = 32
	}

	if err := vbm("modifyvm", d.MachineName,
		"--firmware", "bios",
		"--bioslogofadein", "off",
		"--bioslogofadeout", "off",
		"--bioslogodisplaytime", "0",
		"--biosbootmenu", "disabled",
		"--ostype", "Linux26_64",
		"--cpus", fmt.Sprintf("%d", cpus),
		"--memory", fmt.Sprintf("%d", d.Memory),
		"--acpi", "on",
		"--ioapic", "on",
		"--rtcuseutc", "on",
		"--natdnshostresolver1", "off",
		"--natdnsproxy1", "off",
		"--cpuhotplug", "off",
		"--pae", "on",
		"--hpet", "on",
		"--hwvirtex", "on",
		"--nestedpaging", "on",
		"--largepages", "on",
		"--vtxvpid", "on",
		"--accelerate3d", "off",
		"--boot1", "dvd"); err != nil {
		return err
	}

	if err := vbm("modifyvm", d.MachineName,
		"--nic1", "nat",
		"--nictype1", "82540EM",
		"--cableconnected1", "on"); err != nil {
		return err
	}

	if err := d.setupHostOnlyNetwork(d.MachineName); err != nil {
		return err
	}

	if err := vbm("storagectl", d.MachineName,
		"--name", "SATA",
		"--add", "sata",
		"--hostiocache", "on"); err != nil {
		return err
	}

	if err := vbm("storageattach", d.MachineName,
		"--storagectl", "SATA",
		"--port", "0",
		"--device", "0",
		"--type", "dvddrive",
		"--medium", d.ResolveStorePath(isoFilename)); err != nil {
		return err
	}

	if err := vbm("storageattach", d.MachineName,
		"--storagectl", "SATA",
		"--port", "1",
		"--device", "0",
		"--type", "hdd",
		"--medium", d.diskPath()); err != nil {
		return err
	}

	shareDir := homedir.Get()
	shareName := shareDir

	log.Debugf("creating share: path=%s", shareDir)

	// let VBoxService do nice magic automounting (when it's used)
	if err := vbm("guestproperty", "set", d.MachineName, "/VirtualBox/GuestAdd/SharedFolders/MountDir", "/"); err != nil {
		return err
	}
	if err := vbm("guestproperty", "set", d.MachineName, "/VirtualBox/GuestAdd/SharedFolders/MountPrefix", "/"); err != nil {
		return err
	}

	if shareDir != "" {
		log.Debugf("setting up shareDir")
		if _, err := os.Stat(shareDir); err != nil && !os.IsNotExist(err) {
			log.Debugf("setting up share failed: %s", err)
			return err
		} else if !os.IsNotExist(err) {
			// parts of the VBox internal code are buggy with share names that start with "/"
			shareName = strings.TrimLeft(shareDir, "/")

			// translate to msys git path
			if runtime.GOOS == "windows" {
				mountName, err := translateWindowsMount(shareDir)
				if err != nil {
					return err
				}
				shareName = mountName
			}

			log.Debugf("adding shared folder: name=%q dir=%q", shareName, shareDir)

			// woo, shareDir exists!  let's carry on!
			if err := vbm("sharedfolder", "add", d.MachineName, "--name", shareName, "--hostpath", shareDir, "--automount"); err != nil {
				return err
			}

			// enable symlinks
			if err := vbm("setextradata", d.MachineName, "VBoxInternal2/SharedFoldersEnableSymlinksCreate/"+shareName, "1"); err != nil {
				return err
			}
		}
	}

	log.Infof("Starting VirtualBox VM...")

	if err := d.Start(); err != nil {
		return err
	}

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

	ip, err := d.GetIP()
	if err != nil {
		return err
	}
	d.IPAddress = ip
	return nil
}

func (d *Driver) hostOnlyIpAvailable() bool {
	ip, err := d.GetIP()
	if err != nil {
		log.Debugf("ERROR getting IP: %s", err)
		return false
	}
	if ip != "" {
		log.Debugf("IP is %s", ip)
		return true
	}
	log.Debug("Strangely, there was no error attempting to get the IP, but it was still empty.")
	return false
}

func (d *Driver) Start() error {
	s, err := d.GetState()
	if err != nil {
		return err
	}

	if s == state.Stopped {
		// check network to re-create if needed
		if err := d.setupHostOnlyNetwork(d.MachineName); err != nil {
			return fmt.Errorf("Error setting up host only network on machine start: %s", err)
		}
	}

	switch s {
	case state.Stopped, state.Saved:
		d.SSHPort, err = setPortForwarding(d.MachineName, 1, "ssh", "tcp", 22, d.SSHPort)
		if err != nil {
			return err
		}
		if err := vbm("startvm", d.MachineName, "--type", "headless"); err != nil {
			return err
		}
		log.Infof("Starting VM...")
	case state.Paused:
		if err := vbm("controlvm", d.MachineName, "resume", "--type", "headless"); err != nil {
			return err
		}
		log.Infof("Resuming VM ...")
	default:
		log.Infof("VM not in restartable state")
	}

	addr, err := d.GetSSHHostname()
	if err := ssh.WaitForTCP(fmt.Sprintf("%s:%d", addr, d.SSHPort)); err != nil {
		return err
	}

	// Bail if we don't get an IP from DHCP after a given number of seconds.
	if err := utils.WaitForSpecific(d.hostOnlyIpAvailable, 5, 4*time.Second); err != nil {
		return err
	}

	return err
}

func (d *Driver) Stop() error {
	if err := vbm("controlvm", d.MachineName, "acpipowerbutton"); err != nil {
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

	d.IPAddress = ""

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
	// vbox will not release it's lock immediately after the stop
	time.Sleep(1 * time.Second)
	return vbm("unregistervm", "--delete", d.MachineName)
}

func (d *Driver) Restart() error {
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

func (d *Driver) Kill() error {
	return vbm("controlvm", d.MachineName, "poweroff")
}

func (d *Driver) GetState() (state.State, error) {
	stdout, stderr, err := vbmOutErr("showvminfo", d.MachineName,
		"--machinereadable")
	if err != nil {
		if reMachineNotFound.FindString(stderr) != "" {
			return state.Error, ErrMachineNotExist
		}
		return state.Error, err
	}
	re := regexp.MustCompile(`(?m)^VMState="(\w+)"`)
	groups := re.FindStringSubmatch(stdout)
	if len(groups) < 1 {
		return state.None, nil
	}
	switch groups[1] {
	case "running":
		return state.Running, nil
	case "paused":
		return state.Paused, nil
	case "saved":
		return state.Saved, nil
	case "poweroff", "aborted":
		return state.Stopped, nil
	}
	return state.None, nil
}

func (d *Driver) setMachineNameIfNotSet() {
	if d.MachineName == "" {
		d.MachineName = fmt.Sprintf("docker-machine-unknown")
	}
}

func (d *Driver) GetIP() (string, error) {
	// DHCP is used to get the IP, so virtualbox hosts don't have IPs unless
	// they are running
	s, err := d.GetState()
	if err != nil {
		return "", err
	}
	if s != state.Running {
		return "", drivers.ErrHostIsNotRunning
	}

	sshClient, err := d.getLocalSSHClient()
	if err != nil {
		return "", err
	}

	output, err := sshClient.Output("ip addr show dev eth1")
	if err != nil {
		log.Debug(output)
		return "", err
	}

	log.Debugf("SSH returned: %s\nEND SSH\n", output)

	// parse to find: inet 192.168.59.103/24 brd 192.168.59.255 scope global eth1
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		vals := strings.Split(strings.TrimSpace(line), " ")
		if len(vals) >= 2 && vals[0] == "inet" {
			return vals[1][:strings.Index(vals[1], "/")], nil
		}
	}

	return "", fmt.Errorf("No IP address found %s", output)
}

func (d *Driver) publicSSHKeyPath() string {
	return d.GetSSHKeyPath() + ".pub"
}

func (d *Driver) diskPath() string {
	return d.ResolveStorePath("disk.vmdk")
}

func (d *Driver) setupHostOnlyNetwork(machineName string) error {
	hostOnlyCIDR := d.HostOnlyCIDR

	// This is to assist in migrating from version 0.2 to 0.3 format
	// it should be removed in a later release
	if hostOnlyCIDR == "" {
		hostOnlyCIDR = defaultHostOnlyCIDR
	}

	ip, network, err := net.ParseCIDR(hostOnlyCIDR)

	if err != nil {
		return err
	}

	nAddr := network.IP.To4()

	dhcpAddr, err := getRandomIPinSubnet(network.IP)
	if err != nil {
		return err
	}

	lowerDHCPIP := net.IPv4(nAddr[0], nAddr[1], nAddr[2], byte(100))
	upperDHCPIP := net.IPv4(nAddr[0], nAddr[1], nAddr[2], byte(254))

	log.Debugf("using %s for dhcp address", dhcpAddr)

	hostOnlyNetwork, err := getOrCreateHostOnlyNetwork(
		ip,
		network.Mask,
		dhcpAddr,
		lowerDHCPIP,
		upperDHCPIP,
	)

	if err != nil {
		return err
	}

	if err := vbm("modifyvm", machineName,
		"--nic2", "hostonly",
		"--nictype2", "82540EM",
		"--hostonlyadapter2", hostOnlyNetwork.Name,
		"--cableconnected2", "on"); err != nil {
		return err
	}

	return nil
}

// Select an available port, trying the specified
// port first, falling back on an OS selected port.
func getAvailableTCPPort(port int) (int, error) {
	for i := 0; i <= 10; i++ {
		ln, err := net.Listen("tcp4", fmt.Sprintf("127.0.0.1:%d", port))
		if err != nil {
			return 0, err
		}
		defer ln.Close()
		addr := ln.Addr().String()
		addrParts := strings.SplitN(addr, ":", 2)
		p, err := strconv.Atoi(addrParts[1])
		if err != nil {
			return 0, err
		}
		if p != 0 {
			port = p
			return port, nil
		}
		port = 0 // Throw away the port hint before trying again
		time.Sleep(1)
	}
	return 0, fmt.Errorf("unable to allocate tcp port")
}

// Setup a NAT port forwarding entry.
func setPortForwarding(machine string, interfaceNum int, mapName, protocol string, guestPort, desiredHostPort int) (int, error) {
	actualHostPort, err := getAvailableTCPPort(desiredHostPort)
	if err != nil {
		return -1, err
	}
	if desiredHostPort != actualHostPort && desiredHostPort != 0 {
		log.Debugf("NAT forwarding host port for guest port %d (%s) changed from %d to %d",
			guestPort, mapName, desiredHostPort, actualHostPort)
	}
	cmd := fmt.Sprintf("--natpf%d", interfaceNum)
	vbm("modifyvm", machine, cmd, "delete", mapName)
	if err := vbm("modifyvm", machine,
		cmd, fmt.Sprintf("%s,%s,127.0.0.1,%d,,%d", mapName, protocol, actualHostPort, guestPort)); err != nil {
		return -1, err
	}
	return actualHostPort, nil
}

// getRandomIPinSubnet returns a pseudo-random net.IP in the same
// subnet as the IP passed
func getRandomIPinSubnet(baseIP net.IP) (net.IP, error) {
	var dhcpAddr net.IP

	nAddr := baseIP.To4()
	// select pseudo-random DHCP addr; make sure not to clash with the host
	// only try 5 times and bail if no random received
	for i := 0; i < 5; i++ {
		n := rand.Intn(25)
		if byte(n) != nAddr[3] {
			dhcpAddr = net.IPv4(nAddr[0], nAddr[1], nAddr[2], byte(1))
			break
		}
	}

	if dhcpAddr == nil {
		return nil, ErrUnableToGenerateRandomIP
	}

	return dhcpAddr, nil
}

func (d *Driver) getLocalSSHClient() (ssh.Client, error) {
	sshAuth := &ssh.Auth{
		Passwords: []string{"docker"},
		Keys:      []string{d.GetSSHKeyPath()},
	}
	sshClient, err := ssh.NewNativeClient(d.GetSSHUsername(), "127.0.0.1", d.SSHPort, sshAuth)
	if err != nil {
		return nil, err
	}

	return sshClient, nil
}

func translateWindowsMount(p string) (string, error) {
	re := regexp.MustCompile(`(?P<drive>[^:]+):\\(?P<path>.*)`)
	m := re.FindStringSubmatch(p)

	var drive, fullPath string

	if len(m) < 3 {
		return "", fmt.Errorf("unable to parse home directory")
	}

	drive = m[1]
	fullPath = m[2]

	nPath := strings.Replace(fullPath, "\\", "/", -1)

	tPath := path.Join("/", strings.ToLower(drive), nPath)
	return tPath, nil
}
