package hyperv

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
)

type Driver struct {
	*drivers.BaseDriver
	boot2DockerURL string
	boot2DockerLoc string
	vSwitch        string
	diskImage      string
	DiskSize       int
	MemSize        int
}

const (
	defaultDiskSize = 20000
	defaultMemory   = 1024
)

func init() {
	drivers.Register("hyper-v", &drivers.RegisteredDriver{
		GetCreateFlags: GetCreateFlags,
	})
}

func NewDriver(hostName, storePath string) drivers.Driver {
	return &Driver{
		DiskSize: defaultDiskSize,
		MemSize:  defaultMemory,
		BaseDriver: &drivers.BaseDriver{
			MachineName: hostName,
			StorePath:   storePath,
		},
	}
}

// GetCreateFlags registers the flags this driver adds to
// "docker hosts create"
func GetCreateFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "hyper-v-boot2docker-url",
			Usage: "Hyper-V URL of the boot2docker image. Defaults to the latest available version.",
		},
		cli.StringFlag{
			Name:  "hyper-v-boot2docker-location",
			Usage: "Hyper-V local boot2docker iso. Overrides URL.",
		},
		cli.StringFlag{
			Name:  "hyper-v-virtual-switch",
			Usage: "Hyper-V virtual switch name. Defaults to first found.",
		},
		cli.IntFlag{
			Name:  "hyper-v-disk-size",
			Usage: "Hyper-V disk size for host in MB.",
			Value: defaultDiskSize,
		},
		cli.IntFlag{
			Name:  "hyper-v-memory",
			Usage: "Hyper-V memory size for host in MB.",
			Value: defaultMemory,
		},
	}
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.boot2DockerURL = flags.String("hyper-v-boot2docker-url")
	d.boot2DockerLoc = flags.String("hyper-v-boot2docker-location")
	d.vSwitch = flags.String("hyper-v-virtual-switch")
	d.DiskSize = flags.Int("hyper-v-disk-size")
	d.MemSize = flags.Int("hyper-v-memory")
	d.SwarmMaster = flags.Bool("swarm-master")
	d.SwarmHost = flags.String("swarm-host")
	d.SwarmDiscovery = flags.String("swarm-discovery")
	d.SSHUser = "docker"
	d.SSHPort = 22
	return nil
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
	return "hyper-v"
}

