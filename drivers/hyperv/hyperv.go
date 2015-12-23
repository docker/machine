package hyperv

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"time"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
)

type Driver struct {
	*drivers.BaseDriver
	Boot2DockerURL string
	VSwitch        string
	diskImage      string
	DiskSize       int
	MemSize        int
	CPU            int
}

const (
	defaultDiskSize = 20000
	defaultMemory   = 1024
	defaultCPU      = 1
)

func NewDriver(hostName, storePath string) *Driver {
	return &Driver{
		DiskSize: defaultDiskSize,
		MemSize:  defaultMemory,
		CPU:      defaultCPU,
		BaseDriver: &drivers.BaseDriver{
			MachineName: hostName,
			StorePath:   storePath,
		},
	}
}

// GetCreateFlags registers the flags this driver adds to
// "docker hosts create"
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			Name:  "hyperv-boot2docker-url",
			Usage: "URL of the boot2docker ISO. Defaults to the latest available version.",
		},
		mcnflag.StringFlag{
			Name:  "hyperv-virtual-switch",
			Usage: "Virtual switch name. Defaults to first found.",
		},
		mcnflag.IntFlag{
			Name:  "hyperv-disk-size",
			Usage: "Maximum size of dynamically expanding disk in MB.",
			Value: defaultDiskSize,
		},
		mcnflag.IntFlag{
			Name:  "hyperv-memory",
			Usage: "Memory size for host in MB.",
			Value: defaultMemory,
		},
		mcnflag.IntFlag{
			Name:  "hyperv-cpu-count",
			Usage: "number of CPUs for the machine",
			Value: defaultCPU,
		},
	}
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.Boot2DockerURL = flags.String("hyperv-boot2docker-url")
	d.VSwitch = flags.String("hyperv-virtual-switch")
	d.DiskSize = flags.Int("hyperv-disk-size")
	d.MemSize = flags.Int("hyperv-memory")
	d.CPU = flags.Int("hyperv-cpu-count")
	d.SwarmMaster = flags.Bool("swarm-master")
	d.SwarmHost = flags.String("swarm-host")
	d.SwarmDiscovery = flags.String("swarm-discovery")
	d.SSHUser = "docker"
	return nil
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "hyperv"
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}

	if ip == "" {
		return "", nil
	}

	return fmt.Sprintf("tcp://%s", net.JoinHostPort(ip, "2376")), nil
}

func (d *Driver) GetState() (state.State, error) {
	stdout, err := cmdOut("(", "Get-VM", "-Name", d.MachineName, ").state")
	if err != nil {
		return state.None, fmt.Errorf("Failed to find the VM status")
	}

	resp := parseLines(stdout)
	if len(resp) < 1 {
		return state.None, nil
	}

	switch resp[0] {
	case "Running":
		return state.Running, nil
	case "Off":
		return state.Stopped, nil
	default:
		return state.None, nil
	}
}

// PreCreateCheck checks that the machine creation process can be started safely.
func (d *Driver) PreCreateCheck() error {
	// Check that hyperv is installed
	if err := hypervAvailable(); err != nil {
		return err
	}

	// Check that the user is an Administrator
	isAdmin, err := isAdministrator()
	if err != nil {
		return err
	}
	if !isAdmin {
		return ErrNotAdministrator
	}

	// Check that there is a virtual switch already configured
	if _, err := d.chooseVirtualSwitch(); err != nil {
		return err
	}

	// Downloading boot2docker to cache should be done here to make sure
	// that a download failure will not leave a machine half created.
	b2dutils := mcnutils.NewB2dUtils(d.StorePath)
	if err := b2dutils.UpdateISOCache(d.Boot2DockerURL); err != nil {
		return err
	}

	return nil
}

func (d *Driver) Create() error {
	b2dutils := mcnutils.NewB2dUtils(d.StorePath)
	if err := b2dutils.CopyIsoToMachineDir(d.Boot2DockerURL, d.MachineName); err != nil {
		return err
	}

	log.Infof("Creating SSH key...")
	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return err
	}

	log.Infof("Creating VM...")
	virtualSwitch, err := d.chooseVirtualSwitch()
	if err != nil {
		return err
	}

	log.Infof("Using switch %q", virtualSwitch)

	err = d.generateDiskImage()
	if err != nil {
		return err
	}

	if err := cmd("New-VM", "-Name", d.MachineName, "-Path", fmt.Sprintf("'%s'", d.ResolveStorePath(".")), "-MemoryStartupBytes", fmt.Sprintf("%dMB", d.MemSize)); err != nil {
		return err
	}

	if d.CPU > 1 {
		if err := cmd("SET-VMProcessor", "-Name", d.MachineName, "-Count", fmt.Sprintf("%d", d.CPU)); err != nil {
			return err
		}
	}

	if err := cmd("Set-VMDvdDrive", "-VMName", d.MachineName, "-Path", fmt.Sprintf("'%s'", d.ResolveStorePath("boot2docker.iso"))); err != nil {
		return err
	}

	if err := cmd("Add-VMHardDiskDrive", "-VMName", d.MachineName, "-Path", fmt.Sprintf("'%s'", d.diskImage)); err != nil {
		return err
	}

	if err := cmd("Connect-VMNetworkAdapter", "-VMName", d.MachineName, "-SwitchName", fmt.Sprintf("'%s'", virtualSwitch)); err != nil {
		return err
	}

	log.Infof("Starting VM...")
	return d.Start()
}

