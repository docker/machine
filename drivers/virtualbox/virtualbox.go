package virtualbox

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
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
	isoFilename = "boot2docker.iso"
)

type Driver struct {
	IPAddress           string
	CPU                 int
	MachineName         string
	SSHUser             string
	SSHPort             int
	Memory              int
	DiskSize            int
	Boot2DockerURL      string
	CaCertPath          string
	PrivateKeyPath      string
	SwarmMaster         bool
	SwarmHost           string
	SwarmDiscovery      string
	storePath           string
	Boot2DockerImportVM string
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
	return "localhost", nil
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

func (d *Driver) GetProviderType() provider.ProviderType {
	return provider.Local
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

	b2dutils := utils.NewB2dUtils("", "")
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
		if err := d.generateDiskImage(d.DiskSize); err != nil {
			return err
		}
	}

	if err := vbm("createvm",
		"--basefolder", d.storePath,
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
		"--synthcpu", "off",
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
		"--nictype1", "virtio",
		"--cableconnected1", "on"); err != nil {
		return err
	}

	hostOnlyNetwork, err := getOrCreateHostOnlyNetwork(
		net.ParseIP("192.168.99.1"),
		net.IPv4Mask(255, 255, 255, 0),
		net.ParseIP("192.168.99.2"),
		net.ParseIP("192.168.99.100"),
		net.ParseIP("192.168.99.254"))
	if err != nil {
		return err
	}
	if err := vbm("modifyvm", d.MachineName,
		"--nic2", "hostonly",
		"--nictype2", "virtio",
		"--hostonlyadapter2", hostOnlyNetwork.Name,
		"--cableconnected2", "on"); err != nil {
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
		"--medium", filepath.Join(d.storePath, "boot2docker.iso")); err != nil {
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

	// let VBoxService do nice magic automounting (when it's used)
	if err := vbm("guestproperty", "set", d.MachineName, "/VirtualBox/GuestAdd/SharedFolders/MountPrefix", "/"); err != nil {
		return err
	}
	if err := vbm("guestproperty", "set", d.MachineName, "/VirtualBox/GuestAdd/SharedFolders/MountDir", "/"); err != nil {
		return err
	}

	var shareName, shareDir string // TODO configurable at some point
	switch runtime.GOOS {
	case "windows":
		shareName = "c/Users"
		shareDir = "c:\\Users"
	case "darwin":
		shareName = "Users"
		shareDir = "/Users"
		// TODO "linux"
	}

	if shareDir != "" {
		if _, err := os.Stat(shareDir); err != nil && !os.IsNotExist(err) {
			return err
		} else if !os.IsNotExist(err) {
			if shareName == "" {
				// parts of the VBox internal code are buggy with share names that start with "/"
				shareName = strings.TrimLeft(shareDir, "/")
				// TODO do some basic Windows -> MSYS path conversion
				// ie, s!^([a-z]+):[/\\]+!\1/!; s!\\!/!g
			}

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

	return nil
}

func (d *Driver) Start() error {
	s, err := d.GetState()
	if err != nil {
		return err
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

	if err := drivers.WaitForSSH(d); err != nil {
		return err
	}

	d.IPAddress, err = d.GetIP()
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
	output, err := drivers.RunSSHCommandFromDriver(d, "ip addr show dev eth1")
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(output.Stdout); err != nil {
		return "", err
	}

	out := buf.String()
	log.Debugf("SSH returned: %s\nEND SSH\n", out)
	// parse to find: inet 192.168.59.103/24 brd 192.168.59.255 scope global eth1
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		vals := strings.Split(strings.TrimSpace(line), " ")
		if len(vals) >= 2 && vals[0] == "inet" {
			return vals[1][:strings.Index(vals[1], "/")], nil
		}
	}

	return "", fmt.Errorf("No IP address found %s", out)
}

func (d *Driver) publicSSHKeyPath() string {
	return d.GetSSHKeyPath() + ".pub"
}

func (d *Driver) diskPath() string {
	return filepath.Join(d.storePath, "disk.vmdk")
}

// Make a boot2docker VM disk image.
func (d *Driver) generateDiskImage(size int) error {
	log.Debugf("Creating %d MB hard disk image...", size)

	magicString := "boot2docker, please format-me"

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
	raw := bytes.NewReader(buf.Bytes())
	return createDiskImage(d.diskPath(), size, raw)
}

// createDiskImage makes a disk image at dest with the given size in MB. If r is
// not nil, it will be read as a raw disk image to convert from.
func createDiskImage(dest string, size int, r io.Reader) error {
	// Convert a raw image from stdin to the dest VMDK image.
	sizeBytes := int64(size) << 20 // usually won't fit in 32-bit int (max 2GB)
	// FIXME: why isn't this just using the vbm*() functions?
	cmd := exec.Command(vboxManageCmd, "convertfromraw", "stdin", dest,
		fmt.Sprintf("%d", sizeBytes), "--format", "VMDK")

	if os.Getenv("DEBUG") != "" {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	n, err := io.Copy(stdin, r)
	if err != nil {
		return err
	}

	// The total number of bytes written to stdin must match sizeBytes, or
	// VBoxManage.exe on Windows will fail. Fill remaining with zeros.
	if left := sizeBytes - n; left > 0 {
		if err := zeroFill(stdin, left); err != nil {
			return err
		}
	}

	// cmd won't exit until the stdin is closed.
	if err := stdin.Close(); err != nil {
		return err
	}

	return cmd.Wait()
}

// zeroFill writes n zero bytes into w.
func zeroFill(w io.Writer, n int64) error {
	const blocksize = 32 << 10
	zeros := make([]byte, blocksize)
	var k int
	var err error
	for n > 0 {
		if n > blocksize {
			k, err = w.Write(zeros)
		} else {
			k, err = w.Write(zeros[:n])
		}
		if err != nil {
			return err
		}
		n -= int64(k)
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