func (d *Driver) PreCreateCheck() error {
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

func (d *Driver) GetState() (state.State, error) {

	command := []string{
		"(",
		"Get-VM",
		"-Name", d.MachineName,
		").state"}
	stdout, err := execute(command)
	if err != nil {
		return state.None, fmt.Errorf("Failed to find the VM status")
	}
	resp := parseStdout(stdout)

	if len(resp) < 1 {
		return state.None, nil
	}
	switch resp[0] {
	case "Running":
		return state.Running, nil
	case "Off":
		return state.Stopped, nil
	}
	return state.None, nil
}

func (d *Driver) Create() error {
	err := hypervAvailable()
	if err != nil {
		return err
	}

	d.setMachineNameIfNotSet()

	b2dutils := mcnutils.NewB2dUtils("", "", d.StorePath)
	if err := b2dutils.CopyIsoToMachineDir(d.boot2DockerURL, d.MachineName); err != nil {
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

	err = d.generateDiskImage()
	if err != nil {
		return err
	}

	command := []string{
		"New-VM",
		"-Name", d.MachineName,
		"-Path", fmt.Sprintf("'%s'", d.ResolveStorePath(".")),
		"-MemoryStartupBytes", fmt.Sprintf("%dMB", d.MemSize)}
	_, err = execute(command)
	if err != nil {
		return err
	}

	command = []string{
		"Set-VMDvdDrive",
		"-VMName", d.MachineName,
		"-Path", fmt.Sprintf("'%s'", d.ResolveStorePath("boot2docker.iso"))}
	_, err = execute(command)
	if err != nil {
		return err
	}

	command = []string{
		"Add-VMHardDiskDrive",
		"-VMName", d.MachineName,
		"-Path", fmt.Sprintf("'%s'", d.diskImage)}
	_, err = execute(command)
	if err != nil {
		return err
	}

	command = []string{
		"Connect-VMNetworkAdapter",
		"-VMName", d.MachineName,
		"-SwitchName", fmt.Sprintf("'%s'", virtualSwitch)}
	_, err = execute(command)
	if err != nil {
		return err
	}

	log.Infof("Starting  VM...")
	if err := d.Start(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) chooseVirtualSwitch() (string, error) {
	if d.vSwitch != "" {
		return d.vSwitch, nil
	}
	command := []string{
		"@(Get-VMSwitch).Name"}
	stdout, err := execute(command)
	if err != nil {
		return "", err
	}
	switches := parseStdout(stdout)
	if len(switches) > 0 {
		log.Infof("Using switch %s", switches[0])
		return switches[0], nil
	}
	return "", fmt.Errorf("no vswitch found")
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
	command := []string{
		"Start-VM",
		"-Name", d.MachineName}
	_, err := execute(command)
	if err != nil {
		return err
	}

	if err := d.wait(); err != nil {
		return err
	}

	d.IPAddress, err = d.GetIP()
	return err
}

func (d *Driver) Stop() error {
	command := []string{
		"Stop-VM",
		"-Name", d.MachineName}
	_, err := execute(command)
	if err != nil {
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
		return err
	}
	if s == state.Running {
		if err := d.Kill(); err != nil {
			return err
		}
	}
	command := []string{
		"Remove-VM",
		"-Name", d.MachineName,
		"-Force"}
	_, err = execute(command)
	return err
}

func (d *Driver) Restart() error {
	err := d.Stop()
	if err != nil {
		return err
	}

	return d.Start()
}

func (d *Driver) Kill() error {
	command := []string{
		"Stop-VM",
		"-Name", d.MachineName,
		"-TurnOff"}
	_, err := execute(command)
	if err != nil {
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

func (d *Driver) setMachineNameIfNotSet() {
	if d.MachineName == "" {
		d.MachineName = fmt.Sprintf("docker-machine-unknown")
	}
}

func (d *Driver) GetIP() (string, error) {
	command := []string{
		"((",
		"Get-VM",
		"-Name", d.MachineName,
		").networkadapters[0]).ipaddresses[0]"}
	stdout, err := execute(command)
	if err != nil {
		return "", err
	}
	resp := parseStdout(stdout)
	if len(resp) < 1 {
		return "", fmt.Errorf("IP not found")
	}
	return resp[0], nil
}

func (d *Driver) publicSSHKeyPath() string {
	return d.GetSSHKeyPath() + ".pub"
}

func (d *Driver) generateDiskImage() error {
	// Create a small fixed vhd, put the tar in,
	// convert to dynamic, then resize

	d.diskImage = d.ResolveStorePath("disk.vhd")
	fixed := d.ResolveStorePath("fixed.vhd")
	log.Infof("Creating VHD")
	command := []string{
		"New-VHD",
		"-Path", fmt.Sprintf("'%s'", fixed),
		"-SizeBytes", "10MB",
		"-Fixed"}
	_, err := execute(command)
	if err != nil {
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

	command = []string{
		"Convert-VHD",
		"-Path", fmt.Sprintf("'%s'", fixed),
		"-DestinationPath", fmt.Sprintf("'%s'", d.diskImage),
		"-VHDType", "Dynamic"}
	_, err = execute(command)
	if err != nil {
		return err
	}
	command = []string{
		"Resize-VHD",
		"-Path", fmt.Sprintf("'%s'", d.diskImage),
		"-SizeBytes", fmt.Sprintf("%dMB", d.DiskSize)}
	_, err = execute(command)
	if err != nil {
		return err
	}

	return err
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