func (d *Driver) chooseVirtualSwitch() (string, error) {
	stdout, err := cmdOut("@(Get-VMSwitch).Name")
	if err != nil {
		return "", err
	}

	switches := parseLines(stdout)

	if d.VSwitch == "" {
		if len(switches) < 1 {
			return "", fmt.Errorf("no vswitch found")
		}

		return switches[0], nil
	}

	found := false
	for _, name := range switches {
		if name == d.VSwitch {
			found = true
			break
		}
	}

	if !found {
		return "", fmt.Errorf("vswitch %q not found", d.VSwitch)
	}

	return d.VSwitch, nil
}

func (d *Driver) wait() error {
	log.Infof("Waiting for host to start...")

	for {
		ip, _ := d.GetIP()
		if ip != "" {
			break
		}

		time.Sleep(1 * time.Second)
	}

	return nil
}

func (d *Driver) Start() error {
	if err := cmd("Start-VM", "-Name", d.MachineName); err != nil {
		return err
	}

	if err := d.wait(); err != nil {
		return err
	}

	var err error
	d.IPAddress, err = d.GetIP()

	return err
}

func (d *Driver) Stop() error {
	if err := cmd("Stop-VM", "-Name", d.MachineName); err != nil {
		return err
	}

	for {
		s, err := d.GetState()
		if err != nil {
			return err
		}

		if s != state.Running {
			break
		}

		time.Sleep(1 * time.Second)
	}

	d.IPAddress = ""

	return nil
}

func (d *Driver) Remove() error {
	s, err := d.GetState()
	if err != nil {
		return err
	}

	if s == state.Running {
		if err := d.Kill(); err != nil {
			return err
		}
	}

	return cmd("Remove-VM", "-Name", d.MachineName, "-Force")
}

func (d *Driver) Restart() error {
	err := d.Stop()
	if err != nil {
		return err
	}

	return d.Start()
}

func (d *Driver) Kill() error {
	if err := cmd("Stop-VM", "-Name", d.MachineName, "-TurnOff"); err != nil {
		return err
	}

	for {
		s, err := d.GetState()
		if err != nil {
			return err
		}

		if s != state.Running {
			break
		}

		time.Sleep(1 * time.Second)
	}

	d.IPAddress = ""

	return nil
}

func (d *Driver) GetIP() (string, error) {
	stdout, err := cmdOut("((", "Get-VM", "-Name", d.MachineName, ").networkadapters[0]).ipaddresses[0]")
	if err != nil {
		return "", err
	}

	resp := parseLines(stdout)
	if len(resp) < 1 {
		return "", fmt.Errorf("IP not found")
	}

	return resp[0], nil
}

func (d *Driver) publicSSHKeyPath() string {
	return d.GetSSHKeyPath() + ".pub"
}

// generateDiskImage creates a small fixed vhd, put the tar in, convert to dynamic, then resize
func (d *Driver) generateDiskImage() error {
	d.diskImage = d.ResolveStorePath("disk.vhd")
	fixed := d.ResolveStorePath("fixed.vhd")

	log.Infof("Creating VHD")
	if err := cmd("New-VHD", "-Path", fmt.Sprintf("'%s'", fixed), "-SizeBytes", "10MB", "-Fixed"); err != nil {
		return err
	}

	tarBuf, err := d.generateTar()
	if err != nil {
		return err
	}

	file, err := os.OpenFile(fixed, os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	file.Seek(0, os.SEEK_SET)
	_, err = file.Write(tarBuf.Bytes())
	if err != nil {
		return err
	}
	file.Close()

	if err := cmd("Convert-VHD", "-Path", fmt.Sprintf("'%s'", fixed), "-DestinationPath", fmt.Sprintf("'%s'", d.diskImage), "-VHDType", "Dynamic"); err != nil {
		return err
	}

	return cmd("Resize-VHD", "-Path", fmt.Sprintf("'%s'", d.diskImage), "-SizeBytes", fmt.Sprintf("%dMB", d.DiskSize))
}

// Make a boot2docker VM disk image.
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
